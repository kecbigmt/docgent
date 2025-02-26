package tooluse

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateFile_MarshalXML(t *testing.T) {
	tests := []struct {
		name     string
		input    CreateFile
		expected string
	}{
		{
			name:     "create file with single knowledge source",
			input:    NewCreateFile("test.txt", "hello world", []string{"https://slack.com/archives/C01234567/p123456789"}),
			expected: `<create_file><path>test.txt</path><content>hello world</content><source_uri>https://slack.com/archives/C01234567/p123456789</source_uri></create_file>`,
		},
		{
			name:     "create file with multiple knowledge sources",
			input:    NewCreateFile("test.txt", "hello world", []string{"https://slack.com/archives/C01234567/p123456789", "https://github.com/user/repo/pull/1"}),
			expected: `<create_file><path>test.txt</path><content>hello world</content><source_uri>https://slack.com/archives/C01234567/p123456789</source_uri><source_uri>https://github.com/user/repo/pull/1</source_uri></create_file>`,
		},
		{
			name:     "create file with special characters",
			input:    NewCreateFile("path/to/file.txt", "line1\nline2\n<special>&chars", []string{"https://slack.com/archives/C01234567/p123456789"}),
			expected: `<create_file><path>path/to/file.txt</path><content>line1&#xA;line2&#xA;&lt;special&gt;&amp;chars</content><source_uri>https://slack.com/archives/C01234567/p123456789</source_uri></create_file>`,
		},
		{
			name:     "create file with empty content",
			input:    NewCreateFile("empty.txt", "", []string{"https://slack.com/archives/C01234567/p123456789"}),
			expected: `<create_file><path>empty.txt</path><content></content><source_uri>https://slack.com/archives/C01234567/p123456789</source_uri></create_file>`,
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

func TestCreateFile_UnmarshalXML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected CreateFile
	}{
		{
			name:     "create file with single knowledge source",
			input:    `<create_file><path>test.txt</path><content>hello world</content><source_uri>https://slack.com/archives/C01234567/p123456789</source_uri></create_file>`,
			expected: NewCreateFile("test.txt", "hello world", []string{"https://slack.com/archives/C01234567/p123456789"}),
		},
		{
			name:     "create file with multiple knowledge sources",
			input:    `<create_file><path>test.txt</path><content>hello world</content><source_uri>https://slack.com/archives/C01234567/p123456789</source_uri><source_uri>https://github.com/user/repo/pull/1</source_uri></create_file>`,
			expected: NewCreateFile("test.txt", "hello world", []string{"https://slack.com/archives/C01234567/p123456789", "https://github.com/user/repo/pull/1"}),
		},
		{
			name:     "create file with special characters",
			input:    `<create_file><path>path/to/file.txt</path><content>line1&#xA;line2&#xA;&lt;special&gt;&amp;chars</content><source_uri>https://slack.com/archives/C01234567/p123456789</source_uri></create_file>`,
			expected: NewCreateFile("path/to/file.txt", "line1\nline2\n<special>&chars", []string{"https://slack.com/archives/C01234567/p123456789"}),
		},
		{
			name:     "create file with empty content",
			input:    `<create_file><path>empty.txt</path><content></content><source_uri>https://slack.com/archives/C01234567/p123456789</source_uri></create_file>`,
			expected: NewCreateFile("empty.txt", "", []string{"https://slack.com/archives/C01234567/p123456789"}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result CreateFile
			err := xml.Unmarshal([]byte(tt.input), &result)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
