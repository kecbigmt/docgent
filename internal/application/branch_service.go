package application

import "context"

// BranchService is an interface for managing Git branch operations.
type BranchService interface {
	// CreateBranch creates a new branch from the specified base branch.
	CreateBranch(ctx context.Context, baseBranch, newBranchName string) (string, error)
}
