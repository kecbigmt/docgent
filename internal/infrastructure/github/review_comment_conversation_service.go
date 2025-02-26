package github

import (
	"context"
	"docgent/internal/application/port"
	"docgent/internal/domain/data"
	"fmt"

	"github.com/google/go-github/v68/github"
)

type ReviewCommentConversationService struct {
	client          *github.Client
	owner           string
	repo            string
	prNumber        int
	sourceCommentID int64
	eyesReactionID  int64
	fromUserID      string // ソースコメントの作者のID
}

func NewReviewCommentConversationService(client *github.Client, owner, repo string, prNumber int, sourceCommentID int64, fromUserID string) port.ConversationService {
	return &ReviewCommentConversationService{
		client:          client,
		owner:           owner,
		repo:            repo,
		prNumber:        prNumber,
		sourceCommentID: sourceCommentID,
		fromUserID:      fromUserID,
	}
}

func (s *ReviewCommentConversationService) GetHistory() ([]port.ConversationMessage, error) {
	panic("not implemented")
}

func (s *ReviewCommentConversationService) URI() *data.URI {
	panic("not implemented")
}

func (s *ReviewCommentConversationService) Reply(input string, withMention bool) error {
	ctx := context.Background()

	message := input
	if withMention && s.fromUserID != "" {
		message = fmt.Sprintf("@%s %s", s.fromUserID, input)
	}

	// ReviewCommentの場合は返信として新しいReviewCommentを作成
	_, _, err := s.client.PullRequests.CreateCommentInReplyTo(
		ctx,
		s.owner,
		s.repo,
		s.prNumber,
		message,
		s.sourceCommentID,
	)
	if err != nil {
		return fmt.Errorf("failed to create review comment reply: %w", err)
	}

	return nil
}

func (s *ReviewCommentConversationService) MarkEyes() error {
	ctx := context.Background()
	reaction, _, err := s.client.Reactions.CreatePullRequestCommentReaction(ctx, s.owner, s.repo, s.sourceCommentID, "eyes")
	if err != nil {
		return fmt.Errorf("failed to add eyes reaction to review comment: %w", err)
	}

	s.eyesReactionID = reaction.GetID()

	return nil
}

func (s *ReviewCommentConversationService) RemoveEyes() error {
	if s.eyesReactionID == 0 {
		return nil
	}

	ctx := context.Background()
	_, err := s.client.Reactions.DeletePullRequestCommentReaction(ctx, s.owner, s.repo, s.sourceCommentID, s.eyesReactionID)
	if err != nil {
		return fmt.Errorf("failed to remove eyes reaction from review comment: %w", err)
	}

	return nil
}
