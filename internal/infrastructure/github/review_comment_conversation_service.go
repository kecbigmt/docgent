package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v68/github"

	"docgent-backend/internal/domain/autoagent"
)

type ReviewCommentConversationService struct {
	client          *github.Client
	owner           string
	repo            string
	prNumber        int
	parentCommentID int64
}

func NewReviewCommentConversationService(client *github.Client, owner, repo string, prNumber int, commentID int64) autoagent.ConversationService {
	return &ReviewCommentConversationService{
		client:          client,
		owner:           owner,
		repo:            repo,
		prNumber:        prNumber,
		parentCommentID: commentID,
	}
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
		s.parentCommentID,
	)
	if err != nil {
		return fmt.Errorf("failed to create review comment reply: %w", err)
	}

	return nil
}
