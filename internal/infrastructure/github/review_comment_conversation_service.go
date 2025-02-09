package github

import (
	"context"
	"docgent-backend/internal/application/port"
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
}

func NewReviewCommentConversationService(client *github.Client, owner, repo string, prNumber int, sourceCommentID int64) port.ConversationService {
	return &ReviewCommentConversationService{
		client:          client,
		owner:           owner,
		repo:            repo,
		prNumber:        prNumber,
		sourceCommentID: sourceCommentID,
	}
}

func (s *ReviewCommentConversationService) GetHistory() ([]port.ConversationMessage, error) {
	panic("not implemented")
}

func (s *ReviewCommentConversationService) Reply(input string) error {
	ctx := context.Background()

	// ReviewCommentの場合は返信として新しいReviewCommentを作成
	_, _, err := s.client.PullRequests.CreateCommentInReplyTo(
		ctx,
		s.owner,
		s.repo,
		s.prNumber,
		input,
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
