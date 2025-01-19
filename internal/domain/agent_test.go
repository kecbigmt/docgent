package domain

import (
	"docgent-backend/internal/domain/autoagent"
	"reflect"
	"testing"
)

func TestParseResponseFromProposalRefineAgent(t *testing.T) {
	tests := []struct {
		name    string
		input   autoagent.Response
		want    ProposalRefineAgentResponse
		wantErr bool
	}{
		{
			name: "find_file tool",
			input: autoagent.Response{
				Type:     autoagent.ToolUseResponse,
				Message:  "Finding a file...",
				ToolType: "find_file",
				ToolParams: []autoagent.ToolParam{
					{
						Key:   "name",
						Value: "test.md",
					},
				},
			},
			want: ProposalRefineAgentResponse{
				Type:     autoagent.ToolUseResponse,
				Message:  "Finding a file...",
				ToolType: FindFileTool,
				ToolParams: FindFileToolParams{
					Name: "test.md",
				},
			},
			wantErr: false,
		},
		{
			name: "update_proposal_diffs tool",
			input: autoagent.Response{
				Type:     autoagent.ToolUseResponse,
				Message:  "Updating proposal diffs...",
				ToolType: "update_proposal_diffs",
				ToolParams: []autoagent.ToolParam{
					{
						Key: "diff",
						Value: `--- a/test.md
+++ b/test.md
@@ -1,3 +1,3 @@
-Old content
+New content
 Rest of content`,
					},
				},
			},
			want: ProposalRefineAgentResponse{
				Type:     autoagent.ToolUseResponse,
				Message:  "Updating proposal diffs...",
				ToolType: UpdateProposalDiffsTool,
				ToolParams: UpdateProposalDiffsToolParams{
					Diffs: []Diff{
						{
							OldName: "test.md",
							NewName: "test.md",
							Body: `@@ -1,3 +1,3 @@
-Old content
+New content
 Rest of content`,
							IsNewFile: false,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "update_proposal_diffs tool with new file",
			input: autoagent.Response{
				Type:     autoagent.ToolUseResponse,
				Message:  "Updating proposal diffs...",
				ToolType: "update_proposal_diffs",
				ToolParams: []autoagent.ToolParam{
					{
						Key: "diff",
						Value: `--- /dev/null
+++ b/new-file.md
@@ -0,0 +1,2 @@
+This is a new file
+With some content`,
					},
				},
			},
			want: ProposalRefineAgentResponse{
				Type:     autoagent.ToolUseResponse,
				Message:  "Updating proposal diffs...",
				ToolType: UpdateProposalDiffsTool,
				ToolParams: UpdateProposalDiffsToolParams{
					Diffs: []Diff{
						{
							OldName: "",
							NewName: "new-file.md",
							Body: `@@ -0,0 +1,2 @@
+This is a new file
+With some content`,
							IsNewFile: true,
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "update_proposal_text tool",
			input: autoagent.Response{
				Type:     autoagent.ToolUseResponse,
				Message:  "Updating proposal text...",
				ToolType: "update_proposal_text",
				ToolParams: []autoagent.ToolParam{
					{
						Key:   "title",
						Value: "New Title",
					},
					{
						Key:   "body",
						Value: "New Body",
					},
				},
			},
			want: ProposalRefineAgentResponse{
				Type:     autoagent.ToolUseResponse,
				Message:  "Updating proposal text...",
				ToolType: UpdateProposalTextTool,
				ToolParams: UpdateProposalTextToolParams{
					Title: "New Title",
					Body:  "New Body",
				},
			},
			wantErr: false,
		},
		{
			name: "find_file tool without name parameter",
			input: autoagent.Response{
				Type:       autoagent.ToolUseResponse,
				Message:    "Finding a file...",
				ToolType:   "find_file",
				ToolParams: []autoagent.ToolParam{},
			},
			wantErr: true,
		},
		{
			name: "unknown tool type",
			input: autoagent.Response{
				Type:       autoagent.ToolUseResponse,
				Message:    "Using unknown tool...",
				ToolType:   "unknown_tool",
				ToolParams: []autoagent.ToolParam{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseResponseFromProposalRefineAgent(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseResponseFromProposalRefineAgent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if got.Type != tt.want.Type {
				t.Errorf("ParseResponseFromProposalRefineAgent() Type = %v, want %v", got.Type, tt.want.Type)
			}
			if got.Message != tt.want.Message {
				t.Errorf("ParseResponseFromProposalRefineAgent() Message = %v, want %v", got.Message, tt.want.Message)
			}
			if got.ToolType != tt.want.ToolType {
				t.Errorf("ParseResponseFromProposalRefineAgent() ToolType = %v, want %v", got.ToolType, tt.want.ToolType)
			}
			if !reflect.DeepEqual(got.ToolParams, tt.want.ToolParams) {
				t.Errorf("ParseResponseFromProposalRefineAgent() ToolParams = %v, want %v", got.ToolParams, tt.want.ToolParams)
			}
		})
	}
}
