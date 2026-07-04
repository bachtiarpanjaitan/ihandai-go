// Package agent defines the agent framework for autonomous AI task execution.
//
// Agents combine an LLM with tools to autonomously plan, execute, observe,
// and reflect until a goal is achieved. The ReAct (Reason + Act) pattern
// is the primary implementation.
package agent

import (
	"context"

	"github.com/bachtiarpanjaitan/ihandai-go/pkg/llm"
	"github.com/bachtiarpanjaitan/ihandai-go/pkg/tools"
)

// Agent is an autonomous AI that uses tools to achieve a goal.
type Agent interface {
	// Run executes the agent with the given goal and returns the final result.
	// The context controls cancellation and timeout.
	Run(ctx context.Context, goal string) (*Result, error)
}

// Config holds configuration for an agent.
type Config struct {
	// MaxSteps is the maximum number of thought-action cycles (default: 10).
	MaxSteps int

	// LLM is the chat completer used for reasoning.
	LLM llm.ChatCompleter

	// Tools are the tools available to the agent.
	Tools []tools.Tool

	// SystemPrompt overrides the default ReAct system prompt.
	SystemPrompt string
}

// Result contains the agent's final output.
type Result struct {
	// Answer is the final text answer from the agent.
	Answer string

	// Steps contains the history of all steps the agent took.
	Steps []Step

	// TokensUsed estimates the total tokens consumed.
	TokensUsed int
}

// Step represents one thought-action-observation cycle.
type Step struct {
	// Thought is the agent's reasoning before acting.
	Thought string

	// Action is the tool the agent decided to call (empty for final answer).
	Action ToolCall

	// Observation is the result of the tool call.
	Observation string
}

// ToolCall represents a tool invocation by the agent.
type ToolCall struct {
	// Name is the tool name.
	Name string

	// Input is the JSON input to the tool.
	Input string
}

// defaultSystemPrompt is the ReAct system prompt.
const defaultSystemPrompt = `You are an AI assistant that solves problems step by step.
You have access to tools that help you accomplish your goal.

Use this exact format:

Thought: <your reasoning about what to do next>
Action: tool_name({"key": "value"})

Or when you have the final answer:

Thought: I now have the answer.
Final Answer: <your answer to the user>

Always think step by step. If a tool call fails, try a different approach.
Do not repeat the same failed action.`
