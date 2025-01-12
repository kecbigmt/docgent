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
	// Slackからのリクエストを検証
	sv, err := slack.NewSecretsVerifier(r.Header, h.signingSecret)
	if err != nil {
		h.log.Warn("Failed to create secrets verifier", zap.Error(err))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	bodyReader := io.TeeReader(r.Body, &sv)
	body, err := io.ReadAll(bodyReader)
	if err != nil {
		h.log.Warn("Failed to read request", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if err := sv.Ensure(); err != nil {
		h.log.Warn("Failed to verify request", zap.Error(err))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	event, err := slackevents.ParseEvent(body, slackevents.OptionNoVerifyToken())
	if err != nil {
		h.log.Warn("Failed to parse event", zap.Error(err))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// URLの検証チャレンジに応答
	if event.Type == slackevents.URLVerification {
		var res *slackevents.ChallengeResponse
		if err := json.Unmarshal(body, &res); err != nil {
			h.log.Error("Failed to unmarshal challenge", zap.Error(err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"challenge": res.Challenge,
		})
		return
	}

	// イベントの処理
	if event.Type == slackevents.CallbackEvent {
		innerEvent := event.InnerEvent

		eventJSON, err := json.Marshal(innerEvent)
		if err != nil {
			h.log.Warn("Failed to marshal inner event", zap.Error(err))
			h.log.Info("Received event", zap.Any("event", innerEvent))
		} else {
			h.log.Info("Received event", zap.String("event", string(eventJSON)))
		}

		switch ev := innerEvent.Data.(type) {
		case *slackevents.ReactionAddedEvent:
			// docgent emojiが付与された場合の処理
			if ev.Reaction == "doc_it" {
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
	draftGenerateWorkflow := workflow.NewDraftGenerateWorkflow(
		h.documentationAgent,
		h.documentStore,
	)
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
