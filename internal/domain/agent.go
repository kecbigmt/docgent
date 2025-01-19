package domain

import (
	"docgent-backend/internal/domain/autoagent"
	"fmt"
	"strings"
)

type ToolType int

const (
	FindFileTool ToolType = iota
	UpdateProposalDiffsTool
	UpdateProposalTextTool
)

func (t ToolType) String() string {
	switch t {
	case FindFileTool:
		return "find_file"
	case UpdateProposalDiffsTool:
		return "update_proposal_diffs"
	case UpdateProposalTextTool:
		return "update_proposal_text"
	default:
		return "unknown"
	}
}

type FindFileToolParams struct {
	Name string
}

type UpdateProposalDiffsToolParams struct {
	Diffs []Diff
}

type UpdateProposalTextToolParams struct {
	Title string
	Body  string
}

/**
 * ProposalRefineAgentResponse
 */

type ProposalRefineAgentResponse struct {
	Type       autoagent.ResponseType
	Message    string
	ToolType   ToolType
	ToolParams interface{}
}

func ParseResponseFromProposalRefineAgent(response autoagent.Response) (ProposalRefineAgentResponse, error) {
	switch response.ToolType {
	case "find_file":
		name, ok := response.ToolParams.GetOne("name")
		if !ok {
			return ProposalRefineAgentResponse{}, fmt.Errorf("missing name parameter")
		}
		return ProposalRefineAgentResponse{
			Type:       autoagent.ToolUseResponse,
			Message:    response.Message,
			ToolType:   FindFileTool,
			ToolParams: FindFileToolParams{Name: name},
		}, nil
	case "update_proposal_diffs":
		diffsStr := response.ToolParams.GetAll("diff")
		var diffs []Diff
		for _, diffStr := range diffsStr {
			diff, err := parseDiff(diffStr)
			if err != nil {
				return ProposalRefineAgentResponse{}, fmt.Errorf("failed to parse diff: %w", err)
			}
			diffs = append(diffs, diff)
		}
		return ProposalRefineAgentResponse{
			Type:       autoagent.ToolUseResponse,
			Message:    response.Message,
			ToolType:   UpdateProposalDiffsTool,
			ToolParams: UpdateProposalDiffsToolParams{Diffs: diffs},
		}, nil
	case "update_proposal_text":
		title, _ := response.ToolParams.GetOne("title")
		body, _ := response.ToolParams.GetOne("body")
		return ProposalRefineAgentResponse{
			Type:       autoagent.ToolUseResponse,
			Message:    response.Message,
			ToolType:   UpdateProposalTextTool,
			ToolParams: UpdateProposalTextToolParams{Title: title, Body: body},
		}, nil
	default:
		return ProposalRefineAgentResponse{}, fmt.Errorf("unknown tool type: %s", response.ToolType)
	}
}

func parseDiff(diffStr string) (Diff, error) {
	var oldName string
	var newName string
	var contentLines []string
	for _, line := range strings.Split(diffStr, "\n") {
		if strings.HasPrefix(line, "--- ") {
			parts := strings.Split(line, " ")
			if len(parts) < 2 {
				return Diff{}, fmt.Errorf("invalid diff format")
			}
			oldName = strings.TrimPrefix(parts[1], "a/")
		} else if strings.HasPrefix(line, "+++ ") {
			parts := strings.Split(line, " ")
			if len(parts) < 2 {
				return Diff{}, fmt.Errorf("invalid diff format")
			}
			newName = strings.TrimPrefix(parts[1], "b/")
		} else if strings.HasPrefix(line, "@@ ") || strings.HasPrefix(line, "-") || strings.HasPrefix(line, "+") || strings.HasPrefix(line, " ") {
			contentLines = append(contentLines, line)
		}
	}
	content := strings.Join(contentLines, "\n")

	if oldName == "/dev/null" {
		return NewCreateDiff(newName, content), nil
	}
	return NewUpdateDiff(oldName, newName, content), nil
}

/**
 * ProposalRefineSystemPrompt
 */

type ProposalRefineSystemPrompt struct {
	proposal           Proposal
	remainingStepCount int
	messageHistory     autoagent.MessageHistory
}

func NewProposalRefineSystemPrompt(proposal Proposal, remainingStepCount int, messageHistory autoagent.MessageHistory) ProposalRefineSystemPrompt {
	return ProposalRefineSystemPrompt{
		proposal:           proposal,
		remainingStepCount: remainingStepCount,
		messageHistory:     messageHistory,
	}
}

func (p ProposalRefineSystemPrompt) ToMessage() autoagent.Message {
	return autoagent.Message{
		Role:    "system",
		Content: p.String(),
	}
}

func (p ProposalRefineSystemPrompt) String() string {
	systemPrompt := autoagent.NewSystemPrompt(
		"You are Docgent, a highly skilled documentation agent.",
		[]autoagent.ToolUseGuideline{
			autoagent.NewToolUseGuideline(
				"find_file",
				"It allows you to find a file by its name.",
				[]autoagent.ToolUseParameterGuideline{
					autoagent.NewToolUseParameterGuideline("name", "The name of the file you want to find."),
				},
				`<tool_use:find_file>
<message>Finding a document file...</message>
<param:name>how-to-use-docgent.md</param:name>
</tool_use:find_file>`,
			),
			autoagent.NewToolUseGuideline(
				"update_proposal_diffs",
				"It allows you to update the proposal diffs.",
				[]autoagent.ToolUseParameterGuideline{
					autoagent.NewToolUseParameterGuideline("diff", "The diff you want to add to the proposal diffs. Multiple diffs can be added. diff should be a valid unified format."),
				},
				`<tool_use:update_proposal_diffs>
<message>Updating proposal...</message>
<param:diff>--- a/how-to-use-docgent.md
+++ b/how-to-use-docgent.md
@@ -1,3 +1,3 @@
-Hello
+Hi
 World
</param:diff>
<param:diff>
--- /dev/null
+++ b/bast-practice-for-docgent.md
@@ -0,0 +1,2 @@
+This is a new file.
+It has a few lines of text.
</param:diff>
</tool_use:update_proposal_diffs>`,
			),
			autoagent.NewToolUseGuideline(
				"update_proposal_text",
				"It allows you to update the proposal text. Both of these parameters are optional, but at least one of them must be present.",
				[]autoagent.ToolUseParameterGuideline{
					autoagent.NewToolUseParameterGuideline("title", "The title of the proposal."),
					autoagent.NewToolUseParameterGuideline("body", "The body of the proposal."),
				},
				`<tool_use:update_proposal_text>
<message>Updating proposal...</message>
<param:title>[title]</param:title>
<param:body>[body]</param:body>
</tool_use:update_proposal_text>`,
			),
		},
		autoagent.NewTaskInstruction(
			"You are currently refining your proposal.",
			autoagent.WithRemainingStepCount(p.remainingStepCount),
		),
		[]autoagent.ConversationContext{
			autoagent.NewConversationContext(
				"Proposal Title",
				p.proposal.Title,
			),
			autoagent.NewConversationContext(
				"Proposal Body",
				p.proposal.Body,
			),
			autoagent.NewConversationContext(
				"Proposal Diffs",
				p.proposal.Diffs.ToXMLString(),
			),
			autoagent.NewConversationContext(
				"Message History",
				p.messageHistory.ToXMLString(),
			),
		},
	)

	return systemPrompt.String()
}
