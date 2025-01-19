package autoagent

import (
	"encoding/json"
	"testing"
)

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    Response
		wantErr bool
	}{
		{
			name: "complete response",
			raw:  `{"type":"complete","message":"Task completed successfully"}`,
			want: Response{
				Type:    CompleteResponse,
				Message: "Task completed successfully",
			},
			wantErr: false,
		},
		{
			name: "error response",
			raw:  `{"type":"error","message":"Something went wrong"}`,
			want: Response{
				Type:    ErrorResponse,
				Message: "Something went wrong",
			},
			wantErr: false,
		},
		{
			name: "tool use response",
			raw:  `{"type":"tool_use","message":"Reading file content","toolType":"read_file","toolParams":[{"k":"path","v":"test.txt"}]}`,
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
			raw:  `{"type":"tool_use","message":"Writing content to file","toolType":"write_file","toolParams":[{"k":"path","v":"test.txt"},{"k":"content","v":"Hello, World!"}]}`,
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
			var got Response
			err := json.Unmarshal([]byte(tt.raw), &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("json.Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err != nil {
				return
			}

			if got.Type != tt.want.Type {
				t.Errorf("Response.Type = %v, want %v", got.Type, tt.want.Type)
			}
			if got.Message != tt.want.Message {
				t.Errorf("Response.Message = %v, want %v", got.Message, tt.want.Message)
			}
			if got.ToolType != tt.want.ToolType {
				t.Errorf("Response.ToolType = %v, want %v", got.ToolType, tt.want.ToolType)
			}
			if len(got.ToolParams) != len(tt.want.ToolParams) {
				t.Errorf("Response.ToolParams length = %v, want %v", len(got.ToolParams), len(tt.want.ToolParams))
				return
			}
			for i := range got.ToolParams {
				if got.ToolParams[i].Key != tt.want.ToolParams[i].Key {
					t.Errorf("Response.ToolParams[%d].Key = %v, want %v", i, got.ToolParams[i].Key, tt.want.ToolParams[i].Key)
				}
				if got.ToolParams[i].Value != tt.want.ToolParams[i].Value {
					t.Errorf("Response.ToolParams[%d].Value = %v, want %v", i, got.ToolParams[i].Value, tt.want.ToolParams[i].Value)
				}
			}
		})
	}
}

func TestMarshal(t *testing.T) {
	tests := []struct {
		name string
		resp Response
		want string
	}{
		{
			name: "complete response",
			resp: Response{
				Type:    CompleteResponse,
				Message: "Task completed successfully",
			},
			want: `{"type":"complete","message":"Task completed successfully"}`,
		},
		{
			name: "error response",
			resp: Response{
				Type:    ErrorResponse,
				Message: "Something went wrong",
			},
			want: `{"type":"error","message":"Something went wrong"}`,
		},
		{
			name: "tool use response with params",
			resp: Response{
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
			want: `{"type":"tool_use","message":"Writing content to file","toolType":"write_file","toolParams":[{"k":"path","v":"test.txt"},{"k":"content","v":"Hello, World!"}]}`,
		},
		{
			name: "tool use response without params",
			resp: Response{
				Type:     ToolUseResponse,
				Message:  "Reading file content",
				ToolType: "read_file",
			},
			want: `{"type":"tool_use","message":"Reading file content","toolType":"read_file"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := json.Marshal(tt.resp)
			if err != nil {
				t.Errorf("json.Marshal() error = %v", err)
				return
			}
			if string(got) != tt.want {
				t.Errorf("Response.String() = %v, want %v", string(got), tt.want)
			}
		})
	}
}
