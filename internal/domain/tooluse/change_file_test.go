package tooluse

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChangeFile_MarshalXML(t *testing.T) {
	tests := []struct {
		name     string
		change   ChangeFile
		expected string
	}{
		{
			name: "create file",
			change: NewChangeFile(NewCreateFile(
				"example.txt",
				"Hello, World!",
				[]string{"https://slack.com/archives/C01234567/p123456789"},
			)),
			expected: `<create_file><path>example.txt</path><content>Hello, World!</content><knowledge_source_uri>https://slack.com/archives/C01234567/p123456789</knowledge_source_uri></create_file>`,
		},
		{
			name: "modify file",
			change: NewChangeFile(NewModifyFile(
				"example.txt",
				[]Hunk{
					NewHunk("old text", "new text"),
					NewHunk("another old", "another new"),
				},
			)),
			expected: `<modify_file><path>example.txt</path><hunk><search>old text</search><replace>new text</replace></hunk><hunk><search>another old</search><replace>another new</replace></hunk></modify_file>`,
		},
		{
			name: "rename file",
			change: NewChangeFile(NewRenameFile(
				"old.txt",
				"new.txt",
				[]Hunk{
					NewHunk("old content", "new content"),
					NewHunk("another old content", "another new content"),
				},
			)),
			expected: `<rename_file><old_path>old.txt</old_path><new_path>new.txt</new_path><hunk><search>old content</search><replace>new content</replace></hunk><hunk><search>another old content</search><replace>another new content</replace></hunk></rename_file>`,
		},
		{
			name: "rename file without hunks",
			change: NewChangeFile(NewRenameFile(
				"old.txt",
				"new.txt",
				nil,
			)),
			expected: `<rename_file><old_path>old.txt</old_path><new_path>new.txt</new_path></rename_file>`,
		},
		{
			name: "delete file",
			change: NewChangeFile(NewDeleteFile(
				"to_delete.txt",
			)),
			expected: `<delete_file><path>to_delete.txt</path></delete_file>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := xml.Marshal(tt.change)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, string(data))
		})
	}
}
