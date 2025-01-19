package autoagent

import (
	"fmt"
	"strings"
)

type SystemPrompt struct {
	agentBio             string
	toolUseGuidelines    []ToolUseGuideline
	taskInstruction      TaskInstruction
	conversationContexts []ConversationContext
}

func NewSystemPrompt(agentBio string, toolUseGuidelines []ToolUseGuideline, taskInstruction TaskInstruction, conversationContexts []ConversationContext) SystemPrompt {
	return SystemPrompt{
		agentBio:             agentBio,
		toolUseGuidelines:    toolUseGuidelines,
		taskInstruction:      taskInstruction,
		conversationContexts: conversationContexts,
	}
}

func (s SystemPrompt) String() string {
	var toolUseGuidelines []string
	for _, t := range s.toolUseGuidelines {
		toolUseGuidelines = append(toolUseGuidelines, t.String())
	}
	toolUseGuidelinesStr := strings.Join(toolUseGuidelines, "\n\n")

	taskInstructionStr := s.taskInstruction.String()

	var conversationContexts []string
	for _, c := range s.conversationContexts {
		conversationContexts = append(conversationContexts, c.String())
	}
	conversationContextsStr := strings.Join(conversationContexts, "\n\n")

	return fmt.Sprintf(`%s
	
====

RESPONSE FORMAT

You have to respond in the following XML-like format. You can use 3 response types:

# Complete response

When you have completed the task, respond with the following message:

<complete>[message]</complete>

# Error response

If something goes wrong, respond with the following message:

<error>[message]</error>

# Tool use response

If you want to use a tool, respond with the following message:

<tool_use:[tool_name]>
<[parameter_1_name]>[value1]</[parameter_1_name]>
<[parameter_2_name]>[value2]</[parameter_2_name]>
</tool_use:[tool_name]>

[xxx]: Dynamic values

====

TOOL USE

You have access to a set of tools. You can use one tool per message, and will receive the result of that tool use in the next message.

Almost all tools require parameters. You can find the required parameters in the tool description.

%s

====

YOUR TASK

%s

====

CONTEXT

Here is the current context of the conversation. Use this information to guide your responses.

%s`, s.agentBio, toolUseGuidelinesStr, taskInstructionStr, conversationContextsStr)

}

/**
 * ToolUseGuideline
 */

type ToolUseGuideline struct {
	name        string
	description string
	parameters  []ToolUseParameterGuideline
	example     string
}

func NewToolUseGuideline(name, description string, parameters []ToolUseParameterGuideline, example string) ToolUseGuideline {
	return ToolUseGuideline{
		name:        name,
		description: description,
		parameters:  parameters,
		example:     example,
	}
}

func (t ToolUseGuideline) String() string {
	var toolUseDescriptionStr string
	toolUseDescriptionStr += fmt.Sprintf("# %s\n\n%s\n\n## Parameters\n\n", t.name, t.description)
	var parameters []string
	for _, p := range t.parameters {
		parameters = append(parameters, p.String())
	}
	toolUseDescriptionStr += strings.Join(parameters, "\n")
	toolUseDescriptionStr += fmt.Sprintf("\n\n## Example\n\n%s", t.example)
	return toolUseDescriptionStr
}

/**
 * ToolUseParameterGuideline
 */

type ToolUseParameterGuideline struct {
	name        string
	description string
}

func NewToolUseParameterGuideline(name, description string) ToolUseParameterGuideline {
	return ToolUseParameterGuideline{
		name:        name,
		description: description,
	}
}

func (p ToolUseParameterGuideline) String() string {
	return fmt.Sprintf("%s: %s", p.name, p.description)
}

/**
 * TaskInstruction
 */

type TaskInstruction struct {
	description        string
	remainingStepCount int
}

type TaskInstructionOption func(*TaskInstruction)

func NewTaskInstruction(description string, options ...TaskInstructionOption) TaskInstruction {
	taskInstruction := &TaskInstruction{
		description: description,
	}

	for _, option := range options {
		option(taskInstruction)
	}

	return *taskInstruction
}

func WithRemainingStepCount(remainingStepCount int) TaskInstructionOption {
	return func(t *TaskInstruction) {
		t.remainingStepCount = remainingStepCount
	}
}

func (t TaskInstruction) String() string {
	if t.remainingStepCount == 0 {
		return t.description
	}
	return fmt.Sprintf("%s\n\nYou have to complete a task in %d step(s).", t.description, t.remainingStepCount)
}

/**
 * ConversationContext
 */

type ConversationContext struct {
	name string
	body string
}

func NewConversationContext(name, body string) ConversationContext {
	return ConversationContext{
		name: name,
		body: body,
	}
}

func (c ConversationContext) String() string {
	return fmt.Sprintf("# %s\n\n%s", c.name, c.body)
}
