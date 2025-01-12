package application

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"go.uber.org/zap"

	"docgent-backend/internal/model/infrastructure"
	"docgent-backend/internal/workflow"
)

type SlackEventHandler struct {
	log                *zap.Logger
	documentationAgent infrastructure.DocumentationAgent
	documentStore      infrastructure.DocumentStore
	slackClient        *slack.Client
	signingSecret      string
}

func NewSlackEventHandler(
	log *zap.Logger,
	agent infrastructure.DocumentationAgent,
	store infrastructure.DocumentStore,
) *SlackEventHandler {
	// 環境変数からSlackの認証情報を取得
	token := os.Getenv("SLACK_BOT_TOKEN")
	signingSecret := os.Getenv("SLACK_SIGNING_SECRET")

	return &SlackEventHandler{
		log:                log,
		documentationAgent: agent,
		documentStore:      store,
		slackClient:        slack.New(token),
		signingSecret:      signingSecret,
	}
}

func (h *SlackEventHandler) Pattern() string {
	return "/api/slack/events"
}

func (h *SlackEventHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.log.Warn("Failed to read request", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	event, err := slackevents.ParseEvent(body, slackevents.OptionVerifyToken(&slackevents.TokenComparator{
		VerificationToken: h.signingSecret,
	}))
	if err != nil {
		h.log.Warn("Failed to parse event", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// URLの検証チャレンジに応答
	if event.Type == slackevents.URLVerification {
		var challenge struct {
			Challenge string `json:"challenge"`
		}
		if err := json.Unmarshal(body, &challenge); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(challenge.Challenge))
		return
	}

	// イベントの処理
	if event.Type == slackevents.CallbackEvent {
		innerEvent := event.InnerEvent
		switch ev := innerEvent.Data.(type) {
		case *slackevents.ReactionAddedEvent:
			// docgent emojiが付与された場合の処理
			if ev.Reaction == "docgent" {
				go h.handleReactionEvent(ev)
			}
		}
	}

	// Slackイベントには即座に200 OKを返す
	w.WriteHeader(http.StatusOK)
}

func (h *SlackEventHandler) handleReactionEvent(ev *slackevents.ReactionAddedEvent) {
	// スレッドの内容を取得
	threadTimestamp := ev.Item.Timestamp

	// スレッドのメッセージを取得
	messages, _, _, err := h.slackClient.GetConversationReplies(&slack.GetConversationRepliesParameters{
		ChannelID: ev.Item.Channel,
		Timestamp: threadTimestamp,
	})
	if err != nil {
		h.postErrorMessage(ev.Item.Channel, threadTimestamp, "スレッドの取得に失敗しました")
		return
	}

	// スレッドの内容を結合
	var text string
	for _, msg := range messages {
		text += msg.Text + "\n"
	}

	// ドキュメントを生成
	ctx := context.Background()
	draftGenerateWorkflow := workflow.NewDraftGenerateWorkflow(workflow.DraftGenerateWorkflowParams{
		DocumentationAgent: h.documentationAgent,
		DocumentStore:      h.documentStore,
	})
	draft, err := draftGenerateWorkflow.Execute(ctx, text)
	if err != nil {
		h.postErrorMessage(ev.Item.Channel, threadTimestamp, "ドキュメントの生成に失敗しました")
		return
	}

	// 成功メッセージを投稿
	h.slackClient.PostMessage(ev.Item.Channel,
		slack.MsgOptionText(fmt.Sprintf("ドキュメントを生成しました！\nタイトル: %s", draft.Title), false),
		slack.MsgOptionTS(threadTimestamp),
	)
}

func (h *SlackEventHandler) postErrorMessage(channel, threadTs, message string) {
	h.slackClient.PostMessage(channel,
		slack.MsgOptionText(fmt.Sprintf(":warning: エラー: %s", message), false),
		slack.MsgOptionTS(threadTs),
	)
}
