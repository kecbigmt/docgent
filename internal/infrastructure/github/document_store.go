package github

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v68/github"
	"golang.org/x/oauth2"

	"docgent-backend/internal/model/infrastructure"
)

type DocumentStore struct {
	client     *github.Client
	owner      string
	repo       string
	baseBranch string
}

func NewDocumentStore(config APIConfig) (*DocumentStore, error) {

	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.Token},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)

	return &DocumentStore{
		client:     client,
		owner:      config.Owner,
		repo:       config.Repo,
		baseBranch: config.BaseBranch,
	}, nil
}

func (s *DocumentStore) Save(documentInput infrastructure.DocumentInput) (infrastructure.Document, error) {
	ctx := context.Background()

	// 1. Get the SHA of the base branch
	ref, _, err := s.client.Git.GetRef(ctx, s.owner, s.repo, "refs/heads/"+s.baseBranch)
	if err != nil {
		return infrastructure.Document{}, fmt.Errorf("failed to get ref: %w", err)
	}

	// 2. Create a new branch
	branchName := fmt.Sprintf("docs/%s-%d", documentInput.Title, time.Now().Unix())
	newRef := &github.Reference{
		Ref: github.Ptr("refs/heads/" + branchName),
		Object: &github.GitObject{
			SHA: ref.Object.SHA,
		},
	}
	_, _, err = s.client.Git.CreateRef(ctx, s.owner, s.repo, newRef)
	if err != nil {
		return infrastructure.Document{}, fmt.Errorf("failed to create branch: %w", err)
	}

	// 3. Create or update file
	path := fmt.Sprintf("docs/%s.md", documentInput.Title)
	opts := &github.RepositoryContentFileOptions{
		Message: github.Ptr(fmt.Sprintf("Add document: %s", documentInput.Title)),
		Content: []byte(documentInput.Content),
		Branch:  github.Ptr(branchName),
	}

	_, _, err = s.client.Repositories.CreateFile(ctx, s.owner, s.repo, path, opts)
	if err != nil {
		return infrastructure.Document{}, fmt.Errorf("failed to create file: %w", err)
	}

	document := infrastructure.Document{
		ID:      path,
		Title:   documentInput.Title,
		Content: documentInput.Content,
	}

	return document, nil
}
