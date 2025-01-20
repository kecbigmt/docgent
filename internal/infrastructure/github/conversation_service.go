package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v68/github"

	"docgent-backend/internal/domain"
)

type CommentType int

const (
	IssueComment CommentType = iota
	ReviewComment
)

type ConversationService struct {
	client            *github.Client
	owner             string
	repo              string
	prNumber          int
	parentCommentID   int64
	parentCommentType CommentType
}

func NewConversationService(client *github.Client, owner, repo string, prNumber int, commentID int64, commentType CommentType) domain.ConversationService {
	return &ConversationService{
		client:            client,
		owner:             owner,
		repo:              repo,
		prNumber:          prNumber,
		parentCommentID:   commentID,
		parentCommentType: commentType,
	}
}

func (s *ConversationService) Reply(input string) error {
	ctx := context.Background()

	switch s.parentCommentType {
	case IssueComment:
		// IssueCommentの場合は新しいIssueCommentを作成
		comment := &github.IssueComment{
			Body: github.Ptr(input),
		}
		_, _, err := s.client.Issues.CreateComment(ctx, s.owner, s.repo, s.prNumber, comment)
		if err != nil {
			return fmt.Errorf("failed to create issue comment: %w", err)
		}

	case ReviewComment:
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

	default:
		return fmt.Errorf("unknown comment type")
	}

	return nil
}
