package tools

import (
	"context"

	"github.com/openai/openai-go/v2"
)

// WeatherTool provides weather information for specified locations.
type WeatherTool struct{}

// Name returns the tool's identifier.
func (w *WeatherTool) Name() string {
	return "get_weather"
}

// Description returns what the tool does.
func (w *WeatherTool) Description() string {
	return "Get weather at the given location"
}

// Parameters defines the tool's input schema.
func (w *WeatherTool) Parameters() openai.FunctionParameters {
	return openai.FunctionParameters{
		"type": "object",
		"properties": map[string]any{
			"location": map[string]string{
				"type": "string",
			},
		},
		"required": []string{"location"},
	}
}

// Execute retrieves weather information for the specified location.
func (w *WeatherTool) Execute(ctx context.Context, args string) (string, error) {
	// TODO: Implement real weather API integration (Task 2).
	return "weather is fine", nil
}