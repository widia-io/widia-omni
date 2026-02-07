package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/widia-io/widia-omni/internal/domain"
)

type CounterService struct {
	db *pgxpool.Pool
}

func NewCounterService(db *pgxpool.Pool) *CounterService {
	return &CounterService{db: db}
}

func (s *CounterService) GetCounters(ctx context.Context, wsID uuid.UUID) (*domain.WorkspaceCounter, error) {
	var c domain.WorkspaceCounter
	err := s.db.QueryRow(ctx, `
		SELECT workspace_id, areas_count, goals_count, habits_count,
			   tasks_created_today, tasks_today_date, transactions_month_count,
			   transactions_month, storage_bytes_used, updated_at
		FROM workspace_counters WHERE workspace_id = $1
	`, wsID).Scan(
		&c.WorkspaceID, &c.AreasCount, &c.GoalsCount, &c.HabitsCount,
		&c.TasksCreatedToday, &c.TasksTodayDate, &c.TransactionsMonthCount,
		&c.TransactionsMonth, &c.StorageBytesUsed, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *CounterService) IncrementTasksToday(ctx context.Context, wsID uuid.UUID) error {
	_, err := s.db.Exec(ctx, `
		UPDATE workspace_counters
		SET tasks_created_today = CASE
				WHEN tasks_today_date = CURRENT_DATE THEN tasks_created_today + 1
				ELSE 1
			END,
			tasks_today_date = CURRENT_DATE
		WHERE workspace_id = $1
	`, wsID)
	return err
}

func (s *CounterService) IncrementTransactionsMonth(ctx context.Context, wsID uuid.UUID) error {
	_, err := s.db.Exec(ctx, `
		UPDATE workspace_counters
		SET transactions_month_count = CASE
				WHEN transactions_month = to_char(CURRENT_DATE, 'YYYY-MM') THEN transactions_month_count + 1
				ELSE 1
			END,
			transactions_month = to_char(CURRENT_DATE, 'YYYY-MM')
		WHERE workspace_id = $1
	`, wsID)
	return err
}

func (s *CounterService) ResetDaily(ctx context.Context) error {
	_, err := s.db.Exec(ctx, `
		UPDATE workspace_counters
		SET tasks_created_today = 0, tasks_today_date = CURRENT_DATE
		WHERE tasks_today_date < CURRENT_DATE
	`)
	return err
}
