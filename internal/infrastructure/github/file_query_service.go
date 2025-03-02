package github

import (
	"context"
	"fmt"
	"regexp"
	"sync"

	"docgent/internal/application/port"
	"docgent/internal/domain/data"

	"github.com/google/go-github/v68/github"
)

type FileQueryService struct {
	client *github.Client
	owner  string
	repo   string
	branch string

	// Cache for commit SHA
	commitSHA     string
	commitSHALock sync.RWMutex
}

func NewFileQueryService(client *github.Client, owner, repo, branch string) *FileQueryService {
	return &FileQueryService{
		client: client,
		owner:  owner,
		repo:   repo,
		branch: branch,
	}
}

func (s *FileQueryService) FindFile(ctx context.Context, path string) (data.File, error) {
	// ファイルの内容を取得
	fileContent, _, _, err := s.client.Repositories.GetContents(
		ctx,
		s.owner,
		s.repo,
		path,
		&github.RepositoryContentGetOptions{
			Ref: s.branch,
		},
	)
	if err != nil {
		if _, ok := err.(*github.ErrorResponse); ok && err.(*github.ErrorResponse).Response.StatusCode == 404 {
			return data.File{}, port.ErrFileNotFound
		}
		return data.File{}, fmt.Errorf("failed to get file contents: %w", err)
	}

	// ファイルの内容をデコード
	content, err := fileContent.GetContent()
	if err != nil {
		return data.File{}, fmt.Errorf("failed to decode file content: %w", err)
	}

	return data.File{
		Path:    path,
		Content: content,
	}, nil
}

func (s *FileQueryService) GetTree(ctx context.Context, options ...port.GetTreeOption) ([]port.TreeMetadata, error) {
	treeOptions := &port.GetTreeOptions{
		Recursive: false,
		TreeSHA:   "refs/heads/" + s.branch,
	}
	for _, option := range options {
		option(treeOptions)
	}

	tree, _, err := s.client.Git.GetTree(ctx, s.owner, s.repo, treeOptions.TreeSHA, treeOptions.Recursive)
	if err != nil {
		return nil, fmt.Errorf("failed to get tree: %w", err)
	}

	treeMetadata := make([]port.TreeMetadata, 0)
	for _, entry := range tree.Entries {
		treeType := port.NodeTypeFile
		if entry.GetType() == "tree" {
			treeType = port.NodeTypeDirectory
		}
		treeMetadata = append(treeMetadata, port.TreeMetadata{
			Type: treeType,
			SHA:  entry.GetSHA(),
			Path: entry.GetPath(),
			Size: entry.GetSize(),
		})
	}
	return treeMetadata, nil
}

// GetURI returns a GitHub permalink for the given file path
func (s *FileQueryService) GetURI(ctx context.Context, path string) (*data.URI, error) {
	// Get the commit SHA from cache or API
	commitSHA, err := s.getCommitSHA(ctx)
	if err != nil {
		return nil, err
	}

	// Construct the GitHub permalink
	// Format: https://github.com/{owner}/{repo}/blob/{commitSHA}/{path}
	rawURI := fmt.Sprintf("https://github.com/%s/%s/blob/%s/%s",
		s.owner, s.repo, commitSHA, path)

	uri, err := data.NewURI(rawURI)
	if err != nil {
		return nil, fmt.Errorf("failed to create URI: %w", err)
	}

	return uri, nil
}

// getCommitSHA retrieves the commit SHA for the branch, using cache if available
func (s *FileQueryService) getCommitSHA(ctx context.Context) (string, error) {
	// Try to get from cache first
	s.commitSHALock.RLock()
	cachedSHA := s.commitSHA
	s.commitSHALock.RUnlock()

	if cachedSHA != "" {
		return cachedSHA, nil
	}

	// Not in cache, get from API
	ref, _, err := s.client.Git.GetRef(ctx, s.owner, s.repo, "refs/heads/"+s.branch)
	if err != nil {
		return "", fmt.Errorf("failed to get ref: %w", err)
	}

	commitSHA := ref.GetObject().GetSHA()

	// Store in cache
	s.commitSHALock.Lock()
	s.commitSHA = commitSHA
	s.commitSHALock.Unlock()

	return commitSHA, nil
}

var reFileURI = regexp.MustCompile(`^https://github\.com/(?P<owner>[^/]+)/(?P<repo>[^/]+)/blob/(?P<commitSHA>[^/]+)/(?P<path>[^?]+)`)
var reFileURISubNames = reFileURI.SubexpNames()

func (s *FileQueryService) GetFilePath(uri *data.URI) (string, error) {
	// Extract owner, repo, commitSHA, and path from the URI
	matches := reFileURI.FindStringSubmatch(uri.String())
	if matches == nil {
		return "", fmt.Errorf("invalid GitHub URI: %s", uri)
	}

	// Extract the named subgroups from the regex
	subgroups := make(map[string]string)
	for i, name := range reFileURISubNames {
		if i != 0 && name != "" {
			subgroups[name] = matches[i]
		}
	}

	// Validate the extracted values
	if subgroups["owner"] != s.owner || subgroups["repo"] != s.repo {
		return "", fmt.Errorf("URI does not match the repository: %s", uri)
	}

	return subgroups["path"], nil
}
