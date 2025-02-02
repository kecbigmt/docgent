package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v68/github"

	"docgent-backend/internal/domain/autoagent"
)

type IssueCommentConversationService struct {
	client   *github.Client
	owner    string
	repo     string
	prNumber int
}

func NewIssueCommentConversationService(client *github.Client, owner, repo string, prNumber int) autoagent.ConversationService {
	return &IssueCommentConversationService{
		client:   client,
		owner:    owner,
		repo:     repo,
		prNumber: prNumber,
	}
}

func (s *IssueCommentConversationService) Reply(input string) error {
	ctx := context.Background()

	// IssueCommentの場合は新しいIssueCommentを作成
	comment := &github.IssueComment{
		Body: github.Ptr(input),
	}
	_, _, err := s.client.Issues.CreateComment(ctx, s.owner, s.repo, s.prNumber, comment)
	if err != nil {
		return fmt.Errorf("failed to create issue comment: %w", err)
	}

	return nil
}
