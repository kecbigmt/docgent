package command

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReplaceFile_MarshalXML(t *testing.T) {
	tests := []struct {
		name     string
		input    ReplaceFile
		expected string
	}{
		{
			name:     "basic replace file",
			input:    NewReplaceFile("old.txt", "new.txt", "hello world"),
			expected: `<replace_file><old_path>old.txt</old_path><new_path>new.txt</new_path><new_content>hello world</new_content></replace_file>`,
		},
		{
			name:     "replace file with special characters",
			input:    NewReplaceFile("path/to/old.txt", "path/to/new.txt", "line1\nline2\n<special>&chars"),
			expected: `<replace_file><old_path>path/to/old.txt</old_path><new_path>path/to/new.txt</new_path><new_content>line1&#xA;line2&#xA;&lt;special&gt;&amp;chars</new_content></replace_file>`,
		},
		{
			name:     "replace file with empty content",
			input:    NewReplaceFile("old.txt", "new.txt", ""),
			expected: `<replace_file><old_path>old.txt</old_path><new_path>new.txt</new_path><new_content></new_content></replace_file>`,
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

func TestReplaceFile_UnmarshalXML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected ReplaceFile
	}{
		{
			name:     "basic replace file",
			input:    `<replace_file><old_path>old.txt</old_path><new_path>new.txt</new_path><new_content>hello world</new_content></replace_file>`,
			expected: NewReplaceFile("old.txt", "new.txt", "hello world"),
		},
		{
			name:     "replace file with special characters",
			input:    `<replace_file><old_path>path/to/old.txt</old_path><new_path>path/to/new.txt</new_path><new_content>line1&#xA;line2&#xA;&lt;special&gt;&amp;chars</new_content></replace_file>`,
			expected: NewReplaceFile("path/to/old.txt", "path/to/new.txt", "line1\nline2\n<special>&chars"),
		},
		{
			name:     "replace file with empty content",
			input:    `<replace_file><old_path>old.txt</old_path><new_path>new.txt</new_path><new_content></new_content></replace_file>`,
			expected: NewReplaceFile("old.txt", "new.txt", ""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result ReplaceFile
			err := xml.Unmarshal([]byte(tt.input), &result)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
