package tools

import (
	"context"
	"time"

	"github.com/openai/openai-go/v2"
)

// DateTool provides current date and time information.
type DateTool struct{}

// Name returns the tool's identifier.
func (d *DateTool) Name() string {
	return "get_today_date"
}

// Description returns what the tool does.
func (d *DateTool) Description() string {
	return "Get today's date and time in RFC3339 format"
}

// Parameters defines the tool's input schema.
func (d *DateTool) Parameters() openai.FunctionParameters {
	return openai.FunctionParameters{
		"type":       "object",
		"properties": map[string]any{},
	}
}

// Execute returns the current date and time.
func (d *DateTool) Execute(ctx context.Context, args string) (string, error) {
	return time.Now().Format(time.RFC3339), nil
}