package tooluse

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
			input: NewModifyFile("test.txt", []Hunk{
				NewHunk("old text", "new text"),
			}),
			expected: `<modify_file><path>test.txt</path><hunk><search>old text</search><replace>new text</replace></hunk></modify_file>`,
		},
		{
			name: "modify file with multiple hunks",
			input: NewModifyFile("test.txt", []Hunk{
				NewHunk("first old", "first new"),
				NewHunk("second old", "second new"),
			}),
			expected: `<modify_file><path>test.txt</path><hunk><search>first old</search><replace>first new</replace></hunk><hunk><search>second old</search><replace>second new</replace></hunk></modify_file>`,
		},
		{
			name:     "modify file with empty hunks",
			input:    NewModifyFile("test.txt", []Hunk{}),
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
			expected: NewModifyFile("test.txt", []Hunk{
				NewHunk("old text", "new text"),
			}),
		},
		{
			name:  "modify file with multiple hunks",
			input: `<modify_file><path>test.txt</path><hunk><search>first old</search><replace>first new</replace></hunk><hunk><search>second old</search><replace>second new</replace></hunk></modify_file>`,
			expected: NewModifyFile("test.txt", []Hunk{
				NewHunk("first old", "first new"),
				NewHunk("second old", "second new"),
			}),
		},
		{
			name:    "modify file with empty hunks",
			input:   `<modify_file><path>test.txt</path></modify_file>`,
			wantErr: ErrEmptyHunks,
		},
		{
			name: "modify file with newlines in hunks",
			input: `<modify_file><path>test.txt</path><hunk><search>
				Hello,
				world!
			</search><replace>
				Hi,
				world!
			</replace></hunk></modify_file>`,
			expected: NewModifyFile("test.txt", []Hunk{
				NewHunk("Hello,\nworld!", "Hi,\nworld!"),
			}),
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
