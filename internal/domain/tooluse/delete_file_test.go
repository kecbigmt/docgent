package tooluse

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDeleteFile_MarshalXML(t *testing.T) {
	tests := []struct {
		name     string
		input    DeleteFile
		expected string
	}{
		{
			name:     "basic delete file",
			input:    NewDeleteFile("test.txt"),
			expected: `<delete_file><path>test.txt</path></delete_file>`,
		},
		{
			name:     "delete file with path containing special characters",
			input:    NewDeleteFile("path/to/file with spaces & symbols.txt"),
			expected: `<delete_file><path>path/to/file with spaces &amp; symbols.txt</path></delete_file>`,
		},
		{
			name:     "delete file with empty path",
			input:    NewDeleteFile(""),
			expected: `<delete_file><path></path></delete_file>`,
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

func TestDeleteFile_UnmarshalXML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected DeleteFile
	}{
		{
			name:     "basic delete file",
			input:    `<delete_file><path>test.txt</path></delete_file>`,
			expected: NewDeleteFile("test.txt"),
		},
		{
			name:     "delete file with path containing special characters",
			input:    `<delete_file><path>path/to/file with spaces &amp; symbols.txt</path></delete_file>`,
			expected: NewDeleteFile("path/to/file with spaces & symbols.txt"),
		},
		{
			name:     "delete file with empty path",
			input:    `<delete_file><path></path></delete_file>`,
			expected: NewDeleteFile(""),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result DeleteFile
			err := xml.Unmarshal([]byte(tt.input), &result)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
