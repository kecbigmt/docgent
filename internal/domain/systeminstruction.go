package domain

import (
	"embed"
	"strings"
	"text/template"

	"docgent-backend/internal/domain/tooluse"
)

//go:embed systeminstruction-template.md
var templateFS embed.FS

type SystemInstruction struct {
	contexts []EnvironmentContext
	tools    []tooluse.Usage
}

func NewSystemInstruction(contexts []EnvironmentContext, tools []tooluse.Usage) *SystemInstruction {
	return &SystemInstruction{
		contexts: contexts,
		tools:    tools,
	}
}

func (s *SystemInstruction) String() string {
	tmpl, err := template.ParseFS(templateFS, "systeminstruction-template.md")
	if err != nil {
		panic(err)
	}

	ss := struct {
		Contexts []EnvironmentContext
		Tools    []tooluse.Usage
	}{
		Contexts: s.contexts,
		Tools:    s.tools,
	}

	var b strings.Builder
	if err := tmpl.Execute(&b, ss); err != nil {
		panic(err)
	}

	return b.String()
}

type EnvironmentContext struct {
	Name  string
	Value string
}

func NewEnvironmentContext(name string, value string) EnvironmentContext {
	return EnvironmentContext{
		Name:  name,
		Value: value,
	}
}
