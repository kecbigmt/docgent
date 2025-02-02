package diffutil

import (
	"strconv"
	"strings"

	"docgent-backend/internal/domain"
)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

func (p *Parser) Execute(diff string) []domain.Diff {
	var changes []domain.Diff
	var currentChange *domain.Diff

	lines := strings.Split(diff, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "diff --git") {
			if currentChange != nil {
				changes = append(changes, *currentChange)
			}
			parts := strings.Split(line, " ")
			currentChange = &domain.Diff{
				OldName:   strings.TrimPrefix(unquote(decodeOctalEscapes(parts[2])), "a/"),
				NewName:   strings.TrimPrefix(unquote(decodeOctalEscapes(parts[3])), "b/"),
				IsNewFile: false,
			}
		} else if strings.HasPrefix(line, "new file mode") {
			currentChange.IsNewFile = true
			currentChange.OldName = ""
		} else if strings.HasPrefix(line, "+++") || strings.HasPrefix(line, "---") {
			continue
		} else if strings.HasPrefix(line, "@@") || strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") || strings.HasPrefix(line, " ") {
			currentChange.Body += line + "\n"
		}
	}

	if currentChange != nil {
		changes = append(changes, *currentChange)
	}

	return changes
}

func decodeOctalEscapes(s string) string {
	var buf strings.Builder
	i := 0
	for i < len(s) {
		if s[i] == '\\' && i+3 < len(s) {
			// 8進数エスケープの可能性をチェック
			if isOctal(s[i+1]) && isOctal(s[i+2]) && isOctal(s[i+3]) {
				// 8進数値を変換
				octStr := s[i+1 : i+4]
				num, _ := strconv.ParseInt(octStr, 8, 32)
				buf.WriteByte(byte(num))
				i += 4
				continue
			}
		}
		// 通常の文字を書き込み
		buf.WriteByte(s[i])
		i++
	}
	return buf.String()
}

func isOctal(c byte) bool {
	return c >= '0' && c <= '7'
}

func unquote(s string) string {
	ss := strings.TrimPrefix(s, "\"")
	ss = strings.TrimSuffix(ss, "\"")
	return ss
}
