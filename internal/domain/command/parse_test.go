package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		xmlStr  string
		want    CommandUnion
		wantErr bool
	}{
		{
			name: "create_file",
			xmlStr: `<create_file>
				<path>test.txt</path>
				<content>Hello, World!</content>
			</create_file>`,
			want:    NewChangeFile(NewCreateFile("test.txt", "Hello, World!")),
			wantErr: false,
		},
		{
			name: "modify_file",
			xmlStr: `<modify_file>
				<path>test.txt</path>
				<hunk>
					<search>Hello</search>
					<replace>Hi</replace>
				</hunk>
			</modify_file>`,
			want: NewChangeFile(NewModifyFile("test.txt", []ModifyHunk{
				NewModifyHunk("Hello", "Hi"),
			})),
			wantErr: false,
		},
		{
			name: "replace_file",
			xmlStr: `<replace_file>
				<old_path>old.txt</old_path>
				<new_path>new.txt</new_path>
				<new_content>New content</new_content>
			</replace_file>`,
			want:    NewChangeFile(NewReplaceFile("old.txt", "new.txt", "New content")),
			wantErr: false,
		},
		{
			name: "read_file",
			xmlStr: `<read_file>
				<path>test.txt</path>
			</read_file>`,
			want: ReadFile{
				Path: "test.txt",
			},
			wantErr: false,
		},
		{
			name: "invalid_command",
			xmlStr: `<unknown_command>
				<path>test.txt</path>
			</unknown_command>`,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.xmlStr)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, got)

			// Matchメソッドを使って検証
			got.Match(Cases{
				ChangeFile: func(gotChange ChangeFile) {
					wantChange := tt.want.(ChangeFile)
					gotChange.Unwrap().Match(FileChangeCases{
						CreateFile: func(gotCreate CreateFile) {
							wantCreate := wantChange.Unwrap().(CreateFile)
							assert.Equal(t, wantCreate.Path, gotCreate.Path)
							assert.Equal(t, wantCreate.Content, gotCreate.Content)
						},
						ModifyFile: func(gotModify ModifyFile) {
							wantModify := wantChange.Unwrap().(ModifyFile)
							assert.Equal(t, wantModify.Path, gotModify.Path)
							assert.Equal(t, wantModify.Hunks, gotModify.Hunks)
						},
						ReplaceFile: func(gotReplace ReplaceFile) {
							wantReplace := wantChange.Unwrap().(ReplaceFile)
							assert.Equal(t, wantReplace.OldPath, gotReplace.OldPath)
							assert.Equal(t, wantReplace.NewPath, gotReplace.NewPath)
							assert.Equal(t, wantReplace.NewContent, gotReplace.NewContent)
						},
						DeleteFile: func(DeleteFile) {
							t.Error("unexpected DeleteFile")
						},
					})
				},
				ReadFile: func(gotRead ReadFile) {
					wantRead := tt.want.(ReadFile)
					assert.Equal(t, wantRead.Path, gotRead.Path)
				},
			})
		})
	}
}
