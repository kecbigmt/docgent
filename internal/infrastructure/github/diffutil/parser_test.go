package diffutil

import (
	"docgent-backend/internal/domain"
	"reflect"
	"testing"
)

func TestParser(t *testing.T) {
	diff := `diff --git a/file1.txt b/file1.txt
index 1234567..89abcde 100644
--- a/file1.txt
+++ b/file1.txt
@@ -1,3 +1,3 @@
-Hello
+Hi
 World
diff --git a/file2.txt b/file2.txt
index 1234567..89abcde 100644
--- a/file2.txt
+++ b/file2.txt
@@ -1,3 +1,3 @@
-Foo
+Bar
 Baz
diff --git a/newfile.txt b/newfile.txt
new file mode 100644
index 0000000..e69de29
--- /dev/null
+++ b/newfile.txt
@@ -0,0 +1,3 @@
+This is a new file.
+It has a few lines of text.
+End of the file.`

	expected := []domain.Diff{
		{
			OldName: "file1.txt",
			NewName: "file1.txt",
			Body: `diff --git a/file1.txt b/file1.txt
index 1234567..89abcde 100644
--- a/file1.txt
+++ b/file1.txt
@@ -1,3 +1,3 @@
-Hello
+Hi
 World
`,
			IsNewFile: false,
		},
		{
			OldName: "file2.txt",
			NewName: "file2.txt",
			Body: `diff --git a/file2.txt b/file2.txt
index 1234567..89abcde 100644
--- a/file2.txt
+++ b/file2.txt
@@ -1,3 +1,3 @@
-Foo
+Bar
 Baz
`,
			IsNewFile: false,
		},
		{
			OldName: "",
			NewName: "newfile.txt",
			Body: `diff --git a/newfile.txt b/newfile.txt
new file mode 100644
index 0000000..e69de29
--- /dev/null
+++ b/newfile.txt
@@ -0,0 +1,3 @@
+This is a new file.
+It has a few lines of text.
+End of the file.
`,
			IsNewFile: true,
		},
	}

	parser := NewParser()
	result := parser.Execute(diff)
	if !reflect.DeepEqual(result, expected) {
		t.Errorf("expected %v, got %v", expected, result)
	}
}
