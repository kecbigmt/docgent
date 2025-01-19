package github

import (
	"context"
	"docgent-backend/internal/domain"
	"fmt"
	"strconv"
	"time"

	"github.com/google/go-github/v68/github"
	"github.com/sergi/go-diff/diffmatchpatch"
)

type PullRequestAPI struct {
	client        *github.Client
	owner         string
	repo          string
	defaultBranch string
}

func NewPullRequestAPI(client *github.Client, owner, repo, defaultBranch string) *PullRequestAPI {
	return &PullRequestAPI{
		client:        client,
		owner:         owner,
		repo:          repo,
		defaultBranch: defaultBranch,
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

	// 1. Get the SHA of the base branch
	baseBranchName := s.defaultBranch
	ref, _, err := s.client.Git.GetRef(ctx, s.owner, s.repo, "refs/heads/"+baseBranchName)
	if err != nil {
		return domain.ProposalHandle{}, fmt.Errorf("failed to get ref: %w", err)
	}

	// 2. Create a new branch
	branchName := fmt.Sprintf("docgent/%d", time.Now().Unix())
	newRef := &github.Reference{
		Ref: github.Ptr("refs/heads/" + branchName),
		Object: &github.GitObject{
			SHA: ref.Object.SHA,
		},
	}
	_, _, err = s.client.Git.CreateRef(ctx, s.owner, s.repo, newRef)
	if err != nil {
		return domain.ProposalHandle{}, fmt.Errorf("failed to create branch: %w", err)
	}

	for _, diff := range diffs {
		// 3. Parse the unified diff
		dmp := diffmatchpatch.New()
		patches, err := dmp.PatchFromText(string(diff.Body))
		if err != nil {
			return domain.ProposalHandle{}, fmt.Errorf("failed to parse diff: %w", err)
		}

		if diff.IsNewFile {
			newText, _ := dmp.PatchApply(patches, "")
			opts := &github.RepositoryContentFileOptions{
				Message: github.Ptr(fmt.Sprintf("Create file %s", diff.NewFilename)),
				Content: []byte(newText),
				Branch:  github.Ptr(branchName),
			}

			_, _, err = s.client.Repositories.CreateFile(ctx, s.owner, s.repo, "docs/"+diff.NewFilename, opts)
			if err != nil {
				return domain.ProposalHandle{}, fmt.Errorf("failed to create file: %w", err)
			}
		} else {
			// Update the file
			fileContent, _, _, err := s.client.Repositories.GetContents(ctx, s.owner, s.repo, "docs/"+diff.NewFilename, nil)
			if err != nil {
				return domain.ProposalHandle{}, fmt.Errorf("failed to get file content: %w", err)
			}
			content, err := fileContent.GetContent()
			if err != nil {
				return domain.ProposalHandle{}, fmt.Errorf("failed to get file content: %w", err)
			}

			// Apply the patch
			patchedText, _ := dmp.PatchApply(patches, content)

			if diff.OldFilename != diff.NewFilename {
				// Delete the old file
				opts := &github.RepositoryContentFileOptions{
					Message: github.Ptr(fmt.Sprintf("Delete file %s", diff.OldFilename)),
					Branch:  github.Ptr(branchName),
				}
				_, _, err = s.client.Repositories.DeleteFile(ctx, s.owner, s.repo, "docs/"+diff.OldFilename, opts)
				if err != nil {
					return domain.ProposalHandle{}, fmt.Errorf("failed to delete file: %w", err)
				}

				// Create the new file
				opts = &github.RepositoryContentFileOptions{
					Message: github.Ptr(fmt.Sprintf("Create file %s", diff.NewFilename)),
					Content: []byte(patchedText),
					Branch:  github.Ptr(branchName),
				}

				_, _, err = s.client.Repositories.CreateFile(ctx, s.owner, s.repo, "docs/"+diff.NewFilename, opts)
				if err != nil {
					return domain.ProposalHandle{}, fmt.Errorf("failed to create file: %w", err)
				}
			} else {
				opts := &github.RepositoryContentFileOptions{
					Message: github.Ptr(fmt.Sprintf("Update file %s", diff.NewFilename)),
					Content: []byte(patchedText),
					Branch:  github.Ptr(branchName),
				}

				_, _, err = s.client.Repositories.UpdateFile(ctx, s.owner, s.repo, "docs/"+diff.NewFilename, opts)
				if err != nil {
					return domain.ProposalHandle{}, fmt.Errorf("failed to update file: %w", err)
				}
			}
		}
	}

	// 5. Create Pull Request
	newPR := &github.NewPullRequest{
		Title: github.Ptr(content.Title),
		Body:  github.Ptr(content.Body),
		Head:  github.Ptr(branchName),
		Base:  github.Ptr(baseBranchName),
	}

	pr, _, err := s.client.PullRequests.Create(ctx, s.owner, s.repo, newPR)
	if err != nil {
		return domain.ProposalHandle{}, fmt.Errorf("failed to create pull request: %w", err)
	}

	handle := s.NewProposalHandle(fmt.Sprintf("%d", pr.GetNumber()))
	return handle, nil
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

func (s *PullRequestAPI) UpdateProposal(proposalHandle domain.ProposalHandle, content domain.ProposalContent) error {
	ctx := context.Background()

	number, err := strconv.Atoi(proposalHandle.Value)
	if err != nil {
		return fmt.Errorf("failed to convert pull request number: %w", err)
	}

	// Get current PR to preserve other fields
	pr, _, err := s.client.PullRequests.Get(ctx, s.owner, s.repo, number)
	if err != nil {
		return fmt.Errorf("failed to get pull request: %w", err)
	}

	// Update PR with new content
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
