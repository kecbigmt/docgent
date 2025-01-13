package github

import (
	"context"
	"docgent-backend/internal/domain"
	"fmt"
	"strconv"

	"github.com/google/go-github/v68/github"
)

type PullRequestAPI struct {
	client *github.Client
	owner  string
	repo   string
}

func NewPullRequestAPI(client *github.Client, owner, repo string) *PullRequestAPI {
	return &PullRequestAPI{
		client: client,
		owner:  owner,
		repo:   repo,
	}
}

func (s *PullRequestAPI) NewProposalHandle(value string) domain.ProposalHandle {
	return domain.NewProposalHandle("github-pull-request", value)
}

func (s *PullRequestAPI) NewCommentHandle(issueCommentID string) domain.CommentHandle {
	return domain.NewCommentHandle("github-issue-comment", issueCommentID)
}

func (s *PullRequestAPI) CreateProposal(content domain.ProposalContent, increment domain.Increment) (domain.Proposal, error) {
	ctx := context.Background()

	// Create Pull Request
	newPR := &github.NewPullRequest{
		Title: github.Ptr(content.Title),
		Body:  github.Ptr(content.Body),
		Head:  github.Ptr(increment.Handle.Value),
		Base:  github.Ptr(increment.PreviousHandle.Value),
	}

	pr, _, err := s.client.PullRequests.Create(ctx, s.owner, s.repo, newPR)
	if err != nil {
		return domain.Proposal{}, fmt.Errorf("failed to create pull request: %w", err)
	}

	handle := s.NewProposalHandle(fmt.Sprintf("%d", pr.GetNumber()))
	return domain.NewProposal(handle, content, increment), nil
}

func (s *PullRequestAPI) CreateComment(proposalHandle domain.ProposalHandle, commentBody string) (domain.Comment, error) {
	ctx := context.Background()

	newComment := &github.IssueComment{
		Body: github.Ptr(commentBody),
	}

	number, err := strconv.Atoi(proposalHandle.Value)
	if err != nil {
		return domain.Comment{}, fmt.Errorf("failed to convert pull request number: %w", err)
	}

	issueComment, _, err := s.client.Issues.CreateComment(ctx, s.owner, s.repo, number, newComment)
	if err != nil {
		return domain.Comment{}, fmt.Errorf("failed to add comment: %w", err)
	}

	author := issueComment.GetUser().GetLogin()
	handle := s.NewCommentHandle(strconv.FormatInt(issueComment.GetID(), 10))
	comment := domain.NewComment(handle, author, commentBody)

	return comment, nil
}
