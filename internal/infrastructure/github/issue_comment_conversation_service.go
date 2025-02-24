package github

import (
	"context"
	"docgent/internal/application/port"
	"docgent/internal/domain/data"
	"fmt"

	"github.com/google/go-github/v68/github"
)

type IssueCommentConversationService struct {
	client         *github.Client
	ref            *IssueCommentRef
	eyesReactionID int64
}

func NewIssueCommentConversationService(client *github.Client, ref *IssueCommentRef) port.ConversationService {
	return &IssueCommentConversationService{
		client: client,
		ref:    ref,
	}
}

func (s *IssueCommentConversationService) GetHistory() ([]port.ConversationMessage, error) {
	ctx := context.Background()
	comments, _, err := s.client.PullRequests.ListComments(ctx, s.ref.Owner(), s.ref.Repo(), s.ref.PRNumber(), nil)
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

func (s *IssueCommentConversationService) URI() data.URI {
	return s.ref.ToURI()
}

func (s *IssueCommentConversationService) Reply(input string) error {
	ctx := context.Background()

	// IssueCommentの場合は新しいIssueCommentを作成
	comment := &github.IssueComment{
		Body: github.Ptr(input),
	}
	_, _, err := s.client.Issues.CreateComment(ctx, s.ref.Owner(), s.ref.Repo(), s.ref.PRNumber(), comment)
	if err != nil {
		return fmt.Errorf("failed to create issue comment: %w", err)
	}

	return nil
}

func (s *IssueCommentConversationService) MarkEyes() error {
	ctx := context.Background()
	reaction, _, err := s.client.Reactions.CreateIssueCommentReaction(ctx, s.ref.Owner(), s.ref.Repo(), s.ref.SourceCommentID(), "eyes")
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
	_, err := s.client.Reactions.DeleteIssueCommentReaction(ctx, s.ref.Owner(), s.ref.Repo(), s.ref.SourceCommentID(), s.eyesReactionID)
	if err != nil {
		return fmt.Errorf("failed to remove eyes reaction from issue comment: %w", err)
	}

	return nil
}
