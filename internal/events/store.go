package events

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	surrealdb "github.com/surrealdb/surrealdb.go"

	"github.com/urmzd/zoro/internal/models"
)

type Store struct {
	db  *surrealdb.DB
	ctx context.Context
}

func New(ctx context.Context, db *surrealdb.DB) *Store {
	return &Store{db: db, ctx: ctx}
}

func (s *Store) EnsureSchema() error {
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
		if _, err := surrealdb.Query[any](s.ctx, s.db, stmt, nil); err != nil {
			log.Printf("event schema warning: %v (statement: %s)", err, stmt)
		}
	}
	return nil
}

func (s *Store) CreateSession() (string, error) {
	sessionID := uuid.New().String()

	_, err := surrealdb.Query[any](s.ctx, s.db,
		"CREATE type::thing('chat_session', $id) SET created_at = time::now()",
		map[string]any{"id": sessionID},
	)
	if err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}

	return sessionID, nil
}

func (s *Store) AppendEvent(sessionID string, event models.ChatEvent) error {
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

	query := "CREATE type::thing('chat_event', $id) SET session_id = $session_id, type = $type, role = $role, content = $content"
	if len(event.ToolCalls) > 0 {
		params["tool_calls"] = event.ToolCalls
		query += ", tool_calls = $tool_calls"
	}

	_, err := surrealdb.Query[any](s.ctx, s.db, query, params)
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

func (s *Store) GetSession(sessionID string) (*models.ChatSession, error) {
	sessResult, err := surrealdb.Query[[]sessionRow](s.ctx, s.db,
		"SELECT created_at FROM type::thing('chat_session', $id)",
		map[string]any{"id": sessionID},
	)
	if err != nil || sessResult == nil || len(*sessResult) == 0 || len((*sessResult)[0].Result) == 0 {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	eventResult, err := surrealdb.Query[[]eventRow](s.ctx, s.db,
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
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

type previewRow struct {
	Content string `json:"content"`
	Total   *int64 `json:"total"`
}

func (s *Store) ListSessions() ([]models.ChatSessionSummary, error) {
	sessResult, err := surrealdb.Query[[]sessionListRow](s.ctx, s.db,
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
		sessID := extractRecordID(sess.ID)

		previewResult, err := surrealdb.Query[[]previewRow](s.ctx, s.db,
			`SELECT content,
				(SELECT VALUE count() FROM chat_event WHERE session_id = $id GROUP ALL) AS total
			FROM chat_event
			WHERE session_id = $id AND role = 'user'
			ORDER BY created_at ASC LIMIT 1`,
			map[string]any{"id": sessID},
		)

		preview := ""
		var msgCount int64
		if err == nil && previewResult != nil && len(*previewResult) > 0 && len((*previewResult)[0].Result) > 0 {
			p := (*previewResult)[0].Result[0]
			preview = p.Content
			if len(preview) > 120 {
				preview = preview[:120] + "..."
			}
			if p.Total != nil {
				msgCount = *p.Total
			}
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

func extractRecordID(recordID string) string {
	if pos := strings.LastIndex(recordID, ":"); pos >= 0 {
		return recordID[pos+1:]
	}
	return recordID
}
