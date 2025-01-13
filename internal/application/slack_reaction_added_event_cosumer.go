package application

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"docgent-backend/internal/infrastructure/github"
	"docgent-backend/internal/model/infrastructure"
	"docgent-backend/internal/workflow"
)

type SlackReactionAddedEventConsumerParams struct {
	fx.In

	Logger             *zap.Logger
	SlackAPI           SlackAPI
	GitHubAPI          github.API
	DocumentationAgent infrastructure.DocumentationAgent
}

type SlackReactionAddedEventConsumer struct {
	logger             *zap.Logger
	slackAPI           SlackAPI
	githubAPI          github.API
	documentationAgent infrastructure.DocumentationAgent
}

func NewSlackReactionAddedEventConsumer(params SlackReactionAddedEventConsumerParams) *SlackReactionAddedEventConsumer {
	return &SlackReactionAddedEventConsumer{
		logger:             params.Logger,
		slackAPI:           params.SlackAPI,
		githubAPI:          params.GitHubAPI,
		documentationAgent: params.DocumentationAgent,
	}
}

func (h *SlackReactionAddedEventConsumer) EventType() string {
	return "reaction_added"
}

func (h *SlackReactionAddedEventConsumer) ConsumeEvent(event slackevents.EventsAPIInnerEvent, githubAppParams GitHubAppParams) {
	ev, ok := event.Data.(*slackevents.ReactionAddedEvent)
	if !ok {
		h.logger.Error("Failed to convert event data to ReactionAddedEvent")
		return
	}

	threadTimestamp := ev.Item.Timestamp

	slackClient := h.slackAPI.GetClient()

	githubClient := h.githubAPI.NewClient(githubAppParams.InstallationID)
	documentStore := github.NewDocumentStore(githubClient, githubAppParams.Owner, githubAppParams.Repo, githubAppParams.BaseBranch)

	// スレッドのメッセージを取得
	messages, _, _, err := slackClient.GetConversationReplies(&slack.GetConversationRepliesParameters{
		ChannelID: ev.Item.Channel,
		Timestamp: threadTimestamp,
	})
	if err != nil {
		h.logger.Error("Failed to get thread messages", zap.Error(err))
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
		DocumentStore:      documentStore,
	})
	draft, err := draftGenerateWorkflow.Execute(ctx, text)
	if err != nil {
		h.logger.Error("Failed to generate document", zap.Error(err))
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
