package autoagent

import (
	"testing"
)

func TestParseResponse(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    Response
		wantErr bool
	}{
		{
			name: "complete response",
			raw:  "<complete>Task completed successfully</complete>",
			want: Response{
				Type:    CompleteResponse,
				Message: "Task completed successfully",
			},
			wantErr: false,
		},
		{
			name: "error response",
			raw:  "<error>Something went wrong</error>",
			want: Response{
				Type:    ErrorResponse,
				Message: "Something went wrong",
			},
			wantErr: false,
		},
		{
			name: "tool use response",
			raw: `<tool_use:read_file>
<message>Reading file content</message>
<param:path>test.txt</param:path>
</tool_use:read_file>`,
			want: Response{
				Type:     ToolUseResponse,
				Message:  "Reading file content",
				ToolType: "read_file",
				ToolParams: []ToolParam{
					{
						Key:   "path",
						Value: "test.txt",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "tool use response with multiple params",
			raw: `<tool_use:write_file>
<message>Writing content to file</message>
<param:path>test.txt</param:path>
<param:content>Hello, World!</param:content>
</tool_use:write_file>`,
			want: Response{
				Type:     ToolUseResponse,
				Message:  "Writing content to file",
				ToolType: "write_file",
				ToolParams: []ToolParam{
					{
						Key:   "path",
						Value: "test.txt",
					},
					{
						Key:   "content",
						Value: "Hello, World!",
					},
				},
			},
			wantErr: false,
		},
		{
			name:    "invalid format",
			raw:     "invalid format",
			wantErr: true,
		},
		{
			name:    "empty string",
			raw:     "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseResponse(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if got.Type != tt.want.Type {
				t.Errorf("ParseResponse() Type = %v, want %v", got.Type, tt.want.Type)
			}
			if got.Message != tt.want.Message {
				t.Errorf("ParseResponse() Message = %v, want %v", got.Message, tt.want.Message)
			}
			if got.ToolType != tt.want.ToolType {
				t.Errorf("ParseResponse() ToolType = %v, want %v", got.ToolType, tt.want.ToolType)
			}
			if len(got.ToolParams) != len(tt.want.ToolParams) {
				t.Errorf("ParseResponse() ToolParams length = %v, want %v", len(got.ToolParams), len(tt.want.ToolParams))
				return
			}
			for i := range got.ToolParams {
				if got.ToolParams[i].Key != tt.want.ToolParams[i].Key {
					t.Errorf("ParseResponse() ToolParams[%d].Key = %v, want %v", i, got.ToolParams[i].Key, tt.want.ToolParams[i].Key)
				}
				if got.ToolParams[i].Value != tt.want.ToolParams[i].Value {
					t.Errorf("ParseResponse() ToolParams[%d].Value = %v, want %v", i, got.ToolParams[i].Value, tt.want.ToolParams[i].Value)
				}
			}
		})
	}
}

// ToolType is a helper type that implements fmt.Stringer
type ToolType string

func (t ToolType) String() string {
	return string(t)
}
