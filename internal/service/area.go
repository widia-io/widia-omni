package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/widia-io/widia-omni/internal/domain"
	"github.com/widia-io/widia-omni/internal/observability"
)

type AreaStats struct {
	GoalsActive            int     `json:"goals_active"`
	GoalsCompleted         int     `json:"goals_completed"`
	ProjectsActive         int     `json:"projects_active"`
	ProjectsCompleted      int     `json:"projects_completed"`
	TasksPending           int     `json:"tasks_pending"`
	TasksCompletedThisWeek int     `json:"tasks_completed_this_week"`
	HabitsActive           int     `json:"habits_active"`
	CurrentStreakAvg       float64 `json:"current_streak_avg"`
	AreaScore              *int16  `json:"area_score"`
}

type AreaSummary struct {
	Area  domain.LifeArea `json:"area"`
	Stats AreaStats       `json:"stats"`
}

type AreaWithStats struct {
	domain.LifeArea
	GoalsCount    int    `json:"goals_count"`
	ProjectsCount int    `json:"projects_count"`
	TasksPending  int    `json:"tasks_pending"`
	AreaScore     *int16 `json:"area_score"`
}

type AreaService struct {
	db         *pgxpool.Pool
	counterSvc *CounterService
	rdb        *redis.Client
}

func NewAreaService(db *pgxpool.Pool, counterSvc *CounterService, rdb *redis.Client) *AreaService {
	return &AreaService{db: db, counterSvc: counterSvc, rdb: rdb}
}

func (s *AreaService) List(ctx context.Context, wsID uuid.UUID) ([]domain.LifeArea, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, workspace_id, name, slug, icon, color, weight, sort_order, is_active,
			   created_at, updated_at
		FROM life_areas
		WHERE workspace_id = $1 AND deleted_at IS NULL
		ORDER BY sort_order
	`, wsID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var areas []domain.LifeArea
	for rows.Next() {
		var a domain.LifeArea
		if err := rows.Scan(&a.ID, &a.WorkspaceID, &a.Name, &a.Slug, &a.Icon, &a.Color,
			&a.Weight, &a.SortOrder, &a.IsActive, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		areas = append(areas, a)
	}
	return areas, nil
}

type CreateAreaRequest struct {
	Name      string  `json:"name"`
	Slug      string  `json:"slug"`
	Icon      string  `json:"icon"`
	Color     string  `json:"color"`
	Weight    float64 `json:"weight"`
	SortOrder int     `json:"sort_order"`
}

func (s *AreaService) Create(ctx context.Context, wsID uuid.UUID, limits *domain.EntitlementLimits, req CreateAreaRequest) (*domain.LifeArea, error) {
	counters, err := s.counterSvc.GetCounters(ctx, wsID)
	if err != nil {
		return nil, err
	}
	if !limits.CanCreateArea(counters.AreasCount) {
		observability.EntitlementLimitReachedTotal.WithLabelValues("areas").Inc()
		return nil, errors.New("area limit reached")
	}

	var a domain.LifeArea
	err = s.db.QueryRow(ctx, `
		INSERT INTO life_areas (workspace_id, name, slug, icon, color, weight, sort_order)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, workspace_id, name, slug, icon, color, weight, sort_order, is_active,
				  created_at, updated_at
	`, wsID, req.Name, req.Slug, req.Icon, req.Color, req.Weight, req.SortOrder).Scan(
		&a.ID, &a.WorkspaceID, &a.Name, &a.Slug, &a.Icon, &a.Color,
		&a.Weight, &a.SortOrder, &a.IsActive, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

type UpdateAreaRequest struct {
	Name     string  `json:"name"`
	Icon     string  `json:"icon"`
	Color    string  `json:"color"`
	Weight   float64 `json:"weight"`
	IsActive bool    `json:"is_active"`
}

func (s *AreaService) Update(ctx context.Context, wsID, id uuid.UUID, req UpdateAreaRequest) (*domain.LifeArea, error) {
	var a domain.LifeArea
	err := s.db.QueryRow(ctx, `
		UPDATE life_areas
		SET name = $3, icon = $4, color = $5, weight = $6, is_active = $7, updated_at = now()
		WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
		RETURNING id, workspace_id, name, slug, icon, color, weight, sort_order, is_active,
				  created_at, updated_at
	`, id, wsID, req.Name, req.Icon, req.Color, req.Weight, req.IsActive).Scan(
		&a.ID, &a.WorkspaceID, &a.Name, &a.Slug, &a.Icon, &a.Color,
		&a.Weight, &a.SortOrder, &a.IsActive, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *AreaService) Delete(ctx context.Context, wsID, id uuid.UUID) error {
	_, err := s.db.Exec(ctx, `
		UPDATE life_areas SET deleted_at = now() WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
	`, id, wsID)
	return err
}

func (s *AreaService) Reorder(ctx context.Context, wsID, id uuid.UUID, sortOrder int) error {
	_, err := s.db.Exec(ctx, `
		UPDATE life_areas SET sort_order = $3, updated_at = now()
		WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
	`, id, wsID, sortOrder)
	return err
}

func (s *AreaService) GetByID(ctx context.Context, wsID, id uuid.UUID) (*domain.LifeArea, error) {
	var a domain.LifeArea
	err := s.db.QueryRow(ctx, `
		SELECT id, workspace_id, name, slug, icon, color, weight, sort_order, is_active,
			   created_at, updated_at
		FROM life_areas
		WHERE id = $1 AND workspace_id = $2 AND deleted_at IS NULL
	`, id, wsID).Scan(&a.ID, &a.WorkspaceID, &a.Name, &a.Slug, &a.Icon, &a.Color,
		&a.Weight, &a.SortOrder, &a.IsActive, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *AreaService) GetSummary(ctx context.Context, wsID, areaID uuid.UUID) (*AreaSummary, error) {
	cacheKey := fmt.Sprintf("ws:%s:area:%s:summary", wsID.String(), areaID.String())

	if s.rdb != nil {
		cached, err := s.rdb.Get(ctx, cacheKey).Bytes()
		if err == nil {
			var summary AreaSummary
			if json.Unmarshal(cached, &summary) == nil {
				return &summary, nil
			}
		}
	}

	area, err := s.GetByID(ctx, wsID, areaID)
	if err != nil {
		return nil, err
	}

	var stats AreaStats
	err = s.db.QueryRow(ctx, `
		SELECT
			(SELECT COUNT(*) FROM goals WHERE workspace_id=$1 AND area_id=$2 AND deleted_at IS NULL AND status NOT IN ('completed','cancelled')),
			(SELECT COUNT(*) FROM goals WHERE workspace_id=$1 AND area_id=$2 AND deleted_at IS NULL AND status = 'completed'),
			(SELECT COUNT(*) FROM projects WHERE workspace_id=$1 AND area_id=$2 AND deleted_at IS NULL AND status IN ('planning','active') AND is_archived = false),
			(SELECT COUNT(*) FROM projects WHERE workspace_id=$1 AND area_id=$2 AND deleted_at IS NULL AND status = 'completed'),
			(SELECT COUNT(*) FROM tasks WHERE workspace_id=$1 AND area_id=$2 AND deleted_at IS NULL AND is_completed = false),
			(SELECT COUNT(*) FROM tasks WHERE workspace_id=$1 AND area_id=$2 AND deleted_at IS NULL AND is_completed = true AND completed_at >= date_trunc('week', CURRENT_DATE)),
			(SELECT COUNT(*) FROM habits WHERE workspace_id=$1 AND area_id=$2 AND deleted_at IS NULL AND is_active = true),
			COALESCE((SELECT AVG(streak_len)::float FROM (
				SELECT COUNT(*) AS streak_len FROM (
					SELECT h.id, he.date,
						   he.date - (ROW_NUMBER() OVER (PARTITION BY h.id ORDER BY he.date))::int AS grp
					FROM habits h
					JOIN habit_entries he ON he.habit_id = h.id
					WHERE h.workspace_id=$1 AND h.area_id=$2 AND h.deleted_at IS NULL
				) d GROUP BY id, grp
				HAVING MAX(date) = CURRENT_DATE
			) cs), 0),
			(SELECT score FROM area_scores WHERE workspace_id=$1 AND area_id=$2 ORDER BY week_start DESC LIMIT 1)
	`, wsID, areaID).Scan(
		&stats.GoalsActive, &stats.GoalsCompleted,
		&stats.ProjectsActive, &stats.ProjectsCompleted,
		&stats.TasksPending, &stats.TasksCompletedThisWeek,
		&stats.HabitsActive, &stats.CurrentStreakAvg, &stats.AreaScore,
	)
	if err != nil {
		return nil, err
	}

	summary := &AreaSummary{Area: *area, Stats: stats}

	if s.rdb != nil {
		data, _ := json.Marshal(summary)
		s.rdb.Set(ctx, cacheKey, data, 2*time.Minute)
	}

	return summary, nil
}

func (s *AreaService) ListWithStats(ctx context.Context, wsID uuid.UUID) ([]AreaWithStats, error) {
	rows, err := s.db.Query(ctx, `
		SELECT a.id, a.workspace_id, a.name, a.slug, a.icon, a.color, a.weight, a.sort_order, a.is_active,
			   a.created_at, a.updated_at,
			   (SELECT COUNT(*) FROM goals g WHERE g.area_id = a.id AND g.deleted_at IS NULL AND g.status NOT IN ('completed','cancelled')),
			   (SELECT COUNT(*) FROM projects p WHERE p.area_id = a.id AND p.deleted_at IS NULL AND p.status IN ('planning','active') AND p.is_archived = false),
			   (SELECT COUNT(*) FROM tasks t WHERE t.area_id = a.id AND t.deleted_at IS NULL AND t.is_completed = false),
			   (SELECT score FROM area_scores s WHERE s.area_id = a.id ORDER BY s.week_start DESC LIMIT 1)
		FROM life_areas a
		WHERE a.workspace_id = $1 AND a.deleted_at IS NULL
		ORDER BY a.sort_order
	`, wsID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var areas []AreaWithStats
	for rows.Next() {
		var a AreaWithStats
		if err := rows.Scan(&a.ID, &a.WorkspaceID, &a.Name, &a.Slug, &a.Icon, &a.Color,
			&a.Weight, &a.SortOrder, &a.IsActive, &a.CreatedAt, &a.UpdatedAt,
			&a.GoalsCount, &a.ProjectsCount, &a.TasksPending, &a.AreaScore); err != nil {
			return nil, err
		}
		areas = append(areas, a)
	}
	return areas, nil
}
