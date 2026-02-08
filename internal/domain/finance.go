package domain

import (
	"time"

	"github.com/google/uuid"
)

type TransactionType string

const (
	TransactionIncome     TransactionType = "income"
	TransactionExpense    TransactionType = "expense"
	TransactionInvestment TransactionType = "investment"
	TransactionTransfer   TransactionType = "transfer"
)

type FinanceCategory struct {
	ID          uuid.UUID       `json:"id"`
	WorkspaceID uuid.UUID      `json:"workspace_id"`
	Name        string          `json:"name"`
	Type        TransactionType `json:"type"`
	Color       *string         `json:"color,omitempty"`
	Icon        *string         `json:"icon,omitempty"`
	ParentID    *uuid.UUID      `json:"parent_id,omitempty"`
	IsSystem    bool            `json:"is_system"`
	CreatedAt   time.Time       `json:"created_at"`
	DeletedAt   *time.Time      `json:"deleted_at,omitempty"`
}

type Transaction struct {
	ID             uuid.UUID       `json:"id"`
	WorkspaceID    uuid.UUID       `json:"workspace_id"`
	CategoryID     *uuid.UUID      `json:"category_id,omitempty"`
	AreaID         *uuid.UUID      `json:"area_id,omitempty"`
	Type           TransactionType `json:"type"`
	Amount         float64         `json:"amount"`
	Description    *string         `json:"description,omitempty"`
	Date           time.Time       `json:"date"`
	IsRecurring    bool            `json:"is_recurring"`
	RecurrenceRule *string         `json:"recurrence_rule,omitempty"`
	Tags           []string        `json:"tags,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	DeletedAt      *time.Time      `json:"deleted_at,omitempty"`
}

type Budget struct {
	ID          uuid.UUID  `json:"id"`
	WorkspaceID uuid.UUID  `json:"workspace_id"`
	CategoryID  *uuid.UUID `json:"category_id,omitempty"`
	Month       string     `json:"month"`
	Amount      float64    `json:"amount"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}
