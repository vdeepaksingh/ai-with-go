package chat

import (
	"context"
	"errors"
	"testing"

	"github.com/acai-travel/tech-challenge/internal/chat/model"
	. "github.com/acai-travel/tech-challenge/internal/chat/testing"
	"github.com/acai-travel/tech-challenge/internal/pb"
	"github.com/google/go-cmp/cmp"
	"github.com/twitchtv/twirp"
	"google.golang.org/protobuf/testing/protocmp"
)

// MockAssistant for testing
type MockAssistant struct {
	titleResponse string
	replyResponse string
	titleError    error
	replyError    error
}

func (m *MockAssistant) Title(ctx context.Context, conv *model.Conversation) (string, error) {
	if m.titleError != nil {
		return "", m.titleError
	}
	if m.titleResponse != "" {
		return m.titleResponse, nil
	}
	return "Test Title", nil
}

func (m *MockAssistant) Reply(ctx context.Context, conv *model.Conversation) (string, error) {
	if m.replyError != nil {
		return "", m.replyError
	}
	if m.replyResponse != "" {
		return m.replyResponse, nil
	}
	return "Test reply", nil
}

func TestServer_StartConversation(t *testing.T) {
	ctx := context.Background()

	t.Run("creates new conversation with title and reply", WithFixture(func(t *testing.T, f *Fixture) {
		mockAssist := &MockAssistant{
			titleResponse: "Weather Question",
			replyResponse: "It's sunny today!",
		}
		srv := NewServer(f.Repository, mockAssist)

		resp, err := srv.StartConversation(ctx, &pb.StartConversationRequest{
			Message: "What's the weather like?",
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify response structure
		if resp.GetConversationId() == "" {
			t.Error("expected conversation ID, got empty string")
		}
		if resp.GetTitle() != "Weather Question" {
			t.Errorf("expected title 'Weather Question', got %q", resp.GetTitle())
		}
		if resp.GetReply() != "It's sunny today!" {
			t.Errorf("expected reply 'It's sunny today!', got %q", resp.GetReply())
		}

		// Verify conversation was saved to database
		conv, err := f.Repository.DescribeConversation(ctx, resp.GetConversationId())
		if err != nil {
			t.Fatalf("failed to retrieve saved conversation: %v", err)
		}

		// Verify conversation has 2 messages (user + assistant)
		if len(conv.Messages) != 2 {
			t.Errorf("expected 2 messages, got %d", len(conv.Messages))
		}

		// Verify first message is user message
		if conv.Messages[0].Role != model.RoleUser {
			t.Errorf("expected first message to be user role, got %v", conv.Messages[0].Role)
		}
		if conv.Messages[0].Content != "What's the weather like?" {
			t.Errorf("expected user message content, got %q", conv.Messages[0].Content)
		}

		// Verify second message is assistant reply
		if conv.Messages[1].Role != model.RoleAssistant {
			t.Errorf("expected second message to be assistant role, got %v", conv.Messages[1].Role)
		}
		if conv.Messages[1].Content != "It's sunny today!" {
			t.Errorf("expected assistant message content, got %q", conv.Messages[1].Content)
		}
	}))

	t.Run("returns error when message is empty", WithFixture(func(t *testing.T, f *Fixture) {
		mockAssist := &MockAssistant{}
		srv := NewServer(f.Repository, mockAssist)

		_, err := srv.StartConversation(ctx, &pb.StartConversationRequest{
			Message: "   ", // Empty/whitespace message
		})

		if err == nil {
			t.Fatal("expected error for empty message, got nil")
		}

		if te, ok := err.(twirp.Error); !ok || te.Code() != twirp.InvalidArgument {
			t.Fatalf("expected twirp.InvalidArgument error, got %v", err)
		}
	}))

	t.Run("uses fallback title when assistant title fails", WithFixture(func(t *testing.T, f *Fixture) {
		mockAssist := &MockAssistant{
			titleError:    errors.New("title generation failed"),
			replyResponse: "Test reply",
		}
		srv := NewServer(f.Repository, mockAssist)

		resp, err := srv.StartConversation(ctx, &pb.StartConversationRequest{
			Message: "What is the weather like in Barcelona?",
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Should use fallback title generation
		expectedTitle := "the weather like in Barcelona"
		if resp.GetTitle() != expectedTitle {
			t.Errorf("expected fallback title %q, got %q", expectedTitle, resp.GetTitle())
		}
	}))

	t.Run("returns error when assistant reply fails", WithFixture(func(t *testing.T, f *Fixture) {
		mockAssist := &MockAssistant{
			titleResponse: "Test Title",
			replyError:    errors.New("reply generation failed"),
		}
		srv := NewServer(f.Repository, mockAssist)

		_, err := srv.StartConversation(ctx, &pb.StartConversationRequest{
			Message: "Hello",
		})

		if err == nil {
			t.Fatal("expected error when reply generation fails, got nil")
		}
	}))
}

func TestServer_DescribeConversation(t *testing.T) {
	ctx := context.Background()
	srv := NewServer(model.New(ConnectMongo()), nil)

	t.Run("describe existing conversation", WithFixture(func(t *testing.T, f *Fixture) {
		c := f.CreateConversation()

		out, err := srv.DescribeConversation(ctx, &pb.DescribeConversationRequest{ConversationId: c.ID.Hex()})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got, want := out.GetConversation(), c.Proto()
		if !cmp.Equal(got, want, protocmp.Transform()) {
			t.Errorf("DescribeConversation() mismatch (-got +want):\n%s", cmp.Diff(got, want, protocmp.Transform()))
		}
	}))

	t.Run("describe non existing conversation should return 404", WithFixture(func(t *testing.T, f *Fixture) {
		_, err := srv.DescribeConversation(ctx, &pb.DescribeConversationRequest{ConversationId: "08a59244257c872c5943e2a2"})
		if err == nil {
			t.Fatal("expected error for non-existing conversation, got nil")
		}

		if te, ok := err.(twirp.Error); !ok || te.Code() != twirp.NotFound {
			t.Fatalf("expected twirp.NotFound error, got %v", err)
		}
	}))
}
