package events

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	surrealdb "github.com/surrealdb/surrealdb.go"
	sdbmodels "github.com/surrealdb/surrealdb.go/pkg/models"

	"github.com/urmzd/zoro/internal/models"
)

type Store struct {
	db *surrealdb.DB
}

func New(db *surrealdb.DB) *Store {
	return &Store{db: db}
}

func (s *Store) EnsureSchema(ctx context.Context) error {
	statements := []string{
		"DEFINE TABLE IF NOT EXISTS chat_session SCHEMAFULL",
		"DEFINE FIELD IF NOT EXISTS created_at ON chat_session TYPE datetime DEFAULT time::now()",
		"DEFINE TABLE IF NOT EXISTS chat_event SCHEMAFULL",
		"DEFINE FIELD IF NOT EXISTS session_id ON chat_event TYPE string",
		"DEFINE FIELD IF NOT EXISTS type ON chat_event TYPE string",
		"DEFINE FIELD IF NOT EXISTS role ON chat_event TYPE string",
		"DEFINE FIELD IF NOT EXISTS content ON chat_event TYPE string",
		"DEFINE FIELD IF NOT EXISTS tool_calls ON chat_event TYPE option<array>",
		"DEFINE FIELD IF NOT EXISTS created_at ON chat_event TYPE datetime DEFAULT time::now()",
		"DEFINE INDEX IF NOT EXISTS event_session ON chat_event FIELDS session_id",
		"DEFINE TABLE IF NOT EXISTS produced SCHEMAFULL TYPE RELATION IN chat_event OUT entity",
	}

	for _, stmt := range statements {
		if _, err := surrealdb.Query[any](ctx, s.db, stmt, nil); err != nil {
			log.Printf("event schema warning: %v (statement: %s)", err, stmt)
		}
	}
	return nil
}

func (s *Store) CreateSession(ctx context.Context) (string, error) {
	sessionID := uuid.New().String()

	_, err := surrealdb.Query[any](ctx, s.db,
		"CREATE type::record('chat_session', $id) SET created_at = time::now()",
		map[string]any{"id": sessionID},
	)
	if err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}

	return sessionID, nil
}

func (s *Store) AppendEvent(ctx context.Context, sessionID string, event models.ChatEvent) error {
	eventID := event.ID
	if eventID == "" {
		eventID = uuid.New().String()
	}

	params := map[string]any{
		"id":         eventID,
		"session_id": sessionID,
		"type":       event.Type,
		"role":       event.Role,
		"content":    event.Content,
	}

	query := "CREATE type::record('chat_event', $id) SET session_id = $session_id, type = $type, role = $role, content = $content"
	if len(event.ToolCalls) > 0 {
		params["tool_calls"] = event.ToolCalls
		query += ", tool_calls = $tool_calls"
	}

	_, err := surrealdb.Query[any](ctx, s.db, query, params)
	if err != nil {
		return fmt.Errorf("append event: %w", err)
	}
	return nil
}

type sessionRow struct {
	CreatedAt time.Time `json:"created_at"`
}

type eventRow struct {
	Role      string            `json:"role"`
	Content   string            `json:"content"`
	ToolCalls []models.ToolCall `json:"tool_calls"`
	CreatedAt time.Time         `json:"created_at"`
}

func (s *Store) GetSession(ctx context.Context, sessionID string) (*models.ChatSession, error) {
	sessResult, err := surrealdb.Query[[]sessionRow](ctx, s.db,
		"SELECT created_at FROM type::record('chat_session', $id)",
		map[string]any{"id": sessionID},
	)
	if err != nil || sessResult == nil || len(*sessResult) == 0 || len((*sessResult)[0].Result) == 0 {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	eventResult, err := surrealdb.Query[[]eventRow](ctx, s.db,
		"SELECT type, role, content, tool_calls, created_at FROM chat_event WHERE session_id = $id ORDER BY created_at ASC",
		map[string]any{"id": sessionID},
	)

	messages := make([]models.ChatMessage, 0)
	if err == nil && eventResult != nil && len(*eventResult) > 0 {
		for _, evt := range (*eventResult)[0].Result {
			messages = append(messages, models.ChatMessage{
				Role:      evt.Role,
				Content:   evt.Content,
				ToolCalls: evt.ToolCalls,
			})
		}
	}

	return &models.ChatSession{
		ID:        sessionID,
		Messages:  messages,
		CreatedAt: (*sessResult)[0].Result[0].CreatedAt,
	}, nil
}

type sessionListRow struct {
	ID        *sdbmodels.RecordID `json:"id"`
	CreatedAt time.Time           `json:"created_at"`
}

type previewRow struct {
	Content string `json:"content"`
}

type countRow struct {
	Total int64 `json:"total"`
}

func (s *Store) ListSessions(ctx context.Context) ([]models.ChatSessionSummary, error) {
	sessResult, err := surrealdb.Query[[]sessionListRow](ctx, s.db,
		"SELECT id, created_at FROM chat_session ORDER BY created_at DESC",
		nil,
	)
	if err != nil {
		return nil, err
	}

	summaries := make([]models.ChatSessionSummary, 0)
	if sessResult == nil || len(*sessResult) == 0 {
		return summaries, nil
	}

	for _, sess := range (*sessResult)[0].Result {
		sessID := extractRecordIDFromModel(sess.ID)

		// Get preview (first user message)
		previewResult, err := surrealdb.Query[[]previewRow](ctx, s.db,
			`SELECT content, created_at FROM chat_event WHERE session_id = $id AND role = 'user' ORDER BY created_at ASC LIMIT 1`,
			map[string]any{"id": sessID},
		)

		preview := ""
		if err == nil && previewResult != nil && len(*previewResult) > 0 && len((*previewResult)[0].Result) > 0 {
			preview = (*previewResult)[0].Result[0].Content
			if len(preview) > 120 {
				preview = preview[:120] + "..."
			}
		}

		// Get message count
		countResult, _ := surrealdb.Query[[]countRow](ctx, s.db,
			`SELECT count() AS total FROM chat_event WHERE session_id = $id GROUP ALL`,
			map[string]any{"id": sessID},
		)

		var msgCount int64
		if countResult != nil && len(*countResult) > 0 && len((*countResult)[0].Result) > 0 {
			msgCount = (*countResult)[0].Result[0].Total
		}

		summaries = append(summaries, models.ChatSessionSummary{
			ID:           sessID,
			Preview:      preview,
			MessageCount: msgCount,
			CreatedAt:    sess.CreatedAt,
		})
	}

	return summaries, nil
}

func extractRecordIDFromModel(rid *sdbmodels.RecordID) string {
	if rid == nil {
		return ""
	}
	return fmt.Sprintf("%v", rid.ID)
}
