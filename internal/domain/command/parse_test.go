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
			want: NewChangeFile(NewModifyFile("test.txt", []Hunk{
				NewHunk("Hello", "Hi"),
			})),
			wantErr: false,
		},
		{
			name: "rename_file",
			xmlStr: `<rename_file>
				<old_path>old.txt</old_path>
				<new_path>new.txt</new_path>
				<hunk>
					<search>old content</search>
					<replace>new content</replace>
				</hunk>
			</rename_file>`,
			want: NewChangeFile(NewRenameFile("old.txt", "new.txt", []Hunk{
				NewHunk("old content", "new content"),
			})),
			wantErr: false,
		},
		{
			name: "rename_file_without_hunks",
			xmlStr: `<rename_file>
				<old_path>old.txt</old_path>
				<new_path>new.txt</new_path>
			</rename_file>`,
			want:    NewChangeFile(NewRenameFile("old.txt", "new.txt", nil)),
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
						RenameFile: func(gotRename RenameFile) {
							wantRename := wantChange.Unwrap().(RenameFile)
							assert.Equal(t, wantRename.OldPath, gotRename.OldPath)
							assert.Equal(t, wantRename.NewPath, gotRename.NewPath)
							assert.Equal(t, wantRename.Hunks, gotRename.Hunks)
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
