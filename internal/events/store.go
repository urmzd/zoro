package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/urmzd/zoro/internal/models"
)

// SQL to create chat tables. Called once at startup.
const migrateSQL = `
CREATE TABLE IF NOT EXISTS chat_session (
    id         TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS chat_event (
    id         TEXT PRIMARY KEY,
    session_id TEXT NOT NULL REFERENCES chat_session(id) ON DELETE CASCADE,
    type       TEXT NOT NULL,
    role       TEXT NOT NULL,
    content    TEXT NOT NULL DEFAULT '',
    tool_calls JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_chat_event_session ON chat_event(session_id);
`

type Store struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) EnsureSchema(ctx context.Context) error {
	_, err := s.pool.Exec(ctx, migrateSQL)
	return err
}

func (s *Store) CreateSession(ctx context.Context) (string, error) {
	id := uuid.New().String()
	_, err := s.pool.Exec(ctx,
		"INSERT INTO chat_session (id) VALUES ($1)", id)
	if err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}
	return id, nil
}

func (s *Store) AppendEvent(ctx context.Context, sessionID string, event models.ChatEvent) error {
	eventID := event.ID
	if eventID == "" {
		eventID = uuid.New().String()
	}

	var toolCallsJSON []byte
	if len(event.ToolCalls) > 0 {
		var err error
		toolCallsJSON, err = json.Marshal(event.ToolCalls)
		if err != nil {
			return fmt.Errorf("marshal tool_calls: %w", err)
		}
	}

	_, err := s.pool.Exec(ctx,
		"INSERT INTO chat_event (id, session_id, type, role, content, tool_calls) VALUES ($1, $2, $3, $4, $5, $6)",
		eventID, sessionID, event.Type, event.Role, event.Content, toolCallsJSON,
	)
	if err != nil {
		return fmt.Errorf("append event: %w", err)
	}
	return nil
}

func (s *Store) GetSession(ctx context.Context, sessionID string) (*models.ChatSession, error) {
	var createdAt time.Time
	err := s.pool.QueryRow(ctx,
		"SELECT created_at FROM chat_session WHERE id = $1", sessionID,
	).Scan(&createdAt)
	if err != nil {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	rows, err := s.pool.Query(ctx,
		"SELECT role, content, tool_calls FROM chat_event WHERE session_id = $1 ORDER BY created_at ASC",
		sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("query events: %w", err)
	}
	defer rows.Close()

	messages := make([]models.ChatMessage, 0)
	for rows.Next() {
		var role, content string
		var toolCallsJSON []byte
		if err := rows.Scan(&role, &content, &toolCallsJSON); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}
		msg := models.ChatMessage{Role: role, Content: content}
		if len(toolCallsJSON) > 0 {
			_ = json.Unmarshal(toolCallsJSON, &msg.ToolCalls)
		}
		messages = append(messages, msg)
	}

	return &models.ChatSession{
		ID:        sessionID,
		Messages:  messages,
		CreatedAt: createdAt,
	}, nil
}

func (s *Store) ListSessions(ctx context.Context) ([]models.ChatSessionSummary, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			s.id,
			s.created_at,
			COALESCE((
				SELECT e.content FROM chat_event e
				WHERE e.session_id = s.id AND e.role = 'user'
				ORDER BY e.created_at ASC LIMIT 1
			), '') AS preview,
			(SELECT count(*) FROM chat_event e WHERE e.session_id = s.id) AS msg_count
		FROM chat_session s
		ORDER BY s.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summaries := make([]models.ChatSessionSummary, 0)
	for rows.Next() {
		var sum models.ChatSessionSummary
		var preview string
		if err := rows.Scan(&sum.ID, &sum.CreatedAt, &preview, &sum.MessageCount); err != nil {
			return nil, err
		}
		if len(preview) > 120 {
			preview = preview[:120] + "..."
		}
		sum.Preview = preview
		summaries = append(summaries, sum)
	}
	return summaries, nil
}
