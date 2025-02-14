package diffutil

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/google/go-github/v68/github"
	"github.com/sergi/go-diff/diffmatchpatch"

	"docgent/internal/domain"
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

	if diff.OldName != diff.NewName {
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

	newText, results := dmp.PatchApply(patches, "")
	if !anyPatchApplied(results) {
		return fmt.Errorf("no changes were applied to the file")
	}
	opts := &github.RepositoryContentFileOptions{
		Message: github.Ptr(fmt.Sprintf("Create file %s", diff.NewName)),
		Content: []byte(newText),
		Branch:  github.Ptr(r.branchName),
	}

	_, _, err = r.client.Repositories.CreateFile(ctx, r.owner, r.repo, diff.NewName, opts)
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

	fileContent, _, _, err := r.client.Repositories.GetContents(
		ctx,
		r.owner,
		r.repo,
		diff.NewName,
		&github.RepositoryContentGetOptions{
			Ref: r.branchName,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to get file content: %w", err)
	}
	content, err := fileContent.GetContent()
	if err != nil {
		return fmt.Errorf("failed to get file content: %w", err)
	}

	// Base64デコードを実行
	decodedContent, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return fmt.Errorf("failed to decode base64 content: %w", err)
	}

	patchedText, results := dmp.PatchApply(patches, string(decodedContent))
	if !anyPatchApplied(results) {
		return fmt.Errorf("no changes were applied to the file")
	}

	opts := &github.RepositoryContentFileOptions{
		Message: github.Ptr(fmt.Sprintf("Update file %s", diff.NewName)),
		Content: []byte(patchedText),
		Branch:  github.Ptr(r.branchName),
		SHA:     fileContent.SHA,
	}

	_, _, err = r.client.Repositories.UpdateFile(ctx, r.owner, r.repo, diff.NewName, opts)
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

	fileContent, _, _, err := r.client.Repositories.GetContents(
		ctx,
		r.owner,
		r.repo,
		diff.NewName,
		&github.RepositoryContentGetOptions{
			Ref: r.branchName,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to get file content: %w", err)
	}
	content, err := fileContent.GetContent()
	if err != nil {
		return fmt.Errorf("failed to get file content: %w", err)
	}

	// Base64デコードを実行
	decodedContent, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return fmt.Errorf("failed to decode base64 content: %w", err)
	}

	patchedText, results := dmp.PatchApply(patches, string(decodedContent))
	if !anyPatchApplied(results) {
		return fmt.Errorf("no changes were applied to the file")
	}

	// Delete the old file
	deleteOpts := &github.RepositoryContentFileOptions{
		Message: github.Ptr(fmt.Sprintf("Delete file %s", diff.OldName)),
		Branch:  github.Ptr(r.branchName),
	}

	_, _, err = r.client.Repositories.DeleteFile(ctx, r.owner, r.repo, diff.OldName, deleteOpts)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	// Create the new file
	createOpts := &github.RepositoryContentFileOptions{
		Message: github.Ptr(fmt.Sprintf("Create file %s", diff.NewName)),
		Content: []byte(patchedText),
		Branch:  github.Ptr(r.branchName),
	}

	_, _, err = r.client.Repositories.CreateFile(ctx, r.owner, r.repo, diff.NewName, createOpts)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	return nil
}

// anyPatchApplied checks if any of the patches were successfully applied
func anyPatchApplied(results []bool) bool {
	for _, applied := range results {
		if applied {
			return true
		}
	}
	return false
}
