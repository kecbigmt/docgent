package tooluse

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		xmlStr  string
		want    Union
		wantErr bool
	}{
		{
			name: "create_file",
			xmlStr: `<create_file>
				<path>test.txt</path>
				<content>Hello, World!</content>
				<source_uri>https://slack.com/archives/C01234567/p123456789</source_uri>
			</create_file>`,
			want:    NewChangeFile(NewCreateFile("test.txt", "Hello, World!", []string{"https://slack.com/archives/C01234567/p123456789"})),
			wantErr: false,
		},
		{
			name: "link_sources",
			xmlStr: `<link_sources>
				<file_path>test.txt</file_path>
				<uri>https://slack.com/archives/C01234567/p123456789</uri>
				<uri>https://github.com/user/repo/pull/1</uri>
			</link_sources>`,
			want:    NewLinkSources("test.txt", []string{"https://slack.com/archives/C01234567/p123456789", "https://github.com/user/repo/pull/1"}),
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
			name: "find_file",
			xmlStr: `<find_file>
				<path>test.txt</path>
			</find_file>`,
			want: FindFile{
				Path: "test.txt",
			},
			wantErr: false,
		},
		{
			name: "create_proposal",
			xmlStr: `<create_proposal>
				<title>Test Proposal</title>
				<description>This is a test proposal</description>
			</create_proposal>`,
			want:    NewCreateProposal("Test Proposal", "This is a test proposal"),
			wantErr: false,
		},
		{
			name: "update_proposal",
			xmlStr: `<update_proposal>
				<title>Updated Proposal</title>
				<description>This is an updated proposal</description>
			</update_proposal>`,
			want:    NewUpdateProposal("Updated Proposal", "This is an updated proposal"),
			wantErr: false,
		},
		{
			name: "find_source",
			xmlStr: `<find_source>
				<uri>https://slack.com/archives/C01234567/p123456789</uri>
			</find_source>`,
			want:    NewFindSource("https://slack.com/archives/C01234567/p123456789"),
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
				ChangeFile: func(gotChange ChangeFile) (string, bool, error) {
					wantChange := tt.want.(ChangeFile)
					return gotChange.Unwrap().Match(ChangeFileCases{
						CreateFile: func(gotCreate CreateFile) (string, bool, error) {
							wantCreate := wantChange.Unwrap().(CreateFile)
							assert.Equal(t, wantCreate.Path, gotCreate.Path)
							assert.Equal(t, wantCreate.Content, gotCreate.Content)
							assert.Equal(t, wantCreate.SourceURIs, gotCreate.SourceURIs)
							return "file created", false, nil
						},
						ModifyFile: func(gotModify ModifyFile) (string, bool, error) {
							wantModify := wantChange.Unwrap().(ModifyFile)
							assert.Equal(t, wantModify.Path, gotModify.Path)
							assert.Equal(t, wantModify.Hunks, gotModify.Hunks)
							return "file modified", false, nil
						},
						RenameFile: func(gotRename RenameFile) (string, bool, error) {
							wantRename := wantChange.Unwrap().(RenameFile)
							assert.Equal(t, wantRename.OldPath, gotRename.OldPath)
							assert.Equal(t, wantRename.NewPath, gotRename.NewPath)
							assert.Equal(t, wantRename.Hunks, gotRename.Hunks)
							return "file renamed", false, nil
						},
						DeleteFile: func(gotDelete DeleteFile) (string, bool, error) {
							wantDelete := wantChange.Unwrap().(DeleteFile)
							assert.Equal(t, wantDelete.Path, gotDelete.Path)
							return "file deleted", false, nil
						},
					})
				},
				FindFile: func(gotRead FindFile) (string, bool, error) {
					wantRead := tt.want.(FindFile)
					assert.Equal(t, wantRead.Path, gotRead.Path)
					return "file read", false, nil
				},
				CreateProposal: func(gotCreate CreateProposal) (string, bool, error) {
					wantCreate := tt.want.(CreateProposal)
					assert.Equal(t, wantCreate.Title, gotCreate.Title)
					assert.Equal(t, wantCreate.Description, gotCreate.Description)
					return "proposal created", false, nil
				},
				UpdateProposal: func(gotUpdate UpdateProposal) (string, bool, error) {
					wantUpdate := tt.want.(UpdateProposal)
					assert.Equal(t, wantUpdate.Title, gotUpdate.Title)
					assert.Equal(t, wantUpdate.Description, gotUpdate.Description)
					return "proposal updated", false, nil
				},
				LinkSources: func(gotLink LinkSources) (string, bool, error) {
					wantLink := tt.want.(LinkSources)
					assert.Equal(t, wantLink.FilePath, gotLink.FilePath)
					assert.Equal(t, wantLink.URIs, gotLink.URIs)
					return "knowledge sources added", false, nil
				},
				FindSource: func(gotFindSource FindSource) (string, bool, error) {
					wantFindSource := tt.want.(FindSource)
					assert.Equal(t, wantFindSource.URI, gotFindSource.URI)
					return "knowledge source found", false, nil
				},
			})
		})
	}
}
