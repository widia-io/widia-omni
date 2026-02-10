package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/widia-io/widia-omni/internal/domain"
	"github.com/widia-io/widia-omni/internal/llm"
)

type InsightService struct {
	db  *pgxpool.Pool
	rdb *redis.Client
	llm *llm.Client
}

func NewInsightService(db *pgxpool.Pool, rdb *redis.Client, llmClient *llm.Client) *InsightService {
	return &InsightService{db: db, rdb: rdb, llm: llmClient}
}

type insightData struct {
	Areas       []insightArea       `json:"areas"`
	LifeScore   *insightLifeScore   `json:"life_score"`
	Journal     []insightJournal    `json:"journal"`
	Habits      []insightHabit      `json:"habits"`
	Goals       []insightGoal       `json:"goals"`
	Tasks       insightTasks        `json:"tasks"`
	Finance     *insightFinance     `json:"finance,omitempty"`
}

type insightArea struct {
	Name         string  `json:"name"`
	CurrentScore int16   `json:"current_score"`
	PrevScore    int16   `json:"prev_score"`
	Weight       float64 `json:"weight"`
}

type insightLifeScore struct {
	Current int16 `json:"current"`
	Prev    int16 `json:"prev"`
}

type insightJournal struct {
	Date       string   `json:"date"`
	Mood       *int16   `json:"mood,omitempty"`
	Energy     *int16   `json:"energy,omitempty"`
	Wins       []string `json:"wins,omitempty"`
	Challenges []string `json:"challenges,omitempty"`
}

type insightHabit struct {
	Name         string  `json:"name"`
	AreaName     string  `json:"area_name"`
	Completions  int     `json:"completions"`
	TargetPerWk  int     `json:"target_per_week"`
	AvgIntensity float64 `json:"avg_intensity"`
}

type insightGoal struct {
	Name        string  `json:"name"`
	AreaName    string  `json:"area_name"`
	ProgressPct float64 `json:"progress_pct"`
	DaysLeft    int     `json:"days_left"`
	Status      string  `json:"status"`
}

type insightTasks struct {
	Total     int `json:"total"`
	Completed int `json:"completed"`
	Overdue   int `json:"overdue"`
}

type insightFinance struct {
	TotalIncome  float64 `json:"total_income"`
	TotalExpense float64 `json:"total_expense"`
}

func (s *InsightService) Generate(ctx context.Context, wsID uuid.UUID, insightType domain.InsightType) (*domain.Insight, error) {
	now := time.Now().UTC()
	weekStart := mostRecentMonday(now)

	data, err := s.gatherInsightData(ctx, wsID, weekStart)
	if err != nil {
		return nil, fmt.Errorf("gather data: %w", err)
	}

	messages := s.buildPrompt(data)

	resp, err := s.llm.Complete(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("llm call: %w", err)
	}

	content := s.parseInsightContent(resp.Content)
	contentJSON, err := json.Marshal(content)
	if err != nil {
		return nil, fmt.Errorf("marshal content: %w", err)
	}

	var insight domain.Insight
	err = s.db.QueryRow(ctx, `
		INSERT INTO workspace_insights (workspace_id, type, week_start, content, model, prompt_tokens, completion_tokens)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, workspace_id, type, week_start, content, model, prompt_tokens, completion_tokens, created_at
	`, wsID, insightType, weekStart, contentJSON, s.llm.Model(), resp.PromptTokens, resp.CompletionTokens).Scan(
		&insight.ID, &insight.WorkspaceID, &insight.Type, &insight.WeekStart,
		&insight.Content, &insight.Model, &insight.PromptTokens, &insight.CompletionTokens, &insight.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert insight: %w", err)
	}

	if s.rdb != nil {
		cooldownKey := fmt.Sprintf("ws:%s:insight:cooldown", wsID.String())
		s.rdb.Set(ctx, cooldownKey, "1", 24*time.Hour)

		latestKey := fmt.Sprintf("ws:%s:insight:latest", wsID.String())
		s.rdb.Del(ctx, latestKey)
	}

	return &insight, nil
}

func (s *InsightService) List(ctx context.Context, wsID uuid.UUID, limit, offset int) ([]domain.Insight, error) {
	if limit <= 0 || limit > 50 {
		limit = 10
	}

	rows, err := s.db.Query(ctx, `
		SELECT id, workspace_id, type, week_start, content, model, prompt_tokens, completion_tokens, created_at
		FROM workspace_insights
		WHERE workspace_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, wsID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var insights []domain.Insight
	for rows.Next() {
		var i domain.Insight
		if err := rows.Scan(&i.ID, &i.WorkspaceID, &i.Type, &i.WeekStart,
			&i.Content, &i.Model, &i.PromptTokens, &i.CompletionTokens, &i.CreatedAt); err != nil {
			return nil, err
		}
		insights = append(insights, i)
	}
	return insights, nil
}

func (s *InsightService) GetLatest(ctx context.Context, wsID uuid.UUID) (*domain.Insight, error) {
	cacheKey := fmt.Sprintf("ws:%s:insight:latest", wsID.String())
	if s.rdb != nil {
		cached, err := s.rdb.Get(ctx, cacheKey).Bytes()
		if err == nil {
			var i domain.Insight
			if json.Unmarshal(cached, &i) == nil {
				return &i, nil
			}
		}
	}

	var i domain.Insight
	err := s.db.QueryRow(ctx, `
		SELECT id, workspace_id, type, week_start, content, model, prompt_tokens, completion_tokens, created_at
		FROM workspace_insights
		WHERE workspace_id = $1
		ORDER BY created_at DESC LIMIT 1
	`, wsID).Scan(&i.ID, &i.WorkspaceID, &i.Type, &i.WeekStart,
		&i.Content, &i.Model, &i.PromptTokens, &i.CompletionTokens, &i.CreatedAt)
	if err != nil {
		return nil, err
	}

	if s.rdb != nil {
		data, _ := json.Marshal(i)
		s.rdb.Set(ctx, cacheKey, data, 1*time.Hour)
	}

	return &i, nil
}

func (s *InsightService) CanGenerate(ctx context.Context, wsID uuid.UUID) (bool, error) {
	if s.rdb == nil {
		return true, nil
	}
	cooldownKey := fmt.Sprintf("ws:%s:insight:cooldown", wsID.String())
	exists, err := s.rdb.Exists(ctx, cooldownKey).Result()
	if err != nil {
		return true, nil
	}
	return exists == 0, nil
}

// --- data gathering ---

func (s *InsightService) gatherInsightData(ctx context.Context, wsID uuid.UUID, weekStart time.Time) (*insightData, error) {
	weekEnd := weekStart.AddDate(0, 0, 7)
	prevWeekStart := weekStart.AddDate(0, 0, -7)

	data := &insightData{}

	// Areas + area scores (current + previous)
	areaRows, err := s.db.Query(ctx, `
		SELECT la.name, la.weight,
			COALESCE((SELECT score FROM area_scores WHERE workspace_id = $1 AND area_id = la.id AND week_start = $2), 0),
			COALESCE((SELECT score FROM area_scores WHERE workspace_id = $1 AND area_id = la.id AND week_start = $3), 0)
		FROM life_areas la
		WHERE la.workspace_id = $1 AND la.deleted_at IS NULL AND la.is_active = true
		ORDER BY la.sort_order
	`, wsID, weekStart, prevWeekStart)
	if err != nil {
		return nil, err
	}
	defer areaRows.Close()
	for areaRows.Next() {
		var a insightArea
		if err := areaRows.Scan(&a.Name, &a.Weight, &a.CurrentScore, &a.PrevScore); err != nil {
			return nil, err
		}
		data.Areas = append(data.Areas, a)
	}

	// Life score (current + previous)
	var curScore, prevScore int16
	s.db.QueryRow(ctx, `SELECT COALESCE(score, 0) FROM life_scores WHERE workspace_id = $1 AND week_start = $2`, wsID, weekStart).Scan(&curScore)
	s.db.QueryRow(ctx, `SELECT COALESCE(score, 0) FROM life_scores WHERE workspace_id = $1 AND week_start = $2`, wsID, prevWeekStart).Scan(&prevScore)
	data.LifeScore = &insightLifeScore{Current: curScore, Prev: prevScore}

	// Journal entries
	journalRows, err := s.db.Query(ctx, `
		SELECT date, mood, energy, wins, challenges
		FROM journal_entries
		WHERE workspace_id = $1 AND date >= $2 AND date < $3
		ORDER BY date
	`, wsID, weekStart, weekEnd)
	if err != nil {
		return nil, err
	}
	defer journalRows.Close()
	for journalRows.Next() {
		var j insightJournal
		var d time.Time
		if err := journalRows.Scan(&d, &j.Mood, &j.Energy, &j.Wins, &j.Challenges); err != nil {
			return nil, err
		}
		j.Date = d.Format("2006-01-02")
		data.Journal = append(data.Journal, j)
	}

	// Habits aggregated
	habitRows, err := s.db.Query(ctx, `
		SELECT h.name, la.name AS area_name, h.target_per_week,
			COUNT(he.id) AS completions,
			COALESCE(AVG(he.intensity), 0) AS avg_intensity
		FROM habits h
		JOIN life_areas la ON la.id = h.area_id
		LEFT JOIN habit_entries he ON he.habit_id = h.id AND he.date >= $2 AND he.date < $3
		WHERE h.workspace_id = $1 AND h.deleted_at IS NULL AND h.is_active = true
		GROUP BY h.id, h.name, la.name, h.target_per_week
	`, wsID, weekStart, weekEnd)
	if err != nil {
		return nil, err
	}
	defer habitRows.Close()
	for habitRows.Next() {
		var h insightHabit
		if err := habitRows.Scan(&h.Name, &h.AreaName, &h.TargetPerWk, &h.Completions, &h.AvgIntensity); err != nil {
			return nil, err
		}
		data.Habits = append(data.Habits, h)
	}

	// Goals
	goalRows, err := s.db.Query(ctx, `
		SELECT g.title, la.name AS area_name, g.status,
			CASE WHEN g.target_value > 0 THEN LEAST(g.current_value / g.target_value * 100, 100) ELSE 0 END AS progress_pct,
			GREATEST(g.end_date - CURRENT_DATE, 0)::int AS days_left
		FROM goals g
		JOIN life_areas la ON la.id = g.area_id
		WHERE g.workspace_id = $1 AND g.deleted_at IS NULL AND g.status NOT IN ('cancelled')
		ORDER BY g.end_date
	`, wsID)
	if err != nil {
		return nil, err
	}
	defer goalRows.Close()
	for goalRows.Next() {
		var g insightGoal
		if err := goalRows.Scan(&g.Name, &g.AreaName, &g.Status, &g.ProgressPct, &g.DaysLeft); err != nil {
			return nil, err
		}
		data.Goals = append(data.Goals, g)
	}

	// Tasks
	s.db.QueryRow(ctx, `
		SELECT
			COUNT(*),
			COUNT(*) FILTER (WHERE is_completed = true),
			COUNT(*) FILTER (WHERE is_completed = false AND due_date < CURRENT_DATE)
		FROM tasks
		WHERE workspace_id = $1 AND deleted_at IS NULL AND created_at >= $2 AND created_at < $3
	`, wsID, weekStart, weekEnd).Scan(&data.Tasks.Total, &data.Tasks.Completed, &data.Tasks.Overdue)

	// Finance (conditional)
	var financeEnabled bool
	s.db.QueryRow(ctx, `
		SELECT COALESCE((we.limits->>'finance_enabled')::boolean, false)
		FROM workspace_entitlements we
		WHERE we.workspace_id = $1 AND we.is_current = true
	`, wsID).Scan(&financeEnabled)

	if financeEnabled {
		var fin insightFinance
		s.db.QueryRow(ctx, `
			SELECT
				COALESCE(SUM(CASE WHEN type = 'income' THEN amount ELSE 0 END), 0),
				COALESCE(SUM(CASE WHEN type = 'expense' THEN amount ELSE 0 END), 0)
			FROM finance_transactions
			WHERE workspace_id = $1 AND date >= $2 AND date < $3
		`, wsID, weekStart, weekEnd).Scan(&fin.TotalIncome, &fin.TotalExpense)
		data.Finance = &fin
	}

	return data, nil
}

// --- prompt construction ---

func (s *InsightService) buildPrompt(data *insightData) []llm.ChatMessage {
	systemPrompt := `You are a personal life coach AI analyzing a user's weekly data from their life management platform.
Output ONLY valid JSON matching this exact schema (no markdown, no explanation):

{
  "summary": "~200 word weekly narrative",
  "highlights": ["positive achievement 1", "..."],
  "concerns": ["declining metric or risk 1", "..."],
  "patterns": {
    "best_habit_days": ["Monday", "..."],
    "mood_trend": "improving|stable|declining",
    "energy_trend": "improving|stable|declining",
    "task_completion_pct": 85.0,
    "habit_consistency_pct": 70.0
  },
  "correlations": {
    "mood_productivity": "description of mood-productivity link",
    "habit_score": "description of habit-score relationship"
  },
  "recommendations": [
    {"area": "area name", "action": "specific actionable suggestion", "priority": "high|medium|low"}
  ],
  "area_breakdown": [
    {"area_name": "name", "score_delta": 5.0, "summary": "brief", "suggestion": "actionable tip"}
  ]
}

Guidelines:
- Be encouraging but honest
- Focus on actionable, specific recommendations
- Reference actual numbers from the data
- If data is sparse, note it and provide general advice
- Keep highlights and concerns to 3-5 items each
- Provide 2-4 recommendations prioritized by impact`

	var sb strings.Builder
	sb.WriteString("Here is my weekly data:\n\n")

	// Life score
	if data.LifeScore != nil {
		sb.WriteString(fmt.Sprintf("## Life Score\nCurrent: %d | Previous: %d\n\n", data.LifeScore.Current, data.LifeScore.Prev))
	}

	// Areas
	if len(data.Areas) > 0 {
		sb.WriteString("## Life Areas\n")
		for _, a := range data.Areas {
			delta := int(a.CurrentScore) - int(a.PrevScore)
			sign := "+"
			if delta < 0 {
				sign = ""
			}
			sb.WriteString(fmt.Sprintf("- %s: %d (was %d, %s%d) weight=%.0f%%\n", a.Name, a.CurrentScore, a.PrevScore, sign, delta, a.Weight*100))
		}
		sb.WriteString("\n")
	}

	// Journal
	if len(data.Journal) > 0 {
		sb.WriteString("## Journal Entries\n")
		for _, j := range data.Journal {
			sb.WriteString(fmt.Sprintf("- %s:", j.Date))
			if j.Mood != nil {
				sb.WriteString(fmt.Sprintf(" mood=%d/5", *j.Mood))
			}
			if j.Energy != nil {
				sb.WriteString(fmt.Sprintf(" energy=%d/5", *j.Energy))
			}
			if len(j.Wins) > 0 {
				sb.WriteString(fmt.Sprintf(" wins=[%s]", strings.Join(j.Wins, ", ")))
			}
			if len(j.Challenges) > 0 {
				sb.WriteString(fmt.Sprintf(" challenges=[%s]", strings.Join(j.Challenges, ", ")))
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	// Habits
	if len(data.Habits) > 0 {
		sb.WriteString("## Habits\n")
		for _, h := range data.Habits {
			sb.WriteString(fmt.Sprintf("- %s (%s): %d/%d completions, avg_intensity=%.1f\n",
				h.Name, h.AreaName, h.Completions, h.TargetPerWk, h.AvgIntensity))
		}
		sb.WriteString("\n")
	}

	// Goals
	if len(data.Goals) > 0 {
		sb.WriteString("## Goals\n")
		for _, g := range data.Goals {
			sb.WriteString(fmt.Sprintf("- %s (%s): %.0f%% progress, %d days left, status=%s\n",
				g.Name, g.AreaName, g.ProgressPct, g.DaysLeft, g.Status))
		}
		sb.WriteString("\n")
	}

	// Tasks
	sb.WriteString(fmt.Sprintf("## Tasks\nTotal: %d | Completed: %d | Overdue: %d\n\n", data.Tasks.Total, data.Tasks.Completed, data.Tasks.Overdue))

	// Finance
	if data.Finance != nil {
		sb.WriteString(fmt.Sprintf("## Finance\nIncome: $%.2f | Expenses: $%.2f | Net: $%.2f\n\n",
			data.Finance.TotalIncome, data.Finance.TotalExpense, data.Finance.TotalIncome-data.Finance.TotalExpense))
	}

	return []llm.ChatMessage{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: sb.String()},
	}
}

// --- response parsing ---

func (s *InsightService) parseInsightContent(raw string) *domain.InsightContent {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var content domain.InsightContent
	if err := json.Unmarshal([]byte(raw), &content); err != nil {
		return &domain.InsightContent{
			Summary: raw,
		}
	}
	return &content
}

// --- helpers ---

func mostRecentMonday(t time.Time) time.Time {
	offset := (int(t.Weekday()) - int(time.Monday) + 7) % 7
	monday := t.AddDate(0, 0, -offset)
	return time.Date(monday.Year(), monday.Month(), monday.Day(), 0, 0, 0, 0, time.UTC)
}
