package github

import (
	"context"
	"fmt"

	"github.com/google/go-github/v68/github"

	"docgent-backend/internal/application"
)

// BranchService implements the application.BranchService interface using GitHub API.
type BranchService struct {
	client *github.Client
	owner  string
	repo   string
}

// NewBranchService creates a new BranchService instance.
func NewBranchService(
	client *github.Client,
	owner, repo string,
) application.BranchService {
	return &BranchService{
		client: client,
		owner:  owner,
		repo:   repo,
	}
}

// CreateBranch creates a new branch from the specified base branch.
func (s *BranchService) CreateBranch(ctx context.Context, baseBranchName, newBranchName string) (string, error) {
	// Get the base branch reference
	baseRef, _, err := s.client.Git.GetRef(
		ctx,
		s.owner,
		s.repo,
		fmt.Sprintf("refs/heads/%s", baseBranchName),
	)
	if err != nil {
		return "", fmt.Errorf("failed to get base branch reference: %w", err)
	}

	// Create a new reference for the new branch
	newRef := &github.Reference{
		Ref: github.Ptr(fmt.Sprintf("refs/heads/%s", newBranchName)),
		Object: &github.GitObject{
			SHA: baseRef.Object.SHA,
		},
	}

	_, _, err = s.client.Git.CreateRef(
		ctx,
		s.owner,
		s.repo,
		newRef,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create new branch: %w", err)
	}

	return newBranchName, nil
}
