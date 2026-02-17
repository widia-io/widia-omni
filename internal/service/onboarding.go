package service

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/widia-io/widia-omni/internal/domain"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type OnboardingService struct {
	db *pgxpool.Pool
}

func NewOnboardingService(db *pgxpool.Pool) *OnboardingService {
	return &OnboardingService{db: db}
}

type OnboardingSteps struct {
	Areas     bool `json:"areas"`
	Goals     bool `json:"goals"`
	Habits    bool `json:"habits"`
	Project   bool `json:"project"`
	FirstTask bool `json:"first_task"`
}

type OnboardingStatus struct {
	Completed   bool            `json:"completed"`
	HabitsState string          `json:"habits_state"`
	Steps       OnboardingSteps `json:"steps"`
}

const (
	HabitsStatePending   = "pending"
	HabitsStateCompleted = "completed"
	HabitsStateSkipped   = "skipped"
)

var (
	ErrOnboardingGoalAreaRequired = errors.New("goal area_id is required")
	ErrOnboardingAreaNotFound     = errors.New("onboarding area not found")
)

type OnboardingIncompleteError struct {
	MissingSteps []string
}

func (e *OnboardingIncompleteError) Error() string {
	return "onboarding incomplete"
}

func (s *OnboardingService) GetStatus(ctx context.Context, userID, wsID uuid.UUID) (*OnboardingStatus, error) {
	var completed bool
	err := s.db.QueryRow(ctx, `SELECT onboarding_completed FROM user_profiles WHERE id = $1`, userID).Scan(&completed)
	if err != nil {
		return nil, err
	}

	var areasCount, goalsCount, habitsCount, projectsCount, firstTaskCount int
	err = s.db.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM life_areas WHERE workspace_id = $1 AND deleted_at IS NULL),
			(SELECT COUNT(*) FROM goals WHERE workspace_id = $1 AND deleted_at IS NULL),
			(SELECT COUNT(*) FROM habits WHERE workspace_id = $1 AND deleted_at IS NULL),
			(SELECT COUNT(*) FROM projects WHERE workspace_id = $1 AND deleted_at IS NULL),
			(SELECT COUNT(*) FROM tasks WHERE workspace_id = $1 AND deleted_at IS NULL AND project_id IS NOT NULL)
	`, wsID).Scan(&areasCount, &goalsCount, &habitsCount, &projectsCount, &firstTaskCount)
	if err != nil {
		return nil, err
	}

	var habitsSkipped bool
	err = s.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM audit_log
			WHERE workspace_id = $1
			  AND user_id = $2
			  AND action = 'onboarding_habits_skipped'
		)
	`, wsID, userID).Scan(&habitsSkipped)
	if err != nil {
		return nil, err
	}

	habitsState := HabitsStatePending
	if habitsCount > 0 {
		habitsState = HabitsStateCompleted
	} else if habitsSkipped {
		habitsState = HabitsStateSkipped
	}

	steps := OnboardingSteps{
		Areas:     areasCount > 0,
		Goals:     goalsCount > 0,
		Habits:    habitsState != HabitsStatePending,
		Project:   projectsCount > 0,
		FirstTask: firstTaskCount > 0,
	}
	if completed {
		steps = OnboardingSteps{
			Areas: true, Goals: true, Habits: true, Project: true, FirstTask: true,
		}
		if habitsState == HabitsStatePending {
			habitsState = HabitsStateCompleted
		}
	}

	return &OnboardingStatus{
		Completed:   completed,
		HabitsState: habitsState,
		Steps:       steps,
	}, nil
}

type SetupAreaItem struct {
	Name      string  `json:"name"`
	Slug      string  `json:"slug"`
	Icon      string  `json:"icon"`
	Color     string  `json:"color"`
	Weight    float64 `json:"weight"`
	SortOrder int     `json:"sort_order"`
}

func slugify(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	result, _, _ := transform.String(t, strings.ToLower(s))
	var b strings.Builder
	for _, r := range result {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			b.WriteRune(r)
		} else if r == ' ' || r == '-' {
			b.WriteByte('-')
		}
	}
	return strings.Trim(b.String(), "-")
}

func (s *OnboardingService) SetupAreas(ctx context.Context, wsID uuid.UUID, areas []SetupAreaItem) ([]domain.LifeArea, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	created := make([]domain.LifeArea, 0, len(areas))
	for i, a := range areas {
		slug := a.Slug
		if slug == "" {
			slug = slugify(a.Name)
		}
		order := a.SortOrder
		if order == 0 {
			order = i + 1
		}
		var area domain.LifeArea
		err := tx.QueryRow(ctx, `
			INSERT INTO life_areas (workspace_id, name, slug, icon, color, weight, sort_order)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, workspace_id, name, slug, icon, color, weight, sort_order, is_active,
			          created_at, updated_at
		`, wsID, a.Name, slug, a.Icon, a.Color, a.Weight, order).Scan(
			&area.ID, &area.WorkspaceID, &area.Name, &area.Slug, &area.Icon, &area.Color,
			&area.Weight, &area.SortOrder, &area.IsActive, &area.CreatedAt, &area.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		created = append(created, area)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return created, nil
}

type SetupGoalItem struct {
	AreaID    *uuid.UUID `json:"area_id"`
	Title     string     `json:"title"`
	Period    string     `json:"period"`
	StartDate string     `json:"start_date"`
	EndDate   string     `json:"end_date"`
}

func defaultGoalDates(period string) (string, string) {
	now := time.Now()
	switch period {
	case "yearly":
		return time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location()).Format("2006-01-02"),
			time.Date(now.Year(), 12, 31, 0, 0, 0, 0, now.Location()).Format("2006-01-02")
	case "monthly":
		start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		return start.Format("2006-01-02"),
			start.AddDate(0, 1, -1).Format("2006-01-02")
	case "weekly":
		weekday := int(now.Weekday())
		if weekday == 0 {
			weekday = 7
		}
		start := now.AddDate(0, 0, -(weekday - 1))
		return start.Format("2006-01-02"),
			start.AddDate(0, 0, 6).Format("2006-01-02")
	default: // quarterly
		q := (int(now.Month()) - 1) / 3
		start := time.Date(now.Year(), time.Month(q*3+1), 1, 0, 0, 0, 0, now.Location())
		return start.Format("2006-01-02"),
			start.AddDate(0, 3, -1).Format("2006-01-02")
	}
}

func (s *OnboardingService) SetupGoals(ctx context.Context, wsID uuid.UUID, goals []SetupGoalItem) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, g := range goals {
		if g.AreaID == nil {
			return ErrOnboardingGoalAreaRequired
		}
		exists, err := areaExists(ctx, tx, wsID, *g.AreaID)
		if err != nil {
			return err
		}
		if !exists {
			return ErrOnboardingAreaNotFound
		}

		startDate, endDate := g.StartDate, g.EndDate
		if startDate == "" || endDate == "" {
			startDate, endDate = defaultGoalDates(g.Period)
		}
		_, err = tx.Exec(ctx, `
				INSERT INTO goals (workspace_id, area_id, title, period, start_date, end_date)
				VALUES ($1, $2, $3, $4, $5, $6)
			`, wsID, g.AreaID, g.Title, g.Period, startDate, endDate)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

type SetupHabitItem struct {
	AreaID        *uuid.UUID `json:"area_id"`
	Name          string     `json:"name"`
	Color         string     `json:"color"`
	Frequency     string     `json:"frequency"`
	TargetPerWeek int        `json:"target_per_week"`
}

func (s *OnboardingService) SetupHabits(ctx context.Context, wsID uuid.UUID, habits []SetupHabitItem) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for _, h := range habits {
		color := h.Color
		if color == "" {
			color = "#788c5d"
		}
		freq := h.Frequency
		if freq == "" {
			freq = "daily"
		}
		tpw := h.TargetPerWeek
		if tpw == 0 {
			tpw = 3
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO habits (workspace_id, area_id, name, color, frequency, target_per_week)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, wsID, h.AreaID, h.Name, color, freq, tpw)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (s *OnboardingService) Complete(ctx context.Context, userID, wsID uuid.UUID) error {
	status, err := s.GetStatus(ctx, userID, wsID)
	if err != nil {
		return err
	}
	if status.Completed {
		return nil
	}

	missing := make([]string, 0, 5)
	if !status.Steps.Areas {
		missing = append(missing, "areas")
	}
	if !status.Steps.Goals {
		missing = append(missing, "goals")
	}
	if !status.Steps.Habits {
		missing = append(missing, "habits")
	}
	if !status.Steps.Project {
		missing = append(missing, "project")
	}
	if !status.Steps.FirstTask {
		missing = append(missing, "first_task")
	}
	if len(missing) > 0 {
		return &OnboardingIncompleteError{MissingSteps: missing}
	}

	_, err = s.db.Exec(ctx, `UPDATE user_profiles SET onboarding_completed = true WHERE id = $1`, userID)
	return err
}

func areaExists(ctx context.Context, tx pgx.Tx, wsID, areaID uuid.UUID) (bool, error) {
	var exists bool
	err := tx.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1
			FROM life_areas
			WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
		)
	`, areaID, wsID).Scan(&exists)
	return exists, err
}
