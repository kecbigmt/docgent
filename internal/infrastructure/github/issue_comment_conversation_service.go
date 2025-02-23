package github

import (
	"context"
	"docgent/internal/application/port"
	"fmt"

	"github.com/google/go-github/v68/github"
)

type IssueCommentConversationService struct {
	client          *github.Client
	owner           string
	repo            string
	prNumber        int
	sourceCommentID int64
	eyesReactionID  int64
}

func NewIssueCommentConversationService(client *github.Client, owner, repo string, prNumber int, sourceCommentID int64) port.ConversationService {
	return &IssueCommentConversationService{
		client:          client,
		owner:           owner,
		repo:            repo,
		prNumber:        prNumber,
		sourceCommentID: sourceCommentID,
	}
}

func (s *IssueCommentConversationService) GetHistory() ([]port.ConversationMessage, error) {
	ctx := context.Background()
	comments, _, err := s.client.PullRequests.ListComments(ctx, s.owner, s.repo, s.prNumber, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list review comments: %w", err)
	}

	conversationMessages := make([]port.ConversationMessage, 0, len(comments))
	for _, comment := range comments {
		conversationMessages = append(conversationMessages, port.ConversationMessage{
			Author:  *comment.User.Login,
			Content: *comment.Body,
		})
	}

	return conversationMessages, nil
}

func (s *IssueCommentConversationService) GetURI() string {
	// https://github.com/{owner}/{repo}/pull/{prNumber}#issuecomment-{sourceCommentID}
	return fmt.Sprintf("https://github.com/%s/%s/pull/%d#issuecomment-%d", s.owner, s.repo, s.prNumber, s.sourceCommentID)
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

func (s *IssueCommentConversationService) MarkEyes() error {
	ctx := context.Background()
	reaction, _, err := s.client.Reactions.CreateIssueCommentReaction(ctx, s.owner, s.repo, s.sourceCommentID, "eyes")
	if err != nil {
		return fmt.Errorf("failed to add eyes reaction to issue comment: %w", err)
	}

	s.eyesReactionID = reaction.GetID()

	return nil
}

func (s *IssueCommentConversationService) RemoveEyes() error {
	if s.eyesReactionID == 0 {
		return nil
	}

	ctx := context.Background()
	_, err := s.client.Reactions.DeleteIssueCommentReaction(ctx, s.owner, s.repo, s.sourceCommentID, s.eyesReactionID)
	if err != nil {
		return fmt.Errorf("failed to remove eyes reaction from issue comment: %w", err)
	}

	return nil
}
