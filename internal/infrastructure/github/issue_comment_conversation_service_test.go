package github

import (
	"testing"

	"github.com/google/go-github/v68/github"
	"github.com/stretchr/testify/assert"
)

func TestIssueCommentConversationService_URI(t *testing.T) {
	// テストケースの準備
	service := &IssueCommentConversationService{
		client: github.NewClient(nil),
		ref:    NewIssueCommentRef("kecbigmt", "docgent", 123, 456789),
	}

	// テストの実行
	got := service.URI()

	// 期待値の検証
	want := "https://github.com/kecbigmt/docgent/pull/123#issuecomment-456789"
	assert.Equal(t, want, got.String())
}
