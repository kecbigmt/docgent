package tooluse

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRenameFile_MarshalXML(t *testing.T) {
	tests := []struct {
		name     string
		input    RenameFile
		expected string
	}{
		{
			name: "with hunks",
			input: NewRenameFile(
				"old.txt",
				"new.txt",
				[]Hunk{
					NewHunk("old content", "new content"),
					NewHunk("another old", "another new"),
				},
			),
			expected: `<rename_file><old_path>old.txt</old_path><new_path>new.txt</new_path><hunk><search>old content</search><replace>new content</replace></hunk><hunk><search>another old</search><replace>another new</replace></hunk></rename_file>`,
		},
		{
			name: "without hunks",
			input: NewRenameFile(
				"old.txt",
				"new.txt",
				nil,
			),
			expected: `<rename_file><old_path>old.txt</old_path><new_path>new.txt</new_path></rename_file>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := xml.Marshal(tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, string(data))
		})
	}
}

func TestRenameFile_UnmarshalXML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected RenameFile
	}{
		{
			name:  "with hunks",
			input: `<rename_file><old_path>old.txt</old_path><new_path>new.txt</new_path><hunk><search>old content</search><replace>new content</replace></hunk><hunk><search>another old</search><replace>another new</replace></hunk></rename_file>`,
			expected: NewRenameFile(
				"old.txt",
				"new.txt",
				[]Hunk{
					NewHunk("old content", "new content"),
					NewHunk("another old", "another new"),
				},
			),
		},
		{
			name:  "without hunks",
			input: `<rename_file><old_path>old.txt</old_path><new_path>new.txt</new_path></rename_file>`,
			expected: NewRenameFile(
				"old.txt",
				"new.txt",
				[]Hunk{},
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result RenameFile
			err := xml.Unmarshal([]byte(tt.input), &result)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected.OldPath, result.OldPath)
			assert.Equal(t, tt.expected.NewPath, result.NewPath)
			if tt.expected.Hunks == nil {
				assert.Empty(t, result.Hunks)
			} else {
				assert.Equal(t, tt.expected.Hunks, result.Hunks)
			}
		})
	}
}
