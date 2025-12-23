package assistant

import (
	"context"
	"errors"
	"log/slog"
	"strings"

	"github.com/acai-travel/tech-challenge/internal/chat/model"
	"github.com/acai-travel/tech-challenge/internal/tools"
	"github.com/openai/openai-go/v2"
)

// maxToolCallIterations defines the maximum number of tool call iterations to prevent infinite loops.
const maxToolCallIterations = 15

// Assistant provides AI-powered conversation capabilities with tool support.
type Assistant struct {
	cli   openai.Client
	tools *tools.Registry
}

// New creates a new Assistant with OpenAI client and tool registry.
func New() *Assistant {
	return &Assistant{
		cli:   openai.NewClient(),
		tools: tools.NewRegistry(),
	}
}

func (a *Assistant) Title(ctx context.Context, conv *model.Conversation) (string, error) {
	if len(conv.Messages) == 0 {
		return "An empty conversation", nil
	}

	slog.InfoContext(ctx, "Generating title for conversation", "conversation_id", conv.ID)

	msgs := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage("Generate a concise, descriptive title (2-6 words) that summarizes the main topic of the user's question. Do not answer the question, just create a brief topic summary. Examples: 'Weather in Barcelona', 'Today's Date', 'Barcelona Holidays'."),
	}

	// Add only the first user message for title generation.
	for _, m := range conv.Messages {
		if m.Role == model.RoleUser {
			msgs = append(msgs, openai.UserMessage(m.Content))
			break
		}
	}

	resp, err := a.cli.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model:    openai.ChatModelO1,
		Messages: msgs,
	})

	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 || strings.TrimSpace(resp.Choices[0].Message.Content) == "" {
		return "", errors.New("empty response from OpenAI for title generation")
	}

	title := resp.Choices[0].Message.Content
	title = strings.ReplaceAll(title, "\n", " ")
	title = strings.Trim(title, " \t\r\n-\"'")

	if len(title) > 80 {
		title = title[:80]
	}

	return title, nil
}

func (a *Assistant) Reply(ctx context.Context, conv *model.Conversation) (string, error) {
	if len(conv.Messages) == 0 {
		return "", errors.New("conversation has no messages")
	}

	slog.InfoContext(ctx, "Generating reply for conversation", "conversation_id", conv.ID)

	msgs := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage("You are a helpful, concise AI assistant. Provide accurate, safe, and clear responses."),
	}

	for _, m := range conv.Messages {
		switch m.Role {
		case model.RoleUser:
			msgs = append(msgs, openai.UserMessage(m.Content))
		case model.RoleAssistant:
			msgs = append(msgs, openai.AssistantMessage(m.Content))
		}
	}

	for range maxToolCallIterations {
		resp, err := a.cli.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
			Model:    openai.ChatModelGPT4_1,
			Messages: msgs,
			Tools:    a.tools.GetTools(),
		})

		if err != nil {
			return "", err
		}

		if len(resp.Choices) == 0 {
			return "", errors.New("no choices returned by OpenAI")
		}

		if message := resp.Choices[0].Message; len(message.ToolCalls) > 0 {
			msgs = append(msgs, message.ToParam())

			for _, call := range message.ToolCalls {
				slog.InfoContext(ctx, "Tool call received", "name", call.Function.Name, "args", call.Function.Arguments)

				// Execute tool using registry.
				result, err := a.tools.Execute(ctx, call.Function.Name, call.Function.Arguments)
				if err != nil {
					result = "Error executing tool: " + err.Error()
				}
				msgs = append(msgs, openai.ToolMessage(result, call.ID))
			}

			continue
		}

		return resp.Choices[0].Message.Content, nil
	}

	return "", errors.New("too many tool calls, unable to generate reply")
}
