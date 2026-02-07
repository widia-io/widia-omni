package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/widia-io/widia-omni/internal/domain"
)

type UserService struct {
	db *pgxpool.Pool
}

func NewUserService(db *pgxpool.Pool) *UserService {
	return &UserService{db: db}
}

func (s *UserService) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.UserProfile, error) {
	var p domain.UserProfile
	err := s.db.QueryRow(ctx, `
		SELECT id, display_name, email, avatar_url, timezone, locale,
			   default_workspace_id, onboarding_completed, created_at, updated_at
		FROM user_profiles WHERE id = $1
	`, userID).Scan(
		&p.ID, &p.DisplayName, &p.Email, &p.AvatarURL, &p.Timezone, &p.Locale,
		&p.DefaultWorkspaceID, &p.OnboardingCompleted, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

type UpdateProfileRequest struct {
	DisplayName string  `json:"display_name"`
	AvatarURL   *string `json:"avatar_url"`
	Timezone    string  `json:"timezone"`
	Locale      string  `json:"locale"`
}

func (s *UserService) UpdateProfile(ctx context.Context, userID uuid.UUID, req UpdateProfileRequest) (*domain.UserProfile, error) {
	var p domain.UserProfile
	err := s.db.QueryRow(ctx, `
		UPDATE user_profiles
		SET display_name = $2, avatar_url = $3, timezone = $4, locale = $5
		WHERE id = $1
		RETURNING id, display_name, email, avatar_url, timezone, locale,
				  default_workspace_id, onboarding_completed, created_at, updated_at
	`, userID, req.DisplayName, req.AvatarURL, req.Timezone, req.Locale).Scan(
		&p.ID, &p.DisplayName, &p.Email, &p.AvatarURL, &p.Timezone, &p.Locale,
		&p.DefaultWorkspaceID, &p.OnboardingCompleted, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (s *UserService) GetPreferences(ctx context.Context, userID uuid.UUID) (*domain.UserPreferences, error) {
	var p domain.UserPreferences
	var reviewTime time.Time
	err := s.db.QueryRow(ctx, `
		SELECT user_id, week_starts_on, daily_focus_limit, notification_email, notification_push,
			   weekly_review_day, weekly_review_time, theme, currency, score_weights, created_at, updated_at
		FROM user_preferences WHERE user_id = $1
	`, userID).Scan(
		&p.UserID, &p.WeekStartsOn, &p.DailyFocusLimit, &p.NotificationEmail, &p.NotificationPush,
		&p.WeeklyReviewDay, &reviewTime, &p.Theme, &p.Currency, &p.ScoreWeights, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	p.WeeklyReviewTime = reviewTime.Format("15:04")
	return &p, nil
}

type UpdatePreferencesRequest struct {
	WeekStartsOn     int16           `json:"week_starts_on"`
	DailyFocusLimit  int16           `json:"daily_focus_limit"`
	NotificationEmail bool           `json:"notification_email"`
	NotificationPush  bool           `json:"notification_push"`
	WeeklyReviewDay   int16          `json:"weekly_review_day"`
	WeeklyReviewTime  string         `json:"weekly_review_time"`
	Theme             string         `json:"theme"`
	Currency          string         `json:"currency"`
	ScoreWeights      json.RawMessage `json:"score_weights"`
}

func (s *UserService) UpdatePreferences(ctx context.Context, userID uuid.UUID, req UpdatePreferencesRequest) (*domain.UserPreferences, error) {
	var p domain.UserPreferences
	var reviewTime time.Time
	err := s.db.QueryRow(ctx, `
		UPDATE user_preferences
		SET week_starts_on = $2, daily_focus_limit = $3, notification_email = $4, notification_push = $5,
			weekly_review_day = $6, weekly_review_time = $7, theme = $8, currency = $9, score_weights = $10
		WHERE user_id = $1
		RETURNING user_id, week_starts_on, daily_focus_limit, notification_email, notification_push,
				  weekly_review_day, weekly_review_time, theme, currency, score_weights, created_at, updated_at
	`, userID, req.WeekStartsOn, req.DailyFocusLimit, req.NotificationEmail, req.NotificationPush,
		req.WeeklyReviewDay, req.WeeklyReviewTime, req.Theme, req.Currency, req.ScoreWeights).Scan(
		&p.UserID, &p.WeekStartsOn, &p.DailyFocusLimit, &p.NotificationEmail, &p.NotificationPush,
		&p.WeeklyReviewDay, &reviewTime, &p.Theme, &p.Currency, &p.ScoreWeights, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	p.WeeklyReviewTime = reviewTime.Format("15:04")
	return &p, nil
}

func (s *UserService) DeleteAccount(ctx context.Context, userID uuid.UUID) error {
	_, err := s.db.Exec(ctx, `DELETE FROM user_profiles WHERE id = $1`, userID)
	return err
}
