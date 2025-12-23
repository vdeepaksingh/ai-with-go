package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/openai/openai-go/v2"
)

// CalculatorTool performs basic mathematical operations.
type CalculatorTool struct{}

// Name returns the tool's identifier.
func (c *CalculatorTool) Name() string {
	return "calculate"
}

// Description returns what the tool does.
func (c *CalculatorTool) Description() string {
	return "Perform basic mathematical calculations (add, subtract, multiply, divide)"
}

// Parameters defines the tool's input schema.
func (c *CalculatorTool) Parameters() openai.FunctionParameters {
	return openai.FunctionParameters{
		"type": "object",
		"properties": map[string]any{
			"operation": map[string]any{
				"type":        "string",
				"description": "The mathematical operation to perform",
				"enum":        []string{"add", "subtract", "multiply", "divide"},
			},
			"a": map[string]string{
				"type":        "number",
				"description": "First number",
			},
			"b": map[string]string{
				"type":        "number",
				"description": "Second number",
			},
		},
		"required": []string{"operation", "a", "b"},
	}
}

// Execute performs the specified mathematical operation.
func (c *CalculatorTool) Execute(ctx context.Context, args string) (string, error) {
	var payload struct {
		Operation string  `json:"operation"`
		A         float64 `json:"a"`
		B         float64 `json:"b"`
	}

	if err := json.Unmarshal([]byte(args), &payload); err != nil {
		return "", fmt.Errorf("failed to parse calculator arguments: %w", err)
	}

	var result float64
	var operation string

	switch strings.ToLower(payload.Operation) {
	case "add":
		result = payload.A + payload.B
		operation = "+"
	case "subtract":
		result = payload.A - payload.B
		operation = "-"
	case "multiply":
		result = payload.A * payload.B
		operation = "*"
	case "divide":
		if payload.B == 0 {
			return "Error: Division by zero is not allowed", nil
		}
		result = payload.A / payload.B
		operation = "/"
	default:
		return "", fmt.Errorf("unsupported operation: %s", payload.Operation)
	}

	// Format result nicely.
	resultStr := strconv.FormatFloat(result, 'f', -1, 64)
	aStr := strconv.FormatFloat(payload.A, 'f', -1, 64)
	bStr := strconv.FormatFloat(payload.B, 'f', -1, 64)

	return fmt.Sprintf("%s %s %s = %s", aStr, operation, bStr, resultStr), nil
}