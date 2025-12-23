package assistant

import (
	"context"
	"os"
	"testing"

	"github.com/acai-travel/tech-challenge/internal/chat/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestAssistant_Title(t *testing.T) {
	ctx := context.Background()

	t.Run("returns default title for empty conversation", func(t *testing.T) {
		assistant := New()
		conv := &model.Conversation{
			ID:       primitive.NewObjectID(),
			Messages: []*model.Message{}, // Empty messages
		}

		title, err := assistant.Title(ctx, conv)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		expected := "An empty conversation"
		if title != expected {
			t.Errorf("expected title %q, got %q", expected, title)
		}
	})

	t.Run("generates title from user message", func(t *testing.T) {
		// Skip this test if no OpenAI API key is set
		if !hasOpenAIKey() {
			t.Skip("Skipping OpenAI integration test - no API key set")
		}

		// This test would require actual OpenAI API calls
		// For now, just verify the assistant can be created
		assistant := New()
		if assistant == nil {
			t.Fatal("failed to create assistant")
		}
		t.Skip("OpenAI integration test - requires API key and network")
	})

	t.Run("handles timeout gracefully", func(t *testing.T) {
		t.Skip("Timeout test requires OpenAI integration - skipping for now")
	})

	t.Run("title length constraints", func(t *testing.T) {
		t.Skip("Length constraint test requires OpenAI integration - skipping for now")
	})
}

// hasOpenAIKey checks if OpenAI API key is available for integration tests
func hasOpenAIKey() bool {
	return os.Getenv("OPENAI_API_KEY") != ""
}