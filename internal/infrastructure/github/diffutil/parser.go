package diffutil

import (
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
				OldPath:   strings.TrimPrefix(parts[2], "a/"),
				NewPath:   strings.TrimPrefix(parts[3], "b/"),
				Body:      line + "\n",
				IsNewFile: false,
			}
		} else if currentChange != nil {
			if strings.HasPrefix(line, "new file mode") {
				currentChange.IsNewFile = true
				currentChange.OldPath = ""
			}
			currentChange.Body += line + "\n"
		}
	}

	if currentChange != nil {
		changes = append(changes, *currentChange)
	}

	return changes
}
