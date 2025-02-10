package github

import (
	"context"
	"fmt"
	"strconv"

	"github.com/google/go-github/v68/github"

	"docgent/internal/domain"
	"docgent/internal/infrastructure/github/diffutil"
)

type PullRequestAPI struct {
	client        *github.Client
	owner         string
	repo          string
	defaultBranch string
	headBranch    string
}

func NewPullRequestAPI(client *github.Client, owner, repo, defaultBranch string, headBranch string) *PullRequestAPI {
	return &PullRequestAPI{
		client:        client,
		owner:         owner,
		repo:          repo,
		defaultBranch: defaultBranch,
		headBranch:    headBranch,
	}
}

func (s *PullRequestAPI) NewProposalHandle(value string) domain.ProposalHandle {
	return domain.NewProposalHandle("github-pull-request", value)
}

func (s *PullRequestAPI) NewCommentHandle(issueCommentID string) domain.CommentHandle {
	return domain.NewCommentHandle("github-issue-comment", issueCommentID)
}

func (s *PullRequestAPI) CreateProposal(diffs domain.Diffs, content domain.ProposalContent) (domain.ProposalHandle, error) {
	ctx := context.Background()

	for _, diff := range diffs {
		resolver := diffutil.NewResolver(s.client, s.owner, s.repo, s.headBranch)
		if err := resolver.Execute(diff); err != nil {
			return domain.ProposalHandle{}, fmt.Errorf("failed to resolve diff: %w", err)
		}
	}

	newPR := &github.NewPullRequest{
		Title: github.Ptr(content.Title),
		Body:  github.Ptr(content.Body),
		Head:  github.Ptr(s.headBranch),
		Base:  github.Ptr(s.defaultBranch),
	}

	pr, _, err := s.client.PullRequests.Create(ctx, s.owner, s.repo, newPR)
	if err != nil {
		return domain.ProposalHandle{}, fmt.Errorf("failed to create pull request: %w", err)
	}

	handle := s.NewProposalHandle(fmt.Sprintf("%d", pr.GetNumber()))
	return handle, nil
}

func (s *PullRequestAPI) GetProposal(handle domain.ProposalHandle) (domain.Proposal, error) {
	ctx := context.Background()

	number, err := s.parseHandle(handle)
	if err != nil {
		return domain.Proposal{}, err
	}

	pr, _, err := s.client.PullRequests.Get(ctx, s.owner, s.repo, number)
	if err != nil {
		return domain.Proposal{}, fmt.Errorf("failed to get pull request: %w", err)
	}

	diff, _, err := s.client.PullRequests.GetRaw(ctx, s.owner, s.repo, number, github.RawOptions{Type: github.Diff})
	if err != nil {
		return domain.Proposal{}, fmt.Errorf("failed to get pull request diff: %w", err)
	}

	parser := diffutil.NewParser()
	diffs := parser.Execute(diff)

	comments, _, err := s.client.Issues.ListComments(ctx, s.owner, s.repo, number, nil)
	if err != nil {
		return domain.Proposal{}, fmt.Errorf("failed to get pull request comments: %w", err)
	}

	domainComments := make([]domain.Comment, len(comments))
	for i, comment := range comments {
		handle := s.NewCommentHandle(strconv.FormatInt(comment.GetID(), 10))
		domainComments[i] = domain.NewComment(handle, comment.GetUser().GetLogin(), comment.GetBody())
	}

	content := domain.ProposalContent{
		Title: pr.GetTitle(),
		Body:  pr.GetBody(),
	}

	return domain.NewProposal(handle, diffs, content, domainComments), nil
}

func (s *PullRequestAPI) CreateComment(proposalHandle domain.ProposalHandle, commentBody string) (domain.Comment, error) {
	ctx := context.Background()

	newComment := &github.IssueComment{
		Body: github.Ptr(commentBody),
	}

	number, err := s.parseHandle(proposalHandle)
	if err != nil {
		return domain.Comment{}, fmt.Errorf("failed to parse handle: %w", err)
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

func (s *PullRequestAPI) ApplyProposalDiffs(handle domain.ProposalHandle, diffs domain.Diffs) error {
	ctx := context.Background()

	number, err := s.parseHandle(handle)
	if err != nil {
		return err
	}

	pr, _, err := s.client.PullRequests.Get(ctx, s.owner, s.repo, number)
	if err != nil {
		return fmt.Errorf("failed to get pull request: %w", err)
	}

	branchName := pr.Head.GetRef()

	for _, diff := range diffs {
		resolver := diffutil.NewResolver(s.client, s.owner, s.repo, branchName)
		if err := resolver.Execute(diff); err != nil {
			return fmt.Errorf("failed to resolve diff: %w", err)
		}
	}

	return nil
}

func (s *PullRequestAPI) UpdateProposalContent(proposalHandle domain.ProposalHandle, content domain.ProposalContent) error {
	ctx := context.Background()

	number, err := s.parseHandle(proposalHandle)
	if err != nil {
		return err
	}

	pr, _, err := s.client.PullRequests.Get(ctx, s.owner, s.repo, number)
	if err != nil {
		return fmt.Errorf("failed to get pull request: %w", err)
	}

	updatePR := &github.PullRequest{
		Title: github.Ptr(content.Title),
		Body:  github.Ptr(content.Body),
		Base:  pr.Base,
		Head:  pr.Head,
	}

	_, _, err = s.client.PullRequests.Edit(ctx, s.owner, s.repo, number, updatePR)
	if err != nil {
		return fmt.Errorf("failed to update pull request: %w", err)
	}

	return nil
}

func (s *PullRequestAPI) parseHandle(handle domain.ProposalHandle) (int, error) {
	number, err := strconv.Atoi(handle.Value)
	if err != nil {
		return 0, fmt.Errorf("failed to parse proposal handle to pull request number: %w", err)
	}
	return number, nil
}
