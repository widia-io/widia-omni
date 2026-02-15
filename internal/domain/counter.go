package domain

import (
	"time"

	"github.com/google/uuid"
)

type WorkspaceCounter struct {
	WorkspaceID            uuid.UUID `json:"workspace_id"`
	AreasCount             int       `json:"areas_count"`
	GoalsCount             int       `json:"goals_count"`
	HabitsCount            int       `json:"habits_count"`
	ProjectsCount          int       `json:"projects_count"`
	MembersCount           int       `json:"members_count"`
	TasksCreatedToday      int       `json:"tasks_created_today"`
	TasksTodayDate         time.Time `json:"tasks_today_date"`
	TransactionsMonthCount int       `json:"transactions_month_count"`
	TransactionsMonth      string    `json:"transactions_month"`
	StorageBytesUsed       int64     `json:"storage_bytes_used"`
	UpdatedAt              time.Time `json:"updated_at"`
}
