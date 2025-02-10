package changefile

import (
	"bufio"
	"fmt"
	"strings"

	"docgent/internal/domain/tooluse"
)

type fileOperation int

const (
	opUnknown fileOperation = iota
	opCreate                // 新規作成
	opDelete                // 削除
	opRename                // リネーム
	opModify                // 変更
)

type diffHeader struct {
	oldPath   string
	newPath   string
	operation fileOperation
}

// ParseDiff parses a Git diff format string and returns a ChangeFile
func ParseDiff(gitDiff string) (tooluse.ChangeFile, error) {
	scanner := bufio.NewScanner(strings.NewReader(gitDiff))

	header, err := parseHeader(scanner)
	if err != nil {
		return tooluse.ChangeFile{}, err
	}

	switch header.operation {
	case opCreate:
		return handleCreateFile(scanner, header)
	case opDelete:
		return handleDeleteFile(header)
	case opRename:
		return handleRenameFile(scanner, header)
	case opModify:
		return handleModifyFile(scanner, header)
	default:
		return tooluse.ChangeFile{}, fmt.Errorf("unknown file operation")
	}
}

// parseHeader はGit diffのヘッダー部分を解析し、diffHeader構造体を返します
func parseHeader(scanner *bufio.Scanner) (diffHeader, error) {
	header := diffHeader{operation: opUnknown}

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "diff --git") {
			parts := strings.Split(line, " ")
			if len(parts) != 4 {
				return diffHeader{}, fmt.Errorf("invalid git diff header format: %s", line)
			}
			header.oldPath = strings.TrimPrefix(parts[2], "a/")
			header.newPath = strings.TrimPrefix(parts[3], "b/")
			continue
		}

		// ファイルモードの変更を確認
		if strings.HasPrefix(line, "new file mode") {
			header.operation = opCreate
			header.oldPath = "/dev/null"
			continue
		}
		if strings.HasPrefix(line, "deleted file mode") {
			header.operation = opDelete
			header.newPath = "/dev/null"
			continue
		}
		if strings.HasPrefix(line, "rename from") {
			header.operation = opRename
			header.oldPath = strings.TrimPrefix(strings.TrimSpace(line), "rename from ")
			continue
		}
		if strings.HasPrefix(line, "rename to") {
			header.operation = opRename
			header.newPath = strings.TrimPrefix(strings.TrimSpace(line), "rename to ")
			continue
		}

		// パス情報の確認
		if strings.HasPrefix(line, "---") {
			if strings.TrimSpace(line) == "--- /dev/null" {
				header.oldPath = "/dev/null"
			}
			continue
		}
		if strings.HasPrefix(line, "+++") {
			if strings.TrimSpace(line) == "+++ /dev/null" {
				header.newPath = "/dev/null"
			}
			break
		}
	}

	if header.oldPath == "" || header.newPath == "" {
		return diffHeader{}, fmt.Errorf("failed to find file paths in diff")
	}

	// 操作タイプが未設定の場合は、パス情報から推論
	if header.operation == opUnknown {
		header.operation = determineOperation(header)
	}

	return header, nil
}

// determineOperation はパス情報から操作タイプを判定します
func determineOperation(header diffHeader) fileOperation {
	switch {
	case header.oldPath == "/dev/null" && header.newPath != "/dev/null":
		return opCreate
	case header.newPath == "/dev/null" && header.oldPath != "/dev/null":
		return opDelete
	case header.oldPath != header.newPath:
		return opRename
	default:
		return opModify
	}
}

// handleCreateFile は新規ファイル作成の処理を行います
func handleCreateFile(scanner *bufio.Scanner, header diffHeader) (tooluse.ChangeFile, error) {
	content, err := readNewContent(scanner)
	if err != nil {
		return tooluse.ChangeFile{}, err
	}
	return tooluse.NewChangeFile(tooluse.NewCreateFile(header.newPath, content)), nil
}

// handleDeleteFile はファイル削除の処理を行います
func handleDeleteFile(header diffHeader) (tooluse.ChangeFile, error) {
	return tooluse.NewChangeFile(tooluse.NewDeleteFile(header.oldPath)), nil
}

// handleRenameFile はファイルリネームの処理を行います
func handleRenameFile(scanner *bufio.Scanner, header diffHeader) (tooluse.ChangeFile, error) {
	hunks, err := readHunks(scanner)
	if err != nil {
		return tooluse.ChangeFile{}, err
	}
	return tooluse.NewChangeFile(tooluse.NewRenameFile(header.oldPath, header.newPath, hunks)), nil
}

// handleModifyFile はファイル変更の処理を行います
func handleModifyFile(scanner *bufio.Scanner, header diffHeader) (tooluse.ChangeFile, error) {
	hunks, err := readHunks(scanner)
	if err != nil {
		return tooluse.ChangeFile{}, err
	}
	return tooluse.NewChangeFile(tooluse.NewModifyFile(header.newPath, hunks)), nil
}

// readHunks reads all hunks from the diff
func readHunks(scanner *bufio.Scanner) ([]tooluse.Hunk, error) {
	var hunks []tooluse.Hunk
	var searchLines, replaceLines []string
	inHunk := false

	for scanner.Scan() {
		line := scanner.Text()

		// ハンクヘッダーをスキップ
		if strings.HasPrefix(line, "@@") {
			if inHunk {
				// 前のハンクを保存
				if len(searchLines) > 0 || len(replaceLines) > 0 {
					hunks = append(hunks, tooluse.NewHunk(
						strings.Join(searchLines, "\n"),
						strings.Join(replaceLines, "\n"),
					))
				}
			}
			inHunk = true
			searchLines = []string{}
			replaceLines = []string{}
			continue
		}

		if !inHunk {
			continue
		}

		switch {
		case strings.HasPrefix(line, "-"):
			searchLines = append(searchLines, strings.TrimPrefix(line, "-"))
		case strings.HasPrefix(line, "+"):
			replaceLines = append(replaceLines, strings.TrimPrefix(line, "+"))
		case strings.HasPrefix(line, " "):
			// コンテキスト行は両方に追加
			trimmed := strings.TrimPrefix(line, " ")
			searchLines = append(searchLines, trimmed)
			replaceLines = append(replaceLines, trimmed)
		}
	}

	// 最後のハンクを保存
	if len(searchLines) > 0 || len(replaceLines) > 0 {
		hunks = append(hunks, tooluse.NewHunk(
			strings.Join(searchLines, "\n"),
			strings.Join(replaceLines, "\n"),
		))
	}

	return hunks, nil
}

// readNewContent reads the content of a new file from the diff
func readNewContent(scanner *bufio.Scanner) (string, error) {
	var lines []string
	inHunk := false

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "@@") {
			inHunk = true
			continue
		}

		if !inHunk {
			continue
		}

		if strings.HasPrefix(line, "+") {
			lines = append(lines, strings.TrimPrefix(line, "+"))
		} else if !strings.HasPrefix(line, "-") {
			lines = append(lines, strings.TrimPrefix(line, " "))
		}
	}

	if len(lines) == 0 {
		return "", fmt.Errorf("no content found in new file diff")
	}

	return strings.Join(lines, "\n"), nil
}
