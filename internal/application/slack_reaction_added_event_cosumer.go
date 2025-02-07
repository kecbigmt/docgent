package application

import (
	"context"
	"fmt"
	"time"

	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"docgent-backend/internal/domain"
	"docgent-backend/internal/workflow"
)

type SlackReactionAddedEventConsumerParams struct {
	fx.In

	Logger                *zap.Logger
	SlackAPI              SlackAPI
	GitHubServiceProvider GitHubServiceProvider
	DocumentAgent         domain.DocumentAgent
	ProposalAgent         domain.ProposalAgent
}

type SlackReactionAddedEventConsumer struct {
	logger                *zap.Logger
	slackAPI              SlackAPI
	githubServiceProvider GitHubServiceProvider
	documentAgent         domain.DocumentAgent
	proposalAgent         domain.ProposalAgent
}

func NewSlackReactionAddedEventConsumer(params SlackReactionAddedEventConsumerParams) *SlackReactionAddedEventConsumer {
	return &SlackReactionAddedEventConsumer{
		logger:                params.Logger,
		slackAPI:              params.SlackAPI,
		githubServiceProvider: params.GitHubServiceProvider,
		documentAgent:         params.DocumentAgent,
		proposalAgent:         params.ProposalAgent,
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

	ctx := context.Background()
	baseBranchName := githubAppParams.DefaultBranch
	newBranchName := fmt.Sprintf("docgent/%d", time.Now().Unix())

	branchService := h.githubServiceProvider.NewBranchService(githubAppParams.InstallationID, githubAppParams.Owner, githubAppParams.Repo)
	err = branchService.CreateBranch(ctx, baseBranchName, newBranchName)
	if err != nil {
		h.logger.Error("Failed to create branch", zap.Error(err))
		slackClient.PostMessage(ev.Item.Channel,
			slack.MsgOptionText(":warning: エラー: ブランチの作成に失敗しました", false),
			slack.MsgOptionTS(threadTimestamp),
		)
		return
	}

	githubPullRequestAPI := h.githubServiceProvider.NewPullRequestAPI(githubAppParams.InstallationID, githubAppParams.Owner, githubAppParams.Repo, baseBranchName, newBranchName)

	// スレッドの内容を結合
	var text string
	for _, msg := range messages {
		text += msg.Text + "\n"
	}

	// ドキュメントを生成
	proposalGenerateWorkflow := workflow.NewProposalGenerateWorkflow(
		h.documentAgent,
		h.proposalAgent,
		githubPullRequestAPI,
	)
	proposalHandle, err := proposalGenerateWorkflow.Execute(ctx, text)
	if err != nil {
		h.logger.Error("Failed to generate increment", zap.Error(err))
		slackClient.PostMessage(ev.Item.Channel,
			slack.MsgOptionText(":warning: エラー: ドキュメントの生成に失敗しました", false),
			slack.MsgOptionTS(threadTimestamp),
		)
		return
	}

	// 成功メッセージを投稿
	slackClient.PostMessage(ev.Item.Channel,
		slack.MsgOptionText(fmt.Sprintf(
			"ドキュメントを生成しました！\nPR: https://github.com/%s/%s/pull/%s",
			githubAppParams.Owner,
			githubAppParams.Repo,
			proposalHandle.Value,
		), false),
		slack.MsgOptionTS(threadTimestamp),
	)
}
