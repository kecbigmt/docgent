package handler

import (
	"testing"
	"time"

	"github.com/google/go-github/v68/github"
	"github.com/stretchr/testify/assert"
)

func TestClassifyFilesBySimulation(t *testing.T) {
	tests := []struct {
		name         string
		commits      []*github.HeadCommit
		wantNewFiles []string
		wantModified []string
		wantDeleted  []string
	}{
		{
			name: "success: single commit with added, modified, and deleted files",
			commits: []*github.HeadCommit{
				{
					Timestamp: &github.Timestamp{Time: time.Now()},
					Added:     []string{"new.txt"},
					Modified:  []string{"modified.txt"},
					Removed:   []string{"deleted.txt"},
				},
			},
			wantNewFiles: []string{"new.txt"},
			wantModified: []string{"modified.txt"},
			wantDeleted:  []string{"deleted.txt"},
		},
		{
			name: "success: same file is modified across multiple commits",
			commits: []*github.HeadCommit{
				{
					Timestamp: &github.Timestamp{Time: time.Now().Add(-2 * time.Hour)},
					Added:     []string{"file.txt"},
				},
				{
					Timestamp: &github.Timestamp{Time: time.Now().Add(-1 * time.Hour)},
					Modified:  []string{"file.txt"},
				},
			},
			wantNewFiles: []string{"file.txt"},
			wantModified: []string{},
			wantDeleted:  []string{},
		},
		{
			name: "success: file is added and then deleted",
			commits: []*github.HeadCommit{
				{
					Timestamp: &github.Timestamp{Time: time.Now().Add(-2 * time.Hour)},
					Added:     []string{"temp.txt"},
				},
				{
					Timestamp: &github.Timestamp{Time: time.Now().Add(-1 * time.Hour)},
					Removed:   []string{"temp.txt"},
				},
			},
			wantNewFiles: []string{},
			wantModified: []string{},
			wantDeleted:  []string{},
		},
		{
			name: "success: deleted file is re-added",
			commits: []*github.HeadCommit{
				{
					Timestamp: &github.Timestamp{Time: time.Now().Add(-2 * time.Hour)},
					Removed:   []string{"file.txt"},
				},
				{
					Timestamp: &github.Timestamp{Time: time.Now().Add(-1 * time.Hour)},
					Added:     []string{"file.txt"},
				},
			},
			wantNewFiles: []string{},
			wantModified: []string{"file.txt"},
			wantDeleted:  []string{},
		},
		{
			name: "success: file lifecycle - added, modified, and then deleted",
			commits: []*github.HeadCommit{
				{
					Timestamp: &github.Timestamp{Time: time.Now().Add(-3 * time.Hour)},
					Added:     []string{"lifecycle.txt"},
				},
				{
					Timestamp: &github.Timestamp{Time: time.Now().Add(-2 * time.Hour)},
					Modified:  []string{"lifecycle.txt"},
				},
				{
					Timestamp: &github.Timestamp{Time: time.Now().Add(-1 * time.Hour)},
					Removed:   []string{"lifecycle.txt"},
				},
			},
			wantNewFiles: []string{},
			wantModified: []string{},
			wantDeleted:  []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotNewFiles, gotModified, gotDeleted := classifyFilesBySimulation(tt.commits)

			assert.ElementsMatch(t, tt.wantNewFiles, gotNewFiles, "new files do not match")
			assert.ElementsMatch(t, tt.wantModified, gotModified, "modified files do not match")
			assert.ElementsMatch(t, tt.wantDeleted, gotDeleted, "deleted files do not match")
		})
	}
}
