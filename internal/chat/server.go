package chat

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/acai-travel/tech-challenge/internal/chat/model"
	"github.com/acai-travel/tech-challenge/internal/pb"
	"github.com/twitchtv/twirp"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

var _ pb.ChatService = (*Server)(nil)

type Assistant interface {
	Title(ctx context.Context, conv *model.Conversation) (string, error)
	Reply(ctx context.Context, conv *model.Conversation) (string, error)
}

type Server struct {
	repo   *model.Repository
	assist Assistant
}

func NewServer(repo *model.Repository, assist Assistant) *Server {
	return &Server{repo: repo, assist: assist}
}

func (s *Server) StartConversation(ctx context.Context, req *pb.StartConversationRequest) (*pb.StartConversationResponse, error) {
	tracer := otel.Tracer("chat-service")
	ctx, span := tracer.Start(ctx, "StartConversation")
	defer span.End()

	span.SetAttributes(
		attribute.Int("message.length", len(req.GetMessage())),
	)

	conversation := &model.Conversation{
		ID:        primitive.NewObjectID(),
		Title:     "Untitled conversation",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Messages: []*model.Message{{
			ID:        primitive.NewObjectID(),
			Role:      model.RoleUser,
			Content:   req.GetMessage(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}},
	}

	if strings.TrimSpace(req.GetMessage()) == "" {
		return nil, twirp.RequiredArgumentError("message")
	}

	// Save conversation early with placeholder title
	if err := s.repo.CreateConversation(ctx, conversation); err != nil {
		return nil, err
	}

	// Run title and reply generation in parallel with timeouts
	titleChan := make(chan string, 1)
	replyChan := make(chan string, 1)
	errorChan := make(chan error, 2)

	// Generate title concurrently with timeout
	go func() {
		titleCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		title, err := s.assist.Title(titleCtx, conversation)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to generate conversation title", "error", err)
			title = s.generateFallbackTitle(req.GetMessage())
		}
		titleChan <- title
	}()

	// Generate reply concurrently with timeout
	go func() {
		replyCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()
		reply, err := s.assist.Reply(replyCtx, conversation)
		if err != nil {
			errorChan <- err
			return
		}
		replyChan <- reply
	}()

	// Wait for reply (critical path)
	var reply string
	select {
	case reply = <-replyChan:
	case err := <-errorChan:
		span.RecordError(err)
		return nil, err
	}

	// Get title (may still be generating)
	title := <-titleChan

	// Update conversation with reply and final title
	conversation.Title = title
	conversation.Messages = append(conversation.Messages, &model.Message{
		ID:        primitive.NewObjectID(),
		Role:      model.RoleAssistant,
		Content:   reply,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	if err := s.repo.UpdateConversation(ctx, conversation); err != nil {
		span.RecordError(err)
		return nil, err
	}

	span.SetAttributes(
		attribute.String("conversation.id", conversation.ID.Hex()),
		attribute.String("conversation.title", conversation.Title),
	)

	return &pb.StartConversationResponse{
		ConversationId: conversation.ID.Hex(),
		Title:          conversation.Title,
		Reply:          reply,
	}, nil
}

// generateFallbackTitle creates a simple title from the user message
func (s *Server) generateFallbackTitle(message string) string {
	// Clean and truncate the message
	title := strings.TrimSpace(message)
	title = strings.ReplaceAll(title, "\n", " ")
	
	// Remove question words and common prefixes
	prefixes := []string{"what is", "what's", "how do", "how to", "can you", "please", "tell me"}
	for _, prefix := range prefixes {
		if strings.HasPrefix(strings.ToLower(title), prefix) {
			title = strings.TrimSpace(title[len(prefix):])
			break
		}
	}
	
	// Remove question marks and extra punctuation
	title = strings.Trim(title, "?!.,:;")
	
	// Limit to 50 characters
	if len(title) > 50 {
		title = title[:50] + "..."
	}
	
	// Fallback if title is empty or too short
	if len(strings.TrimSpace(title)) < 3 {
		return "New conversation"
	}
	
	return title
}

func (s *Server) ContinueConversation(ctx context.Context, req *pb.ContinueConversationRequest) (*pb.ContinueConversationResponse, error) {
	if req.GetConversationId() == "" {
		return nil, twirp.RequiredArgumentError("conversation_id")
	}

	if strings.TrimSpace(req.GetMessage()) == "" {
		return nil, twirp.RequiredArgumentError("message")
	}

	conversation, err := s.repo.DescribeConversation(ctx, req.GetConversationId())
	if err != nil {
		return nil, err
	}

	conversation.UpdatedAt = time.Now()
	conversation.Messages = append(conversation.Messages, &model.Message{
		ID:        primitive.NewObjectID(),
		Role:      model.RoleUser,
		Content:   req.GetMessage(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	reply, err := s.assist.Reply(ctx, conversation)
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}

	conversation.Messages = append(conversation.Messages, &model.Message{
		ID:        primitive.NewObjectID(),
		Role:      model.RoleAssistant,
		Content:   reply,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})

	if err := s.repo.UpdateConversation(ctx, conversation); err != nil {
		return nil, twirp.InternalErrorWith(err)
	}

	return &pb.ContinueConversationResponse{Reply: reply}, nil
}

func (s *Server) ListConversations(ctx context.Context, req *pb.ListConversationsRequest) (*pb.ListConversationsResponse, error) {
	conversations, err := s.repo.ListConversations(ctx)
	if err != nil {
		return nil, twirp.InternalErrorWith(err)
	}

	resp := &pb.ListConversationsResponse{}
	for _, conv := range conversations {
		conv.Messages = nil // Clear messages to avoid sending large data
		resp.Conversations = append(resp.Conversations, conv.Proto())
	}

	return resp, nil
}

func (s *Server) DescribeConversation(ctx context.Context, req *pb.DescribeConversationRequest) (*pb.DescribeConversationResponse, error) {
	if req.GetConversationId() == "" {
		return nil, twirp.RequiredArgumentError("conversation_id")
	}

	conversation, err := s.repo.DescribeConversation(ctx, req.GetConversationId())
	if err != nil {
		return nil, err
	}

	if conversation == nil {
		return nil, twirp.NotFoundError("conversation not found")
	}

	return &pb.DescribeConversationResponse{Conversation: conversation.Proto()}, nil
}
