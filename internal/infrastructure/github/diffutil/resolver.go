package diffutil

import (
	"context"
	"fmt"

	"github.com/google/go-github/v68/github"
	"github.com/sergi/go-diff/diffmatchpatch"

	"docgent-backend/internal/domain"
)

type Resolver struct {
	client     *github.Client
	owner      string
	repo       string
	branchName string
}

func NewResolver(client *github.Client, owner, repo, branchName string) *Resolver {
	return &Resolver{
		client:     client,
		owner:      owner,
		repo:       repo,
		branchName: branchName,
	}
}

func (r *Resolver) Execute(diff domain.Diff) error {
	if diff.IsNewFile {
		return r.resolveCreateDiff(diff)
	}

	if diff.OldPath != diff.NewPath {
		return r.resolveUpdateDiffWithRename(diff)
	}

	return r.resolveUpdateDiffWithoutRename(diff)
}

func (r *Resolver) resolveCreateDiff(diff domain.Diff) error {
	ctx := context.Background()
	dmp := diffmatchpatch.New()
	patches, err := dmp.PatchFromText(string(diff.Body))
	if err != nil {
		return fmt.Errorf("failed to parse diff: %w", err)
	}

	newText, _ := dmp.PatchApply(patches, "")
	opts := &github.RepositoryContentFileOptions{
		Message: github.Ptr(fmt.Sprintf("Create file %s", diff.NewPath)),
		Content: []byte(newText),
		Branch:  github.Ptr(r.branchName),
	}

	_, _, err = r.client.Repositories.CreateFile(ctx, r.owner, r.repo, "docs/"+diff.NewPath, opts)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	return nil
}

func (r *Resolver) resolveUpdateDiffWithoutRename(diff domain.Diff) error {
	ctx := context.Background()
	dmp := diffmatchpatch.New()
	patches, err := dmp.PatchFromText(string(diff.Body))
	if err != nil {
		return fmt.Errorf("failed to parse diff: %w", err)
	}

	fileContent, _, _, err := r.client.Repositories.GetContents(ctx, r.owner, r.repo, "docs/"+diff.NewPath, nil)
	if err != nil {
		return fmt.Errorf("failed to get file content: %w", err)
	}
	content, err := fileContent.GetContent()
	if err != nil {
		return fmt.Errorf("failed to get file content: %w", err)
	}

	patchedText, _ := dmp.PatchApply(patches, content)
	opts := &github.RepositoryContentFileOptions{
		Message: github.Ptr(fmt.Sprintf("Update file %s", diff.NewPath)),
		Content: []byte(patchedText),
		Branch:  github.Ptr(r.branchName),
	}

	_, _, err = r.client.Repositories.UpdateFile(ctx, r.owner, r.repo, "docs/"+diff.NewPath, opts)
	if err != nil {
		return fmt.Errorf("failed to update file: %w", err)
	}

	return nil
}

func (r *Resolver) resolveUpdateDiffWithRename(diff domain.Diff) error {
	ctx := context.Background()
	dmp := diffmatchpatch.New()
	patches, err := dmp.PatchFromText(string(diff.Body))
	if err != nil {
		return fmt.Errorf("failed to parse diff: %w", err)
	}

	fileContent, _, _, err := r.client.Repositories.GetContents(ctx, r.owner, r.repo, "docs/"+diff.NewPath, nil)
	if err != nil {
		return fmt.Errorf("failed to get file content: %w", err)
	}
	content, err := fileContent.GetContent()
	if err != nil {
		return fmt.Errorf("failed to get file content: %w", err)
	}

	patchedText, _ := dmp.PatchApply(patches, content)

	// Delete the old file
	deleteOpts := &github.RepositoryContentFileOptions{
		Message: github.Ptr(fmt.Sprintf("Delete file %s", diff.OldPath)),
		Branch:  github.Ptr(r.branchName),
	}
	_, _, err = r.client.Repositories.DeleteFile(ctx, r.owner, r.repo, "docs/"+diff.OldPath, deleteOpts)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	// Create the new file
	createOpts := &github.RepositoryContentFileOptions{
		Message: github.Ptr(fmt.Sprintf("Create file %s", diff.NewPath)),
		Content: []byte(patchedText),
		Branch:  github.Ptr(r.branchName),
	}

	_, _, err = r.client.Repositories.CreateFile(ctx, r.owner, r.repo, "docs/"+diff.NewPath, createOpts)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	return nil
}
