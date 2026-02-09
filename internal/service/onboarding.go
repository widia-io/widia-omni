package service

import (
	"context"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
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
	Areas  bool `json:"areas"`
	Goals  bool `json:"goals"`
	Habits bool `json:"habits"`
}

type OnboardingStatus struct {
	Completed bool           `json:"completed"`
	Steps     OnboardingSteps `json:"steps"`
}

func (s *OnboardingService) GetStatus(ctx context.Context, userID, wsID uuid.UUID) (*OnboardingStatus, error) {
	var completed bool
	err := s.db.QueryRow(ctx, `SELECT onboarding_completed FROM user_profiles WHERE id = $1`, userID).Scan(&completed)
	if err != nil {
		return nil, err
	}

	var areasCount, goalsCount, habitsCount int
	err = s.db.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM life_areas WHERE workspace_id = $1 AND deleted_at IS NULL),
			(SELECT COUNT(*) FROM goals WHERE workspace_id = $1 AND deleted_at IS NULL),
			(SELECT COUNT(*) FROM habits WHERE workspace_id = $1 AND deleted_at IS NULL)
	`, wsID).Scan(&areasCount, &goalsCount, &habitsCount)
	if err != nil {
		return nil, err
	}

	return &OnboardingStatus{
		Completed: completed,
		Steps: OnboardingSteps{
			Areas:  areasCount > 0,
			Goals:  goalsCount > 0,
			Habits: habitsCount > 0,
		},
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

func (s *OnboardingService) SetupAreas(ctx context.Context, wsID uuid.UUID, areas []SetupAreaItem) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	for i, a := range areas {
		slug := a.Slug
		if slug == "" {
			slug = slugify(a.Name)
		}
		order := a.SortOrder
		if order == 0 {
			order = i + 1
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO life_areas (workspace_id, name, slug, icon, color, weight, sort_order)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, wsID, a.Name, slug, a.Icon, a.Color, a.Weight, order)
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
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
		startDate, endDate := g.StartDate, g.EndDate
		if startDate == "" || endDate == "" {
			startDate, endDate = defaultGoalDates(g.Period)
		}
		_, err := tx.Exec(ctx, `
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

func (s *OnboardingService) Complete(ctx context.Context, userID uuid.UUID) error {
	_, err := s.db.Exec(ctx, `UPDATE user_profiles SET onboarding_completed = true WHERE id = $1`, userID)
	return err
}
