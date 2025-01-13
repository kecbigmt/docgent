package github

import (
	"context"
	"docgent-backend/internal/domain"
	"fmt"
	"time"

	"github.com/google/go-github/v68/github"
)

type BranchAPI struct {
	client        *github.Client
	owner         string
	repo          string
	defaultBranch string
}

func NewBranchAPI(client *github.Client, owner, repo, defaultBranch string) *BranchAPI {
	return &BranchAPI{
		client:        client,
		owner:         owner,
		repo:          repo,
		defaultBranch: defaultBranch,
	}
}

func (s *BranchAPI) NewIncrementHandle(branchName string) domain.IncrementHandle {
	return domain.NewIncrementHandle("github-branch", branchName)
}

func (s *BranchAPI) IssueIncrementHandle() (domain.IncrementHandle, error) {
	return s.NewIncrementHandle(fmt.Sprintf("docgent/%d", time.Now().Unix())), nil
}

func (s *BranchAPI) CreateIncrement(increment domain.Increment) (domain.Increment, error) {
	ctx := context.Background()

	// 1. Get the SHA of the base branch
	baseBranchName := increment.PreviousHandle.Value
	ref, _, err := s.client.Git.GetRef(ctx, s.owner, s.repo, "refs/heads/"+baseBranchName)
	if err != nil {
		return domain.Increment{}, fmt.Errorf("failed to get ref: %w", err)
	}

	// 2. Create a new branch
	branchName := increment.Handle.Value
	newRef := &github.Reference{
		Ref: github.Ptr("refs/heads/" + branchName),
		Object: &github.GitObject{
			SHA: ref.Object.SHA,
		},
	}
	_, _, err = s.client.Git.CreateRef(ctx, s.owner, s.repo, newRef)
	if err != nil {
		return domain.Increment{}, fmt.Errorf("failed to create branch: %w", err)
	}

	return increment, nil
}

func (s *BranchAPI) AddDocumentChangeToIncrement(increment domain.Increment, change domain.DocumentChange) (domain.Increment, error) {
	ctx := context.Background()

	if change.Type != domain.DocumentCreateChange {
		return domain.Increment{}, fmt.Errorf("unsupported change type: %d", change.Type)
	}

	documentContent := change.DocumentContent

	// 3. Create or update file
	path := fmt.Sprintf("docs/%s.md", documentContent.Title)
	opts := &github.RepositoryContentFileOptions{
		Message: github.Ptr(fmt.Sprintf("Add document: %s", documentContent.Title)),
		Content: []byte(documentContent.Body),
		Branch:  github.Ptr(increment.Handle.Value),
	}

	_, _, err := s.client.Repositories.CreateFile(ctx, s.owner, s.repo, path, opts)
	if err != nil {
		return domain.Increment{}, fmt.Errorf("failed to create file: %w", err)
	}

	increment.DocumentChanges = append(increment.DocumentChanges, change)

	return increment, nil
}
