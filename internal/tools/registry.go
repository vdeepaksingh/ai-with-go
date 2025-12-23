// Package tools provides a flexible architecture for AI assistant tools.
package tools

import (
	"context"
	"fmt"

	"github.com/openai/openai-go/v2"
)

// Tool defines the interface that all assistant tools must implement.
type Tool interface {
	Name() string
	Description() string
	Parameters() openai.FunctionParameters
	Execute(ctx context.Context, args string) (string, error)
}

// Registry manages the collection of available tools for the assistant.
type Registry struct {
	tools map[string]Tool
}

// NewRegistry creates a new tool registry with all available tools registered.
func NewRegistry() *Registry {
	r := &Registry{tools: make(map[string]Tool)}
	
	// Register all available tools.
	r.Register(&WeatherTool{})
	r.Register(&DateTool{})
	r.Register(&HolidaysTool{})
	r.Register(&CalculatorTool{})
	
	return r
}

// Register adds a tool to the registry.
func (r *Registry) Register(tool Tool) {
	r.tools[tool.Name()] = tool
}

// GetTools returns all tools in OpenAI function format.
func (r *Registry) GetTools() []openai.ChatCompletionToolUnionParam {
	var tools []openai.ChatCompletionToolUnionParam
	
	for _, tool := range r.tools {
		tools = append(tools, openai.ChatCompletionFunctionTool(openai.FunctionDefinitionParam{
			Name:        tool.Name(),
			Description: openai.String(tool.Description()),
			Parameters:  tool.Parameters(),
		}))
	}
	
	return tools
}

// Execute runs the specified tool with the given arguments.
func (r *Registry) Execute(ctx context.Context, name, args string) (string, error) {
	tool, exists := r.tools[name]
	if !exists {
		return "", fmt.Errorf("unknown tool: %s", name)
	}
	return tool.Execute(ctx, args)
}