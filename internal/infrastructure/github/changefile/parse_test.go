package changefile

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"docgent-backend/internal/domain/command"
)

func TestParseDiff(t *testing.T) {
	tests := []struct {
		name       string
		diff       string
		want       command.ChangeFile
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
			want: command.NewChangeFile(command.NewModifyFile("main.go", []command.Hunk{
				command.NewHunk(
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
			want: command.NewChangeFile(command.NewModifyFile("main.go", []command.Hunk{
				command.NewHunk(
					`import "fmt"`,
					`import (
  "fmt"
  "os"
)`,
				),
				command.NewHunk(
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
			want: command.NewChangeFile(command.NewCreateFile("newfile.txt", `package main

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
			want:    command.NewChangeFile(command.NewDeleteFile("old.go")),
			wantErr: false,
		},
		{
			name: "rename file without changes",
			diff: `diff --git a/old.go b/new.go
similarity index 100%
rename from old.go
rename to new.go`,
			want:    command.NewChangeFile(command.NewRenameFile("old.go", "new.go", []command.Hunk{})),
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
			want: command.NewChangeFile(command.NewRenameFile("old.go", "new.go", []command.Hunk{
				command.NewHunk(
					"package main\n\nfunc oldName() {}",
					"package main\n\nfunc newName() {}",
				),
			})),
			wantErr: false,
		},
		{
			name:       "invalid diff without git header",
			diff:       "+++ b/main.go\n@@ -1,1 +1,1 @@\n-test\n+test2\n",
			want:       command.ChangeFile{},
			wantErr:    true,
			wantErrMsg: "failed to find file paths in diff",
		},
		{
			name: "invalid git diff header format",
			diff: `diff --git invalid header format
@@ -1,1 +1,1 @@
-test
+test2`,
			want:       command.ChangeFile{},
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
			want:       command.ChangeFile{},
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
			got.Unwrap().Match(command.FileChangeCases{
				ModifyFile: func(gotModify command.ModifyFile) {
					wantModify := tt.want.Unwrap().(command.ModifyFile)
					assert.Equal(t, wantModify.Path, gotModify.Path)
					assert.Equal(t, len(wantModify.Hunks), len(gotModify.Hunks))
					for i := range wantModify.Hunks {
						assert.Equal(t, wantModify.Hunks[i].Search, gotModify.Hunks[i].Search)
						assert.Equal(t, wantModify.Hunks[i].Replace, gotModify.Hunks[i].Replace)
					}
				},
				CreateFile: func(gotCreate command.CreateFile) {
					wantCreate := tt.want.Unwrap().(command.CreateFile)
					assert.Equal(t, wantCreate.Path, gotCreate.Path)
					assert.Equal(t, wantCreate.Content, gotCreate.Content)
				},
				DeleteFile: func(gotDelete command.DeleteFile) {
					wantDelete := tt.want.Unwrap().(command.DeleteFile)
					assert.Equal(t, wantDelete.Path, gotDelete.Path)
				},
				RenameFile: func(gotRename command.RenameFile) {
					wantRename := tt.want.Unwrap().(command.RenameFile)
					assert.Equal(t, wantRename.OldPath, gotRename.OldPath)
					assert.Equal(t, wantRename.NewPath, gotRename.NewPath)
					assert.Equal(t, len(wantRename.Hunks), len(gotRename.Hunks))
					for i := range wantRename.Hunks {
						assert.Equal(t, wantRename.Hunks[i].Search, gotRename.Hunks[i].Search)
						assert.Equal(t, wantRename.Hunks[i].Replace, gotRename.Hunks[i].Replace)
					}
				},
			})
		})
	}
}
