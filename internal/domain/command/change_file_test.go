package command

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
			)),
			expected: `<create_file><path>example.txt</path><content>Hello, World!</content></create_file>`,
		},
		{
			name: "modify file",
			change: NewChangeFile(NewModifyFile(
				"example.txt",
				[]ModifyHunk{
					NewModifyHunk("old text", "new text"),
					NewModifyHunk("another old", "another new"),
				},
			)),
			expected: `<modify_file><path>example.txt</path><hunk><search>old text</search><replace>new text</replace></hunk><hunk><search>another old</search><replace>another new</replace></hunk></modify_file>`,
		},
		{
			name: "replace file",
			change: NewChangeFile(NewReplaceFile(
				"old.txt",
				"new.txt",
				"updated content",
			)),
			expected: `<replace_file><old_path>old.txt</old_path><new_path>new.txt</new_path><new_content>updated content</new_content></replace_file>`,
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
