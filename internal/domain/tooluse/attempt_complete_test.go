package tooluse

import (
	"encoding/xml"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAttemptComplete_MarshalXML(t *testing.T) {
	tests := []struct {
		name     string
		input    AttemptComplete
		expected string
	}{
		{
			name: "simple message without source",
			input: NewAttemptComplete(
				[]Message{
					NewMessage("Here is the answer:\n- Docgent is a agent that can help you with your documentation.\n- Docgent can create documents based on chat history."),
				},
				nil,
			),
			expected: `<attempt_complete><message>Here is the answer:&#xA;- Docgent is a agent that can help you with your documentation.&#xA;- Docgent can create documents based on chat history.</message></attempt_complete>`,
		},
		{
			name: "multiple messages with sources",
			input: NewAttemptComplete(
				[]Message{
					NewMessage("Here is the answer:\n"),
					NewMessageWithSourceID("- Docgent is a agent that can help you with your documentation", "1,2"),
					NewMessageWithSourceID("- Docgent can create documents based on chat history.", "2"),
				},
				[]Source{
					NewSource("1", "https://github.com/owner/repo/blob/a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0/docs/what-is-docgent.md", "What is Docgent?"),
					NewSource("2", "https://github.com/owner/repo/blob/a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0/docs/docgent-features.md", "Docgent Features"),
				},
			),
			expected: `<attempt_complete><message>Here is the answer:&#xA;</message><message source="1,2">- Docgent is a agent that can help you with your documentation</message><message source="2">- Docgent can create documents based on chat history.</message><source id="1" uri="https://github.com/owner/repo/blob/a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0/docs/what-is-docgent.md">What is Docgent?</source><source id="2" uri="https://github.com/owner/repo/blob/a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0/docs/docgent-features.md">Docgent Features</source></attempt_complete>`,
		},
		{
			name: "message with special characters",
			input: NewAttemptComplete(
				[]Message{
					NewMessage("Special characters: <, >, &, \", '\nAnd a new line"),
				},
				nil,
			),
			expected: `<attempt_complete><message>Special characters: &lt;, &gt;, &amp;, &#34;, &#39;&#xA;And a new line</message></attempt_complete>`,
		},
		{
			name: "empty message",
			input: NewAttemptComplete(
				[]Message{
					NewMessage(""),
				},
				nil,
			),
			expected: `<attempt_complete><message></message></attempt_complete>`,
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

func TestAttemptComplete_UnmarshalXML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected AttemptComplete
	}{
		{
			name:  "simple message without source",
			input: `<attempt_complete><message>Test message</message></attempt_complete>`,
			expected: NewAttemptComplete(
				[]Message{
					NewMessage("Test message"),
				},
				nil,
			),
		},
		{
			name: "multiple messages with sources",
			input: `<attempt_complete>
				<message>First message</message>
				<message source="1,2">Second message</message>
				<message source="2">Third message</message>
				<source id="1" uri="https://example.com/doc1">Source 1</source>
				<source id="2" uri="https://example.com/doc2">Source 2</source>
			</attempt_complete>`,
			expected: NewAttemptComplete(
				[]Message{
					NewMessage("First message"),
					NewMessageWithSourceID("Second message", "1,2"),
					NewMessageWithSourceID("Third message", "2"),
				},
				[]Source{
					NewSource("1", "https://example.com/doc1", "Source 1"),
					NewSource("2", "https://example.com/doc2", "Source 2"),
				},
			),
		},
		{
			name:  "message with special characters",
			input: `<attempt_complete><message>Special characters: &lt;, &gt;, &amp;, &#34;, &#39;&#xA;And a new line</message></attempt_complete>`,
			expected: NewAttemptComplete(
				[]Message{
					NewMessage("Special characters: <, >, &, \", '\nAnd a new line"),
				},
				nil,
			),
		},
		{
			name:  "empty message",
			input: `<attempt_complete><message></message></attempt_complete>`,
			expected: NewAttemptComplete(
				[]Message{
					NewMessage(""),
				},
				nil,
			),
		},
		{
			name: "example from usage documentation (simple)",
			input: `<attempt_complete>
<message>Here is the answer:
- Docgent is a agent that can help you with your documentation.
- Docgent can create documents based on chat history.</message>
</attempt_complete>`,
			expected: NewAttemptComplete(
				[]Message{
					NewMessage("Here is the answer:\n- Docgent is a agent that can help you with your documentation.\n- Docgent can create documents based on chat history."),
				},
				nil,
			),
		},
		{
			name: "example from usage documentation (with sources)",
			input: `<attempt_complete>
<message>Here is the answer:
</message>
<message source="1,2">- Docgent is a agent that can help you with your documentation</message>
<message source="2">- Docgent can create documents based on chat history.</message>
<source id="1" uri="https://github.com/owner/repo/blob/a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0/docs/what-is-docgent.md">What is Docgent?</source>
<source id="2" uri="https://github.com/owner/repo/blob/a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0/docs/docgent-features.md">Docgent Features</source>
</attempt_complete>`,
			expected: NewAttemptComplete(
				[]Message{
					NewMessage("Here is the answer:\n"),
					NewMessageWithSourceID("- Docgent is a agent that can help you with your documentation", "1,2"),
					NewMessageWithSourceID("- Docgent can create documents based on chat history.", "2"),
				},
				[]Source{
					NewSource("1", "https://github.com/owner/repo/blob/a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0/docs/what-is-docgent.md", "What is Docgent?"),
					NewSource("2", "https://github.com/owner/repo/blob/a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0/docs/docgent-features.md", "Docgent Features"),
				},
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var result AttemptComplete
			err := xml.Unmarshal([]byte(tt.input), &result)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected.Messages, result.Messages)
			assert.Equal(t, tt.expected.Sources, result.Sources)
		})
	}
}

func TestMessage_GetSourceIDs(t *testing.T) {
	tests := []struct {
		name     string
		message  Message
		expected []string
	}{
		{
			name:     "empty source ID",
			message:  NewMessage("Test message"),
			expected: nil,
		},
		{
			name:     "single source ID",
			message:  NewMessageWithSourceID("Test message", "1"),
			expected: []string{"1"},
		},
		{
			name:     "multiple source IDs",
			message:  NewMessageWithSourceID("Test message", "1,2,3"),
			expected: []string{"1", "2", "3"},
		},
		{
			name:     "using NewMessageWithSourceIDs",
			message:  NewMessageWithSourceIDs("Test message", []string{"1", "2", "3"}),
			expected: []string{"1", "2", "3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.message.GetSourceIDs()
			assert.Equal(t, tt.expected, result)
		})
	}
}
