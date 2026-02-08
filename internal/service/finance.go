package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/widia-io/widia-omni/internal/domain"
	"github.com/widia-io/widia-omni/internal/observability"
)

type FinanceService struct {
	db         *pgxpool.Pool
	counterSvc *CounterService
}

func NewFinanceService(db *pgxpool.Pool, counterSvc *CounterService) *FinanceService {
	return &FinanceService{db: db, counterSvc: counterSvc}
}

// --- DTOs ---

type CreateCategoryRequest struct {
	Name     string                 `json:"name"`
	Type     domain.TransactionType `json:"type"`
	Color    *string                `json:"color"`
	Icon     *string                `json:"icon"`
	ParentID *uuid.UUID             `json:"parent_id"`
}

type UpdateCategoryRequest struct {
	Name     string     `json:"name"`
	Color    *string    `json:"color"`
	Icon     *string    `json:"icon"`
	ParentID *uuid.UUID `json:"parent_id"`
}

type TransactionFilters struct {
	Type       *string    `json:"type"`
	CategoryID *uuid.UUID `json:"category_id"`
	AreaID     *uuid.UUID `json:"area_id"`
	DateFrom   *string    `json:"date_from"`
	DateTo     *string    `json:"date_to"`
	Tag        *string    `json:"tag"`
	Limit      int        `json:"limit"`
	Offset     int        `json:"offset"`
}

type CreateTransactionRequest struct {
	CategoryID     *uuid.UUID             `json:"category_id"`
	AreaID         *uuid.UUID             `json:"area_id"`
	Type           domain.TransactionType `json:"type"`
	Amount         float64                `json:"amount"`
	Description    *string                `json:"description"`
	Date           string                 `json:"date"`
	IsRecurring    bool                   `json:"is_recurring"`
	RecurrenceRule *string                `json:"recurrence_rule"`
	Tags           []string               `json:"tags"`
}

type UpdateTransactionRequest struct {
	CategoryID     *uuid.UUID             `json:"category_id"`
	AreaID         *uuid.UUID             `json:"area_id"`
	Type           domain.TransactionType `json:"type"`
	Amount         float64                `json:"amount"`
	Description    *string                `json:"description"`
	Date           string                 `json:"date"`
	IsRecurring    bool                   `json:"is_recurring"`
	RecurrenceRule *string                `json:"recurrence_rule"`
	Tags           []string               `json:"tags"`
}

type UpsertBudgetRequest struct {
	CategoryID *uuid.UUID `json:"category_id"`
	Month      string     `json:"month"`
	Amount     float64    `json:"amount"`
}

type FinanceSummary struct {
	Month            string             `json:"month"`
	TotalIncome      float64            `json:"total_income"`
	TotalExpenses    float64            `json:"total_expenses"`
	TotalInvestments float64            `json:"total_investments"`
	NetBalance       float64            `json:"net_balance"`
	ByCategory       []CategoryBreakdown `json:"by_category"`
	BudgetStatus     []BudgetComparison  `json:"budget_status"`
}

type CategoryBreakdown struct {
	CategoryID   *uuid.UUID `json:"category_id,omitempty"`
	CategoryName *string    `json:"category_name,omitempty"`
	Type         string     `json:"type"`
	Total        float64    `json:"total"`
	Count        int        `json:"count"`
}

type BudgetComparison struct {
	CategoryID   *uuid.UUID `json:"category_id,omitempty"`
	CategoryName *string    `json:"category_name,omitempty"`
	BudgetAmount float64    `json:"budget_amount"`
	SpentAmount  float64    `json:"spent_amount"`
	Remaining    float64    `json:"remaining"`
	Percentage   float64    `json:"percentage"`
}

// --- Categories ---

func (s *FinanceService) ListCategories(ctx context.Context, wsID uuid.UUID) ([]domain.FinanceCategory, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, workspace_id, name, type, color, icon, parent_id, is_system, created_at
		FROM finance_categories
		WHERE workspace_id = $1 AND deleted_at IS NULL
		ORDER BY name
	`, wsID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cats []domain.FinanceCategory
	for rows.Next() {
		var c domain.FinanceCategory
		if err := rows.Scan(&c.ID, &c.WorkspaceID, &c.Name, &c.Type, &c.Color, &c.Icon,
			&c.ParentID, &c.IsSystem, &c.CreatedAt); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, nil
}

func (s *FinanceService) CreateCategory(ctx context.Context, wsID uuid.UUID, req CreateCategoryRequest) (*domain.FinanceCategory, error) {
	var c domain.FinanceCategory
	err := s.db.QueryRow(ctx, `
		INSERT INTO finance_categories (workspace_id, name, type, color, icon, parent_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, workspace_id, name, type, color, icon, parent_id, is_system, created_at
	`, wsID, req.Name, req.Type, req.Color, req.Icon, req.ParentID).Scan(
		&c.ID, &c.WorkspaceID, &c.Name, &c.Type, &c.Color, &c.Icon,
		&c.ParentID, &c.IsSystem, &c.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *FinanceService) UpdateCategory(ctx context.Context, wsID, id uuid.UUID, req UpdateCategoryRequest) (*domain.FinanceCategory, error) {
	var c domain.FinanceCategory
	err := s.db.QueryRow(ctx, `
		UPDATE finance_categories
		SET name = $3, color = $4, icon = $5, parent_id = $6
		WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
		RETURNING id, workspace_id, name, type, color, icon, parent_id, is_system, created_at
	`, id, wsID, req.Name, req.Color, req.Icon, req.ParentID).Scan(
		&c.ID, &c.WorkspaceID, &c.Name, &c.Type, &c.Color, &c.Icon,
		&c.ParentID, &c.IsSystem, &c.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *FinanceService) DeleteCategory(ctx context.Context, wsID, id uuid.UUID) error {
	_, err := s.db.Exec(ctx, `
		UPDATE finance_categories SET deleted_at = now()
		WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
	`, id, wsID)
	return err
}

// --- Transactions ---

func (s *FinanceService) ListTransactions(ctx context.Context, wsID uuid.UUID, f TransactionFilters) ([]domain.Transaction, error) {
	query := `
		SELECT id, workspace_id, category_id, area_id, type, amount, description,
			   date, is_recurring, recurrence_rule, tags, created_at
		FROM transactions
		WHERE workspace_id = $1 AND deleted_at IS NULL
	`
	args := []any{wsID}
	idx := 2

	if f.Type != nil {
		query += fmt.Sprintf(` AND type = $%d`, idx)
		args = append(args, *f.Type)
		idx++
	}
	if f.CategoryID != nil {
		query += fmt.Sprintf(` AND category_id = $%d`, idx)
		args = append(args, *f.CategoryID)
		idx++
	}
	if f.AreaID != nil {
		query += fmt.Sprintf(` AND area_id = $%d`, idx)
		args = append(args, *f.AreaID)
		idx++
	}
	if f.DateFrom != nil {
		query += fmt.Sprintf(` AND date >= $%d`, idx)
		args = append(args, *f.DateFrom)
		idx++
	}
	if f.DateTo != nil {
		query += fmt.Sprintf(` AND date <= $%d`, idx)
		args = append(args, *f.DateTo)
		idx++
	}
	if f.Tag != nil {
		query += fmt.Sprintf(` AND $%d = ANY(tags)`, idx)
		args = append(args, *f.Tag)
		idx++
	}

	query += ` ORDER BY date DESC`

	limit := f.Limit
	if limit <= 0 {
		limit = 50
	}
	query += fmt.Sprintf(` LIMIT $%d`, idx)
	args = append(args, limit)
	idx++

	if f.Offset > 0 {
		query += fmt.Sprintf(` OFFSET $%d`, idx)
		args = append(args, f.Offset)
	}

	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txns []domain.Transaction
	for rows.Next() {
		var t domain.Transaction
		if err := rows.Scan(&t.ID, &t.WorkspaceID, &t.CategoryID, &t.AreaID, &t.Type,
			&t.Amount, &t.Description, &t.Date, &t.IsRecurring, &t.RecurrenceRule,
			&t.Tags, &t.CreatedAt); err != nil {
			return nil, err
		}
		txns = append(txns, t)
	}
	return txns, nil
}

func (s *FinanceService) CreateTransaction(ctx context.Context, wsID uuid.UUID, limits *domain.EntitlementLimits, req CreateTransactionRequest) (*domain.Transaction, error) {
	counters, err := s.counterSvc.GetCounters(ctx, wsID)
	if err != nil {
		return nil, err
	}
	if !limits.CanCreateTransaction(counters.TransactionsMonthCount) {
		observability.EntitlementLimitReachedTotal.WithLabelValues("transactions").Inc()
		return nil, errors.New("monthly transaction limit reached")
	}

	var t domain.Transaction
	err = s.db.QueryRow(ctx, `
		INSERT INTO transactions (workspace_id, category_id, area_id, type, amount, description, date, is_recurring, recurrence_rule, tags)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, workspace_id, category_id, area_id, type, amount, description,
				  date, is_recurring, recurrence_rule, tags, created_at
	`, wsID, req.CategoryID, req.AreaID, req.Type, req.Amount, req.Description,
		req.Date, req.IsRecurring, req.RecurrenceRule, req.Tags).Scan(
		&t.ID, &t.WorkspaceID, &t.CategoryID, &t.AreaID, &t.Type,
		&t.Amount, &t.Description, &t.Date, &t.IsRecurring, &t.RecurrenceRule,
		&t.Tags, &t.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	_ = s.counterSvc.IncrementTransactionsMonth(ctx, wsID)
	return &t, nil
}

func (s *FinanceService) UpdateTransaction(ctx context.Context, wsID, id uuid.UUID, req UpdateTransactionRequest) (*domain.Transaction, error) {
	var t domain.Transaction
	err := s.db.QueryRow(ctx, `
		UPDATE transactions
		SET category_id = $3, area_id = $4, type = $5, amount = $6, description = $7,
			date = $8, is_recurring = $9, recurrence_rule = $10, tags = $11
		WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
		RETURNING id, workspace_id, category_id, area_id, type, amount, description,
				  date, is_recurring, recurrence_rule, tags, created_at
	`, id, wsID, req.CategoryID, req.AreaID, req.Type, req.Amount, req.Description,
		req.Date, req.IsRecurring, req.RecurrenceRule, req.Tags).Scan(
		&t.ID, &t.WorkspaceID, &t.CategoryID, &t.AreaID, &t.Type,
		&t.Amount, &t.Description, &t.Date, &t.IsRecurring, &t.RecurrenceRule,
		&t.Tags, &t.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *FinanceService) DeleteTransaction(ctx context.Context, wsID, id uuid.UUID) error {
	_, err := s.db.Exec(ctx, `
		UPDATE transactions SET deleted_at = now()
		WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
	`, id, wsID)
	return err
}

// --- Budgets ---

func (s *FinanceService) ListBudgets(ctx context.Context, wsID uuid.UUID, month string) ([]domain.Budget, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, workspace_id, category_id, month, amount, created_at, updated_at
		FROM budgets
		WHERE workspace_id = $1 AND month = $2
	`, wsID, month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var budgets []domain.Budget
	for rows.Next() {
		var b domain.Budget
		if err := rows.Scan(&b.ID, &b.WorkspaceID, &b.CategoryID, &b.Month,
			&b.Amount, &b.CreatedAt, &b.UpdatedAt); err != nil {
			return nil, err
		}
		budgets = append(budgets, b)
	}
	return budgets, nil
}

func (s *FinanceService) UpsertBudget(ctx context.Context, wsID uuid.UUID, req UpsertBudgetRequest) (*domain.Budget, error) {
	var b domain.Budget
	err := s.db.QueryRow(ctx, `
		INSERT INTO budgets (workspace_id, category_id, month, amount)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (workspace_id, COALESCE(category_id, '00000000-0000-0000-0000-000000000000'::uuid), month)
		DO UPDATE SET amount = EXCLUDED.amount, updated_at = now()
		RETURNING id, workspace_id, category_id, month, amount, created_at, updated_at
	`, wsID, req.CategoryID, req.Month, req.Amount).Scan(
		&b.ID, &b.WorkspaceID, &b.CategoryID, &b.Month, &b.Amount, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (s *FinanceService) DeleteBudget(ctx context.Context, wsID, id uuid.UUID) error {
	_, err := s.db.Exec(ctx, `
		DELETE FROM budgets WHERE id = $1 AND workspace_id = $2
	`, id, wsID)
	return err
}

// --- Summary ---

func (s *FinanceService) GetSummary(ctx context.Context, wsID uuid.UUID, month string) (*FinanceSummary, error) {
	summary := &FinanceSummary{Month: month}

	// Totals by type
	err := s.db.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(amount) FILTER (WHERE type = 'income'), 0),
			COALESCE(SUM(amount) FILTER (WHERE type = 'expense'), 0),
			COALESCE(SUM(amount) FILTER (WHERE type = 'investment'), 0)
		FROM transactions
		WHERE workspace_id = $1 AND to_char(date, 'YYYY-MM') = $2 AND deleted_at IS NULL
	`, wsID, month).Scan(&summary.TotalIncome, &summary.TotalExpenses, &summary.TotalInvestments)
	if err != nil {
		return nil, err
	}
	summary.NetBalance = summary.TotalIncome - summary.TotalExpenses - summary.TotalInvestments

	// Category breakdown
	rows, err := s.db.Query(ctx, `
		SELECT t.category_id, fc.name, t.type, SUM(t.amount), COUNT(*)
		FROM transactions t
		LEFT JOIN finance_categories fc ON fc.id = t.category_id
		WHERE t.workspace_id = $1 AND to_char(t.date, 'YYYY-MM') = $2 AND t.deleted_at IS NULL
		GROUP BY t.category_id, fc.name, t.type
		ORDER BY SUM(t.amount) DESC
	`, wsID, month)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var cb CategoryBreakdown
		if err := rows.Scan(&cb.CategoryID, &cb.CategoryName, &cb.Type, &cb.Total, &cb.Count); err != nil {
			return nil, err
		}
		summary.ByCategory = append(summary.ByCategory, cb)
	}

	// Budget comparison
	bRows, err := s.db.Query(ctx, `
		SELECT b.category_id, fc.name, b.amount,
			   COALESCE(SUM(t.amount) FILTER (WHERE t.type = 'expense'), 0)
		FROM budgets b
		LEFT JOIN finance_categories fc ON fc.id = b.category_id
		LEFT JOIN transactions t ON t.category_id = b.category_id
			AND t.workspace_id = b.workspace_id
			AND to_char(t.date, 'YYYY-MM') = b.month
			AND t.deleted_at IS NULL
		WHERE b.workspace_id = $1 AND b.month = $2
		GROUP BY b.id, b.category_id, fc.name, b.amount
	`, wsID, month)
	if err != nil {
		return nil, err
	}
	defer bRows.Close()

	for bRows.Next() {
		var bc BudgetComparison
		if err := bRows.Scan(&bc.CategoryID, &bc.CategoryName, &bc.BudgetAmount, &bc.SpentAmount); err != nil {
			return nil, err
		}
		bc.Remaining = bc.BudgetAmount - bc.SpentAmount
		if bc.BudgetAmount > 0 {
			bc.Percentage = (bc.SpentAmount / bc.BudgetAmount) * 100
		}
		summary.BudgetStatus = append(summary.BudgetStatus, bc)
	}

	return summary, nil
}
