package tools

import (
	"context"
	"testing"
)

func TestCalculatorTool(t *testing.T) {
	calc := &CalculatorTool{}
	ctx := context.Background()

	tests := []struct {
		name     string
		args     string
		expected string
		hasError bool
	}{
		{
			name:     "addition",
			args:     `{"operation": "add", "a": 5, "b": 3}`,
			expected: "5 + 3 = 8",
			hasError: false,
		},
		{
			name:     "subtraction",
			args:     `{"operation": "subtract", "a": 10, "b": 4}`,
			expected: "10 - 4 = 6",
			hasError: false,
		},
		{
			name:     "multiplication",
			args:     `{"operation": "multiply", "a": 6, "b": 7}`,
			expected: "6 * 7 = 42",
			hasError: false,
		},
		{
			name:     "division",
			args:     `{"operation": "divide", "a": 15, "b": 3}`,
			expected: "15 / 3 = 5",
			hasError: false,
		},
		{
			name:     "division by zero",
			args:     `{"operation": "divide", "a": 10, "b": 0}`,
			expected: "Error: Division by zero is not allowed",
			hasError: false,
		},
		{
			name:     "invalid operation",
			args:     `{"operation": "power", "a": 2, "b": 3}`,
			expected: "",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := calc.Execute(ctx, tt.args)
			
			if tt.hasError && err == nil {
				t.Errorf("expected error but got none")
			}
			
			if !tt.hasError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			
			if !tt.hasError && result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestToolRegistry(t *testing.T) {
	registry := NewRegistry()
	
	// Test that all tools are registered.
	tools := registry.GetTools()
	expectedTools := []string{"get_weather", "get_today_date", "get_holidays", "calculate"}
	
	if len(tools) != len(expectedTools) {
		t.Errorf("expected %d tools, got %d", len(expectedTools), len(tools))
	}
	
	// Test tool execution.
	result, err := registry.Execute(context.Background(), "calculate", `{"operation": "add", "a": 2, "b": 3}`)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	
	expected := "2 + 3 = 5"
	if result != expected {
		t.Errorf("expected %q, got %q", expected, result)
	}
	
	// Test unknown tool.
	_, err = registry.Execute(context.Background(), "unknown_tool", "{}")
	if err == nil {
		t.Error("expected error for unknown tool")
	}
}