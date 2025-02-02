package changefile

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"docgent-backend/internal/domain/autoagent/tooluse"
)

func TestParseDiff(t *testing.T) {
	tests := []struct {
		name       string
		diff       string
		want       tooluse.ChangeFile
		wantErr    bool
		wantErrMsg string
	}{
		{
			name: "simple hunk",
			diff: `diff --git a/main.go b/main.go
index 1234567..89abcde 100644
--- a/main.go
+++ b/main.go
@@ -1,2 +1,4 @@
-import "fmt"
+import (
+  "fmt"
+  "os"
+)`,
			want: tooluse.NewChangeFile(tooluse.NewModifyFile("main.go", []tooluse.Hunk{
				tooluse.NewHunk(
					`import "fmt"`,
					`import (
  "fmt"
  "os"
)`,
				),
			})),
			wantErr: false,
		},
		{
			name: "multiple hunks",
			diff: `diff --git a/main.go b/main.go
index 1234567..89abcde 100644
--- a/main.go
+++ b/main.go
@@ -1,2 +1,4 @@
-import "fmt"
+import (
+  "fmt"
+  "os"
+)
@@ -9,4 +12,4 @@
 func greeting() {
-  fmt.Println("hi")
+  fmt.Println("hello")
 }`,
			want: tooluse.NewChangeFile(tooluse.NewModifyFile("main.go", []tooluse.Hunk{
				tooluse.NewHunk(
					`import "fmt"`,
					`import (
  "fmt"
  "os"
)`,
				),
				tooluse.NewHunk(
					`func greeting() {
  fmt.Println("hi")
}`,
					`func greeting() {
  fmt.Println("hello")
}`,
				),
			})),
			wantErr: false,
		},
		{
			name: "file creation",
			diff: `diff --git a/newfile.txt b/newfile.txt
new file mode 100644
index 0000000..e69de29
--- /dev/null
+++ b/newfile.txt
@@ -0,0 +1,5 @@
+package main
+
+func main() {
+  println("Hello, World!")
+}`,
			want: tooluse.NewChangeFile(tooluse.NewCreateFile("newfile.txt", `package main

func main() {
  println("Hello, World!")
}`)),
			wantErr: false,
		},
		{
			name: "file deletion",
			diff: `diff --git a/old.go b/old.go
deleted file mode 100644
index 1234567..0000000
--- a/old.go
+++ /dev/null
@@ -1,5 +0,0 @@
-package main
-
-func main() {
-  println("Goodbye!")
-}`,
			want:    tooluse.NewChangeFile(tooluse.NewDeleteFile("old.go")),
			wantErr: false,
		},
		{
			name: "rename file without changes",
			diff: `diff --git a/old.go b/new.go
similarity index 100%
rename from old.go
rename to new.go`,
			want:    tooluse.NewChangeFile(tooluse.NewRenameFile("old.go", "new.go", []tooluse.Hunk{})),
			wantErr: false,
		},
		{
			name: "rename file with changes",
			diff: `diff --git a/old.go b/new.go
similarity index 50%
rename from old.go
rename to new.go
index 1234567..89abcde 100644
--- a/old.go
+++ b/new.go
@@ -1,3 +1,3 @@
 package main
 
-func oldName() {}
+func newName() {}`,
			want: tooluse.NewChangeFile(tooluse.NewRenameFile("old.go", "new.go", []tooluse.Hunk{
				tooluse.NewHunk(
					"package main\n\nfunc oldName() {}",
					"package main\n\nfunc newName() {}",
				),
			})),
			wantErr: false,
		},
		{
			name:       "invalid diff without git header",
			diff:       "+++ b/main.go\n@@ -1,1 +1,1 @@\n-test\n+test2\n",
			want:       tooluse.ChangeFile{},
			wantErr:    true,
			wantErrMsg: "failed to find file paths in diff",
		},
		{
			name: "invalid git diff header format",
			diff: `diff --git invalid header format
@@ -1,1 +1,1 @@
-test
+test2`,
			want:       tooluse.ChangeFile{},
			wantErr:    true,
			wantErrMsg: "invalid git diff header format: diff --git invalid header format",
		},
		{
			name: "invalid file creation without content",
			diff: `diff --git a/newfile.txt b/newfile.txt
new file mode 100644
index 0000000..e69de29
--- /dev/null
+++ b/newfile.txt
@@ -0,0 +0,0 @@`,
			want:       tooluse.ChangeFile{},
			wantErr:    true,
			wantErrMsg: "no content found in new file diff",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseDiff(tt.diff)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrMsg != "" {
					assert.Equal(t, tt.wantErrMsg, err.Error())
				}
				return
			}
			assert.NoError(t, err)

			// Matchメソッドを使って検証
			got.Unwrap().Match(tooluse.FileChangeCases{
				ModifyFile: func(gotModify tooluse.ModifyFile) (string, bool, error) {
					wantModify := tt.want.Unwrap().(tooluse.ModifyFile)
					assert.Equal(t, wantModify.Path, gotModify.Path)
					assert.Equal(t, len(wantModify.Hunks), len(gotModify.Hunks))
					for i := range wantModify.Hunks {
						assert.Equal(t, wantModify.Hunks[i].Search, gotModify.Hunks[i].Search)
						assert.Equal(t, wantModify.Hunks[i].Replace, gotModify.Hunks[i].Replace)
					}
					return "File modified", false, nil
				},
				CreateFile: func(gotCreate tooluse.CreateFile) (string, bool, error) {
					wantCreate := tt.want.Unwrap().(tooluse.CreateFile)
					assert.Equal(t, wantCreate.Path, gotCreate.Path)
					assert.Equal(t, wantCreate.Content, gotCreate.Content)
					return "File created", false, nil
				},
				DeleteFile: func(gotDelete tooluse.DeleteFile) (string, bool, error) {
					wantDelete := tt.want.Unwrap().(tooluse.DeleteFile)
					assert.Equal(t, wantDelete.Path, gotDelete.Path)
					return "File deleted", false, nil
				},
				RenameFile: func(gotRename tooluse.RenameFile) (string, bool, error) {
					wantRename := tt.want.Unwrap().(tooluse.RenameFile)
					assert.Equal(t, wantRename.OldPath, gotRename.OldPath)
					assert.Equal(t, wantRename.NewPath, gotRename.NewPath)
					assert.Equal(t, len(wantRename.Hunks), len(gotRename.Hunks))
					for i := range wantRename.Hunks {
						assert.Equal(t, wantRename.Hunks[i].Search, gotRename.Hunks[i].Search)
						assert.Equal(t, wantRename.Hunks[i].Replace, gotRename.Hunks[i].Replace)
					}
					return "File renamed", false, nil
				},
			})
		})
	}
}
