// Package tools provides built-in agent tools.
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/bachtiarpanjaitan/ihandai-go/pkg/core"
)

// Calculator is a simple arithmetic tool for agent use.
type Calculator struct{}

// NewCalculator creates a new Calculator tool.
func NewCalculator() *Calculator { return &Calculator{} }

func (c *Calculator) Name() string        { return "calculator" }
func (c *Calculator) Description() string { return "Evaluate a mathematical expression. Supports +, -, *, /, and parentheses." }
func (c *Calculator) InputSchema() *core.JSONSchema {
	return &core.JSONSchema{
		Type: "object",
		Properties: map[string]*core.JSONSchemaProp{
			"expression": {Type: "string", Description: "The mathematical expression to evaluate (e.g., '2 + 3 * 4')"},
		},
		Required: []string{"expression"},
	}
}

func (c *Calculator) Execute(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	_ = ctx
	var params struct {
		Expression string `json:"expression"`
	}
	if err := json.Unmarshal(input, &params); err != nil {
		return nil, fmt.Errorf("calculator: %w", err)
	}

	result, err := evaluate(params.Expression)
	if err != nil {
		return json.RawMessage(fmt.Sprintf(`{"error":"%s"}`, err.Error())), nil
	}
	return json.RawMessage(fmt.Sprintf(`{"result":%s}`, strconv.FormatFloat(result, 'f', -1, 64))), nil
}

// Simple expression evaluator (no eval, just basic arithmetic).
func evaluate(expr string) (float64, error) {
	expr = strings.ReplaceAll(expr, " ", "")
	return parseAddSub(expr)
}

func parseAddSub(expr string) (float64, error) {
	// find the last + or - that is not inside parentheses
	depth := 0
	for i := len(expr) - 1; i >= 0; i-- {
		switch expr[i] {
		case ')':
			depth++
		case '(':
			depth--
		case '+':
			if depth == 0 {
				left, err := parseAddSub(expr[:i])
				if err != nil {
					return 0, err
				}
				right, err := parseMulDiv(expr[i+1:])
				if err != nil {
					return 0, err
				}
				return left + right, nil
			}
		case '-':
			if depth == 0 && i > 0 {
				left, err := parseAddSub(expr[:i])
				if err != nil {
					return 0, err
				}
				right, err := parseMulDiv(expr[i+1:])
				if err != nil {
					return 0, err
				}
				return left - right, nil
			}
		}
	}
	return parseMulDiv(expr)
}

func parseMulDiv(expr string) (float64, error) {
	depth := 0
	for i := len(expr) - 1; i >= 0; i-- {
		switch expr[i] {
		case ')':
			depth++
		case '(':
			depth--
		case '*':
			if depth == 0 {
				left, err := parseMulDiv(expr[:i])
				if err != nil {
					return 0, err
				}
				right, err := parseAtom(expr[i+1:])
				if err != nil {
					return 0, err
				}
				return left * right, nil
			}
		case '/':
			if depth == 0 {
				left, err := parseMulDiv(expr[:i])
				if err != nil {
					return 0, err
				}
				right, err := parseAtom(expr[i+1:])
				if err != nil {
					return 0, err
				}
				if right == 0 {
					return 0, fmt.Errorf("division by zero")
				}
				return left / right, nil
			}
		}
	}
	return parseAtom(expr)
}

func parseAtom(expr string) (float64, error) {
	if expr == "" {
		return 0, fmt.Errorf("empty expression")
	}
	if expr[0] == '(' && expr[len(expr)-1] == ')' {
		return parseAddSub(expr[1 : len(expr)-1])
	}
	if expr[0] == '-' {
		val, err := parseAtom(expr[1:])
		if err != nil {
			return 0, err
		}
		return -val, nil
	}
	return strconv.ParseFloat(expr, 64)
}
