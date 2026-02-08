package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/widia-io/widia-omni/internal/domain"
)

type JournalService struct {
	db *pgxpool.Pool
}

func NewJournalService(db *pgxpool.Pool) *JournalService {
	return &JournalService{db: db}
}

func (s *JournalService) List(ctx context.Context, wsID uuid.UUID, limit, offset int) ([]domain.JournalEntry, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	rows, err := s.db.Query(ctx, `
		SELECT id, workspace_id, date, mood, energy, wins, challenges, gratitude, notes, tags,
			   created_at, updated_at
		FROM journal_entries
		WHERE workspace_id = $1
		ORDER BY date DESC
		LIMIT $2 OFFSET $3
	`, wsID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []domain.JournalEntry
	for rows.Next() {
		var e domain.JournalEntry
		if err := rows.Scan(&e.ID, &e.WorkspaceID, &e.Date, &e.Mood, &e.Energy,
			&e.Wins, &e.Challenges, &e.Gratitude, &e.Notes, &e.Tags,
			&e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}

func (s *JournalService) GetByDate(ctx context.Context, wsID uuid.UUID, date time.Time) (*domain.JournalEntry, error) {
	var e domain.JournalEntry
	err := s.db.QueryRow(ctx, `
		SELECT id, workspace_id, date, mood, energy, wins, challenges, gratitude, notes, tags,
			   created_at, updated_at
		FROM journal_entries
		WHERE workspace_id = $1 AND date = $2
	`, wsID, date).Scan(&e.ID, &e.WorkspaceID, &e.Date, &e.Mood, &e.Energy,
		&e.Wins, &e.Challenges, &e.Gratitude, &e.Notes, &e.Tags,
		&e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

type UpsertJournalRequest struct {
	Mood       *int16   `json:"mood"`
	Energy     *int16   `json:"energy"`
	Wins       []string `json:"wins"`
	Challenges []string `json:"challenges"`
	Gratitude  []string `json:"gratitude"`
	Notes      *string  `json:"notes"`
	Tags       []string `json:"tags"`
}

func (s *JournalService) Upsert(ctx context.Context, wsID uuid.UUID, date time.Time, req UpsertJournalRequest) (*domain.JournalEntry, error) {
	var e domain.JournalEntry
	err := s.db.QueryRow(ctx, `
		INSERT INTO journal_entries (workspace_id, date, mood, energy, wins, challenges, gratitude, notes, tags)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (workspace_id, date)
		DO UPDATE SET mood = $3, energy = $4, wins = $5, challenges = $6, gratitude = $7,
					  notes = $8, tags = $9, updated_at = now()
		RETURNING id, workspace_id, date, mood, energy, wins, challenges, gratitude, notes, tags,
				  created_at, updated_at
	`, wsID, date, req.Mood, req.Energy, req.Wins, req.Challenges, req.Gratitude, req.Notes, req.Tags).Scan(
		&e.ID, &e.WorkspaceID, &e.Date, &e.Mood, &e.Energy,
		&e.Wins, &e.Challenges, &e.Gratitude, &e.Notes, &e.Tags,
		&e.CreatedAt, &e.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

func (s *JournalService) Delete(ctx context.Context, wsID uuid.UUID, date time.Time) error {
	_, err := s.db.Exec(ctx, `
		DELETE FROM journal_entries WHERE workspace_id = $1 AND date = $2
	`, wsID, date)
	return err
}
