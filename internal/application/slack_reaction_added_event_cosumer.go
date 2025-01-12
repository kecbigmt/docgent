package application

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"go.uber.org/fx"

	"docgent-backend/internal/model/infrastructure"
	"docgent-backend/internal/workflow"
)

type SlackReactionAddedEventConsumerParams struct {
	fx.In

	SlackAPI           SlackAPI
	DocumentationAgent infrastructure.DocumentationAgent
	DocumentStore      infrastructure.DocumentStore
}

type SlackReactionAddedEventConsumer struct {
	slackAPI           SlackAPI
	documentationAgent infrastructure.DocumentationAgent
	documentStore      infrastructure.DocumentStore
}

func NewSlackReactionAddedEventConsumer(params SlackReactionAddedEventConsumerParams) *SlackReactionAddedEventConsumer {
	return &SlackReactionAddedEventConsumer{
		slackAPI:           params.SlackAPI,
		documentationAgent: params.DocumentationAgent,
		documentStore:      params.DocumentStore,
	}
}

func (h *SlackReactionAddedEventConsumer) EventType() string {
	return "reaction_added"
}

func (h *SlackReactionAddedEventConsumer) ConsumeEvent(event slackevents.EventsAPIInnerEvent) {
	ev, ok := event.Data.(*slackevents.ReactionAddedEvent)
	if !ok {
		return
	}

	threadTimestamp := ev.Item.Timestamp

	slackClient := h.slackAPI.GetClient()

	// スレッドのメッセージを取得
	messages, _, _, err := slackClient.GetConversationReplies(&slack.GetConversationRepliesParameters{
		ChannelID: ev.Item.Channel,
		Timestamp: threadTimestamp,
	})
	if err != nil {
		slackClient.PostMessage(ev.Item.Channel,
			slack.MsgOptionText(":warning: エラー: スレッドの取得に失敗しました", false),
			slack.MsgOptionTS(threadTimestamp),
		)
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
		slackClient.PostMessage(ev.Item.Channel,
			slack.MsgOptionText(":warning: エラー: スレッドの取得に失敗しました", false),
			slack.MsgOptionTS(threadTimestamp),
		)
		return
	}

	// 成功メッセージを投稿
	slackClient.PostMessage(ev.Item.Channel,
		slack.MsgOptionText(fmt.Sprintf("ドキュメントを生成しました！\nタイトル: %s", draft.Title), false),
		slack.MsgOptionTS(threadTimestamp),
	)
}
