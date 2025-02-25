package github

import (
	"context"
	"docgent/internal/domain/data"
	"fmt"
	"strings"

	"github.com/google/go-github/v68/github"
)

type SourceRepository struct {
	client *github.Client
}

func NewSourceRepository(client *github.Client) *SourceRepository {
	return &SourceRepository{client: client}
}

func (r *SourceRepository) Match(uri *data.URI) bool {
	return uri.Host() == "github.com"
}

func (r *SourceRepository) Find(ctx context.Context, uri *data.URI) (*data.Source, error) {
	ref, err := ParseIssueCommentRef(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to parse issue comment ref: %w", err)
	}

	// PRのコメント一覧を取得
	comments, _, err := r.client.Issues.ListComments(ctx, ref.Owner(), ref.Repo(), ref.PRNumber(), &github.IssueListCommentsOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list PR comments: %w", err)
	}

	var content strings.Builder
	content.WriteString(fmt.Sprintf("<conversation uri=%q>\n", uri))

	for _, comment := range comments {
		// コメントの作成者名を取得
		var user string
		if comment.User != nil && comment.User.Login != nil {
			user = *comment.User.Login
		} else {
			user = "unknown"
		}

		// コメントの本文を取得
		var body string
		if comment.Body != nil {
			body = *comment.Body
		} else {
			body = ""
		}

		// 特定のコメントの場合はハイライト属性を付与
		if comment.ID != nil && *comment.ID == ref.SourceCommentID() {
			content.WriteString(fmt.Sprintf("<message user=%q highlighted=\"true\">\n%s\n</message>\n", user, body))
		} else {
			content.WriteString(fmt.Sprintf("<message user=%q>\n%s\n</message>\n", user, body))
		}
	}

	content.WriteString("</conversation>")

	return data.NewSource(uri, content.String()), nil
}
