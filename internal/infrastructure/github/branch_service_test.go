package github

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/go-github/v68/github"
	"github.com/stretchr/testify/assert"
)

func TestBranchService_CreateBranch(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(*mockTransport)
		baseBranch    string
		newBranchName string
		wantErr       bool
		expectedReqs  []mockRequest
	}{
		{
			name: "success: create branch from main",
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"GET /repos/owner/repo/git/ref/heads/main": {
						statusCode: http.StatusOK,
						body: &github.Reference{
							Object: &github.GitObject{
								SHA: github.Ptr("abc123"),
							},
						},
					},
					"POST /repos/owner/repo/git/refs": {
						statusCode: http.StatusCreated,
						body:       &github.Reference{},
					},
				}
			},
			baseBranch:    "main",
			newBranchName: "docgent/test",
			wantErr:       false,
			expectedReqs: []mockRequest{
				{
					method: "GET",
					path:   "/repos/owner/repo/git/ref/heads/main",
				},
				{
					method: "POST",
					path:   "/repos/owner/repo/git/refs",
					body: map[string]interface{}{
						"ref": "refs/heads/docgent/test",
						"sha": "abc123",
					},
				},
			},
		},
		{
			name: "error: base branch not found",
			setup: func(mt *mockTransport) {
				mt.responses = map[string]mockResponse{
					"GET /repos/owner/repo/git/ref/heads/not-exist": {
						statusCode: http.StatusNotFound,
						body:       &github.ErrorResponse{},
					},
				}
			},
			baseBranch:    "not-exist",
			newBranchName: "docgent/test",
			wantErr:       true,
			expectedReqs: []mockRequest{
				{
					method: "GET",
					path:   "/repos/owner/repo/git/ref/heads/not-exist",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt := &mockTransport{
				responses:    make(map[string]mockResponse),
				expectedReqs: tt.expectedReqs,
			}
			tt.setup(mt)

			client := github.NewClient(&http.Client{Transport: mt})
			s := NewBranchService(client, "owner", "repo")

			branchName, err := s.CreateBranch(context.Background(), tt.baseBranch, tt.newBranchName)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.newBranchName, branchName)
			}

			mt.verify(t)
		})
	}
}
