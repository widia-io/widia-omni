-- name: GetProfile :one
SELECT * FROM user_profiles WHERE id = $1;

-- name: UpdateProfile :one
UPDATE user_profiles
SET display_name = $2, avatar_url = $3, timezone = $4, locale = $5
WHERE id = $1
RETURNING *;

-- name: GetPreferences :one
SELECT * FROM user_preferences WHERE user_id = $1;

-- name: UpdatePreferences :one
UPDATE user_preferences
SET week_starts_on = $2, daily_focus_limit = $3,
    notification_email = $4, notification_push = $5,
    weekly_review_day = $6, weekly_review_time = $7,
    theme = $8, currency = $9, score_weights = $10
WHERE user_id = $1
RETURNING *;

-- name: DeleteAccount :exec
DELETE FROM user_profiles WHERE id = $1;
