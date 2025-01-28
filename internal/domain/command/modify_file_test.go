package command

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestModifyFile_MarshalXML(t *testing.T) {
	tests := []struct {
		name     string
		input    ModifyFile
		expected string
	}{
		{
			name: "modify file with single hunk",
			input: NewModifyFile("test.txt", []ModifyHunk{
				NewModifyHunk("old text", "new text"),
			}),
			expected: `<modify_file><path>test.txt</path><hunk><search>old text</search><replace>new text</replace></hunk></modify_file>`,
		},
		{
			name: "modify file with multiple hunks",
			input: NewModifyFile("test.txt", []ModifyHunk{
				NewModifyHunk("first old", "first new"),
				NewModifyHunk("second old", "second new"),
			}),
			expected: `<modify_file><path>test.txt</path><hunk><search>first old</search><replace>first new</replace></hunk><hunk><search>second old</search><replace>second new</replace></hunk></modify_file>`,
		},
		{
			name:     "modify file with empty hunks",
			input:    NewModifyFile("test.txt", []ModifyHunk{}),
			expected: `<modify_file><path>test.txt</path></modify_file>`,
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

func TestModifyFile_UnmarshalXML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ModifyFile
		wantErr  error
	}{
		{
			name:  "modify file with single hunk",
			input: `<modify_file><path>test.txt</path><hunk><search>old text</search><replace>new text</replace></hunk></modify_file>`,
			expected: NewModifyFile("test.txt", []ModifyHunk{
				NewModifyHunk("old text", "new text"),
			}),
		},
		{
			name:  "modify file with multiple hunks",
			input: `<modify_file><path>test.txt</path><hunk><search>first old</search><replace>first new</replace></hunk><hunk><search>second old</search><replace>second new</replace></hunk></modify_file>`,
			expected: NewModifyFile("test.txt", []ModifyHunk{
				NewModifyHunk("first old", "first new"),
				NewModifyHunk("second old", "second new"),
			}),
		},
		{
			name:    "modify file with empty hunks",
			input:   `<modify_file><path>test.txt</path></modify_file>`,
			wantErr: ErrEmptyModifyHunks,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result ModifyFile
			err := xml.Unmarshal([]byte(tt.input), &result)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
