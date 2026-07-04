package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/bachtiarpanjaitan/ihandai-go/pkg/core"
	"github.com/bachtiarpanjaitan/ihandai-go/pkg/tools"
)

// ReAct implements the Reason + Act agent pattern.
// The agent iterates: think → act → observe → repeat until the goal is reached
// or the maximum steps are exceeded.
type ReAct struct {
	cfg Config
}

// NewReAct creates a new ReAct agent with the given configuration.
func NewReAct(cfg Config) *ReAct {
	if cfg.MaxSteps <= 0 {
		cfg.MaxSteps = 10
	}
	if cfg.SystemPrompt == "" {
		cfg.SystemPrompt = defaultSystemPrompt
	}
	return &ReAct{cfg: cfg}
}

// Run implements Agent.
func (r *ReAct) Run(ctx context.Context, goal string) (*Result, error) {
	result := &Result{}

	// Build tool descriptions
	toolDesc := r.buildToolDescriptions()

	// Initial messages: system prompt + user goal
	messages := []core.Message{
		{Role: "system", Content: r.cfg.SystemPrompt + "\n\n" + toolDesc},
		{Role: "user", Content: goal},
	}

	for step := 0; step < r.cfg.MaxSteps; step++ {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		// Call LLM
		resp, err := r.cfg.LLM.Chat(ctx, messages)
		if err != nil {
			return result, fmt.Errorf("agent step %d: %w", step, err)
		}
		result.TokensUsed += tokenEstimate(resp.Content)

		// Parse the response
		agentStep, isFinal := r.parseResponse(resp.Content)
		result.Steps = append(result.Steps, agentStep)

		if isFinal {
			result.Answer = agentStep.Observation // observation holds final answer
			break
		}

		// Execute the tool
		if agentStep.Action.Name != "" {
			obs := r.executeTool(ctx, agentStep.Action)
			agentStep.Observation = obs
			// Update the step in result (it was a copy)
			result.Steps[len(result.Steps)-1] = agentStep

			// Feed observation back to LLM
			messages = append(messages,
				core.Message{Role: "assistant", Content: resp.Content},
				core.Message{Role: "user", Content: fmt.Sprintf("Observation: %s", obs)},
			)
		} else {
			// No action — prompt the LLM to try again
			messages = append(messages,
				core.Message{Role: "assistant", Content: resp.Content},
				core.Message{Role: "user", Content: "You must either call a tool or give a Final Answer. Use the format exactly as specified."},
			)
		}
	}

	// If we exhausted steps without a final answer, generate one
	if result.Answer == "" {
		result.Answer = "Agent reached maximum steps without a final answer."
	}

	result.TokensUsed += tokenEstimate(result.Answer)
	return result, nil
}

var (
	actionRe = regexp.MustCompile(`Action:\s*(.+)`)
	inputRe  = regexp.MustCompile(`Action Input:\s*(.+)`)
	thoughtRe = regexp.MustCompile(`Thought:\s*(.+)`)
	finalRe  = regexp.MustCompile(`Final Answer:\s*(.+)`)
)

func (r *ReAct) parseResponse(text string) (Step, bool) {
	s := Step{}

	// Extract thought
	if m := thoughtRe.FindStringSubmatch(text); len(m) > 1 {
		s.Thought = strings.TrimSpace(m[1])
	}

	// Check for final answer
	if m := finalRe.FindStringSubmatch(text); len(m) > 1 {
		s.Observation = strings.TrimSpace(m[1])
		return s, true
	}

	// Extract action
	if m := actionRe.FindStringSubmatch(text); len(m) > 1 {
		actionStr := strings.TrimSpace(m[1])

		// Parse "tool_name(args)" format
		parenIdx := strings.Index(actionStr, "(")
		if parenIdx > 0 {
			s.Action.Name = strings.TrimSpace(actionStr[:parenIdx])
			inputStr := actionStr[parenIdx+1:]
			if lastParen := strings.LastIndex(inputStr, ")"); lastParen > 0 {
				inputStr = inputStr[:lastParen]
			}
			s.Action.Input = strings.TrimSpace(inputStr)
		} else {
			s.Action.Name = actionStr
		}
	} else if m := inputRe.FindStringSubmatch(text); len(m) > 1 {
		// Alternative format: Action Input: ...
		s.Action.Input = strings.TrimSpace(m[1])
	}

	return s, false
}

func (r *ReAct) executeTool(ctx context.Context, call ToolCall) string {
	var tool tools.Tool
	for _, t := range r.cfg.Tools {
		if strings.EqualFold(t.Name(), call.Name) {
			tool = t
			break
		}
	}
	if tool == nil {
		return fmt.Sprintf("Error: unknown tool %q. Available: %s", call.Name, r.toolNames())
	}

	input := json.RawMessage(call.Input)
	if !json.Valid(input) {
		// Try wrapping in quotes if not valid JSON
		input = json.RawMessage(fmt.Sprintf("%q", call.Input))
	}

	output, err := tool.Execute(ctx, input)
	if err != nil {
		return fmt.Sprintf("Error executing %s: %v", call.Name, err)
	}

	return string(output)
}

func (r *ReAct) buildToolDescriptions() string {
	if len(r.cfg.Tools) == 0 {
		return "No tools available."
	}

	var b strings.Builder
	b.WriteString("Available tools:\n")
	for _, t := range r.cfg.Tools {
		schema, _ := json.Marshal(t.InputSchema())
		b.WriteString(fmt.Sprintf("- %s: %s\n  Schema: %s\n", t.Name(), t.Description(), string(schema)))
	}
	return b.String()
}

func (r *ReAct) toolNames() string {
	names := make([]string, len(r.cfg.Tools))
	for i, t := range r.cfg.Tools {
		names[i] = t.Name()
	}
	return strings.Join(names, ", ")
}

func tokenEstimate(text string) int {
	// ~4 chars per token for English
	if len(text) == 0 {
		return 0
	}
	return len(text) / 4
}
