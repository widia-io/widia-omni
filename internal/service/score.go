package service

import (
	"context"
	"encoding/json"
	"math"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/widia-io/widia-omni/internal/domain"
)

type ScoreService struct {
	db *pgxpool.Pool
}

func NewScoreService(db *pgxpool.Pool) *ScoreService {
	return &ScoreService{db: db}
}

func (s *ScoreService) GetHistory(ctx context.Context, wsID uuid.UUID, weeks int) (*domain.ScoreHistory, error) {
	if weeks <= 0 || weeks > 52 {
		weeks = 4
	}

	lifeRows, err := s.db.Query(ctx, `
		SELECT id, workspace_id, score, week_start, area_scores, created_at
		FROM life_scores
		WHERE workspace_id = $1
		ORDER BY week_start DESC
		LIMIT $2
	`, wsID, weeks)
	if err != nil {
		return nil, err
	}
	defer lifeRows.Close()

	var lifeScores []domain.LifeScore
	for lifeRows.Next() {
		var ls domain.LifeScore
		if err := lifeRows.Scan(&ls.ID, &ls.WorkspaceID, &ls.Score, &ls.WeekStart, &ls.AreaScores, &ls.CreatedAt); err != nil {
			return nil, err
		}
		lifeScores = append(lifeScores, ls)
	}

	areaRows, err := s.db.Query(ctx, `
		SELECT id, workspace_id, area_id, score, week_start, breakdown, created_at
		FROM area_scores
		WHERE workspace_id = $1 AND week_start >= (CURRENT_DATE - make_interval(weeks => $2))::date
		ORDER BY week_start DESC, area_id
	`, wsID, weeks)
	if err != nil {
		return nil, err
	}
	defer areaRows.Close()

	var areaScores []domain.AreaScore
	for areaRows.Next() {
		var as domain.AreaScore
		if err := areaRows.Scan(&as.ID, &as.WorkspaceID, &as.AreaID, &as.Score, &as.WeekStart, &as.Breakdown, &as.CreatedAt); err != nil {
			return nil, err
		}
		areaScores = append(areaScores, as)
	}

	return &domain.ScoreHistory{LifeScores: lifeScores, AreaScores: areaScores}, nil
}

func (s *ScoreService) GetCurrent(ctx context.Context, wsID uuid.UUID) (*domain.LifeScore, error) {
	var ls domain.LifeScore
	err := s.db.QueryRow(ctx, `
		SELECT id, workspace_id, score, week_start, area_scores, created_at
		FROM life_scores
		WHERE workspace_id = $1
		ORDER BY week_start DESC LIMIT 1
	`, wsID).Scan(&ls.ID, &ls.WorkspaceID, &ls.Score, &ls.WeekStart, &ls.AreaScores, &ls.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &ls, nil
}

func (s *ScoreService) Calculate(ctx context.Context, wsID uuid.UUID, weekStart time.Time) error {
	weekEnd := weekStart.AddDate(0, 0, 7)

	rows, err := s.db.Query(ctx, `
		SELECT id, name, weight FROM life_areas
		WHERE workspace_id = $1 AND deleted_at IS NULL AND is_active = true
	`, wsID)
	if err != nil {
		return err
	}
	defer rows.Close()

	type areaInfo struct {
		ID     uuid.UUID
		Name   string
		Weight float64
	}
	var areas []areaInfo
	for rows.Next() {
		var a areaInfo
		if err := rows.Scan(&a.ID, &a.Name, &a.Weight); err != nil {
			return err
		}
		areas = append(areas, a)
	}

	if len(areas) == 0 {
		return nil
	}

	type areaScoreEntry struct {
		AreaID string  `json:"area_id"`
		Name   string  `json:"name"`
		Score  int16   `json:"score"`
		Weight float64 `json:"weight"`
	}
	var areaEntries []areaScoreEntry
	var weightedSum, totalWeight float64

	for _, area := range areas {
		habitScore := s.calcHabitScore(ctx, wsID, area.ID, weekStart, weekEnd)
		goalScore := s.calcGoalScore(ctx, wsID, area.ID)
		taskScore := s.calcTaskScore(ctx, wsID, area.ID, weekStart, weekEnd)

		aScore := habitScore*0.50 + goalScore*0.30 + taskScore*0.20
		capped := int16(math.Min(math.Round(aScore), 100))

		breakdown := domain.ScoreBreakdown{
			HabitsScore: math.Round(habitScore*100) / 100,
			GoalsScore:  math.Round(goalScore*100) / 100,
			TasksScore:  math.Round(taskScore*100) / 100,
		}
		breakdownJSON, _ := json.Marshal(breakdown)

		_, err := s.db.Exec(ctx, `
			INSERT INTO area_scores (workspace_id, area_id, score, week_start, breakdown)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (workspace_id, area_id, week_start)
			DO UPDATE SET score = $3, breakdown = $5
		`, wsID, area.ID, capped, weekStart, breakdownJSON)
		if err != nil {
			return err
		}

		areaEntries = append(areaEntries, areaScoreEntry{
			AreaID: area.ID.String(), Name: area.Name, Score: capped, Weight: area.Weight,
		})
		weightedSum += float64(capped) * area.Weight
		totalWeight += area.Weight
	}

	var lifeScore int16
	if totalWeight > 0 {
		lifeScore = int16(math.Min(math.Round(weightedSum/totalWeight), 100))
	}

	areaScoresJSON, _ := json.Marshal(areaEntries)

	_, err = s.db.Exec(ctx, `
		INSERT INTO life_scores (workspace_id, score, week_start, area_scores)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (workspace_id, week_start)
		DO UPDATE SET score = $2, area_scores = $4
	`, wsID, lifeScore, weekStart, areaScoresJSON)
	return err
}

func (s *ScoreService) calcHabitScore(ctx context.Context, wsID, areaID uuid.UUID, weekStart, weekEnd time.Time) float64 {
	var sumIntensity, sumTarget float64
	err := s.db.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(he.intensity), 0),
			COALESCE(SUM(h.target_per_week * 3), 0)
		FROM habits h
		LEFT JOIN habit_entries he ON he.habit_id = h.id AND he.date >= $3 AND he.date < $4
		WHERE h.workspace_id = $1 AND h.area_id = $2 AND h.deleted_at IS NULL AND h.is_active = true
	`, wsID, areaID, weekStart, weekEnd).Scan(&sumIntensity, &sumTarget)
	if err != nil || sumTarget == 0 {
		return 0
	}
	return math.Min(sumIntensity/sumTarget*100, 100)
}

func (s *ScoreService) calcGoalScore(ctx context.Context, wsID, areaID uuid.UUID) float64 {
	var avgProgress float64
	var count int
	err := s.db.QueryRow(ctx, `
		SELECT COUNT(*),
			   COALESCE(AVG(CASE WHEN target_value > 0 THEN LEAST(current_value / target_value * 100, 100) ELSE 0 END), 0)
		FROM goals
		WHERE workspace_id = $1 AND area_id = $2 AND deleted_at IS NULL
			  AND status NOT IN ('completed', 'cancelled')
	`, wsID, areaID).Scan(&count, &avgProgress)
	if err != nil || count == 0 {
		return 0
	}
	return math.Min(avgProgress, 100)
}

func (s *ScoreService) calcTaskScore(ctx context.Context, wsID, areaID uuid.UUID, weekStart, weekEnd time.Time) float64 {
	var total, completed int
	err := s.db.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE is_completed = true)
		FROM tasks
		WHERE workspace_id = $1 AND area_id = $2 AND deleted_at IS NULL
			  AND created_at >= $3 AND created_at < $4
	`, wsID, areaID, weekStart, weekEnd).Scan(&total, &completed)
	if err != nil || total == 0 {
		return 100
	}
	return math.Min(float64(completed)/float64(total)*100, 100)
}
