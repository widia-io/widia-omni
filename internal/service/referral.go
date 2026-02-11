package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/widia-io/widia-omni/internal/domain"
)

var ErrReferralCodeNotFound = errors.New("referral code not found")

type ReferralService struct {
	db     *pgxpool.Pool
	appURL string
}

func NewReferralService(db *pgxpool.Pool, appURL string) *ReferralService {
	return &ReferralService{db: db, appURL: strings.TrimRight(appURL, "/")}
}

func (s *ReferralService) GetMe(ctx context.Context, wsID uuid.UUID) (*domain.ReferralMe, error) {
	if err := s.expirePendingAttributions(ctx, wsID); err != nil {
		return nil, err
	}

	code, err := s.ensureCode(ctx, wsID)
	if err != nil {
		return nil, err
	}

	stats := domain.ReferralStats{}
	rows, err := s.db.Query(ctx, `
		SELECT status, COUNT(*)
		FROM referral_attributions
		WHERE referrer_workspace_id = $1
		GROUP BY status
	`, wsID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var st domain.ReferralAttributionStatus
		var count int
		if err := rows.Scan(&st, &count); err != nil {
			return nil, err
		}
		switch st {
		case domain.ReferralAttributionPending:
			stats.Pending = count
		case domain.ReferralAttributionConverted:
			stats.Converted = count
		case domain.ReferralAttributionExpired:
			stats.Expired = count
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	var totalDays int
	var hasAvailable bool
	err = s.db.QueryRow(ctx, `
		SELECT COALESCE(SUM(days), 0), COALESCE(BOOL_OR(status = 'available'), false)
		FROM referral_credits
		WHERE workspace_id = $1
	`, wsID).Scan(&totalDays, &hasAvailable)
	if err != nil {
		return nil, err
	}

	return &domain.ReferralMe{
		Code:         code,
		ShareURL:     fmt.Sprintf("%s/register?ref=%s", s.appURL, code),
		Stats:        stats,
		CreditDays:   totalDays,
		HasAvailable: hasAvailable,
	}, nil
}

func (s *ReferralService) RegenerateCode(ctx context.Context, wsID, actorUserID uuid.UUID) (*domain.ReferralCode, error) {
	var out domain.ReferralCode
	err := s.db.QueryRow(ctx, `
		INSERT INTO workspace_referral_codes (workspace_id, code, created_by, regenerated_at)
		VALUES ($1, generate_referral_code(), $2, now())
		ON CONFLICT (workspace_id) DO UPDATE
		SET code = generate_referral_code(),
			created_by = EXCLUDED.created_by,
			regenerated_at = now()
		RETURNING workspace_id, code, created_by, created_at, regenerated_at
	`, wsID, actorUserID).Scan(&out.WorkspaceID, &out.Code, &out.CreatedBy, &out.CreatedAt, &out.RegeneratedAt)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (s *ReferralService) ListAttributions(ctx context.Context, wsID uuid.UUID, limit, offset int) ([]domain.ReferralAttribution, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	if err := s.expirePendingAttributions(ctx, wsID); err != nil {
		return nil, err
	}

	rows, err := s.db.Query(ctx, `
		SELECT id, referral_code, referrer_workspace_id, referred_workspace_id,
			   referred_user_id, expires_at, status, converted_at, created_at
		FROM referral_attributions
		WHERE referrer_workspace_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, wsID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.ReferralAttribution, 0)
	for rows.Next() {
		var it domain.ReferralAttribution
		if err := rows.Scan(
			&it.ID, &it.ReferralCode, &it.ReferrerWorkspaceID, &it.ReferredWorkspaceID,
			&it.ReferredUserID, &it.ExpiresAt, &it.Status, &it.ConvertedAt, &it.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

func (s *ReferralService) ListCredits(ctx context.Context, wsID uuid.UUID, limit, offset int) ([]domain.ReferralCredit, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := s.db.Query(ctx, `
		SELECT id, attribution_id, workspace_id, credit_type, days, status,
			   expires_at, consumed_at, created_at
		FROM referral_credits
		WHERE workspace_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, wsID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.ReferralCredit, 0)
	for rows.Next() {
		var it domain.ReferralCredit
		if err := rows.Scan(
			&it.ID, &it.AttributionID, &it.WorkspaceID, &it.CreditType, &it.Days, &it.Status,
			&it.ExpiresAt, &it.ConsumedAt, &it.CreatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

func (s *ReferralService) ProcessConversion(ctx context.Context, referredWorkspaceID uuid.UUID) error {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var attrID uuid.UUID
	var referrerWS uuid.UUID
	var expiresAt time.Time
	var status domain.ReferralAttributionStatus
	err = tx.QueryRow(ctx, `
		SELECT id, referrer_workspace_id, expires_at, status
		FROM referral_attributions
		WHERE referred_workspace_id = $1
		FOR UPDATE
	`, referredWorkspaceID).Scan(&attrID, &referrerWS, &expiresAt, &status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}

	if status == domain.ReferralAttributionConverted {
		return tx.Commit(ctx)
	}

	if time.Now().UTC().After(expiresAt) {
		_, err = tx.Exec(ctx, `
			UPDATE referral_attributions
			SET status = 'expired'
			WHERE id = $1 AND status = 'pending'
		`, attrID)
		if err != nil {
			return err
		}
		return tx.Commit(ctx)
	}

	_, err = tx.Exec(ctx, `
		UPDATE referral_attributions
		SET status = 'converted', converted_at = now()
		WHERE id = $1 AND status = 'pending'
	`, attrID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO referral_credits (attribution_id, workspace_id, credit_type, days, status)
		VALUES ($1, $2, 'referrer_bonus', 30, 'available')
		ON CONFLICT (attribution_id, workspace_id, credit_type) DO NOTHING
	`, attrID, referrerWS)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO referral_credits (attribution_id, workspace_id, credit_type, days, status)
		VALUES ($1, $2, 'referred_bonus', 30, 'available')
		ON CONFLICT (attribution_id, workspace_id, credit_type) DO NOTHING
	`, attrID, referredWorkspaceID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *ReferralService) ensureCode(ctx context.Context, wsID uuid.UUID) (string, error) {
	_, err := s.db.Exec(ctx, `
		INSERT INTO workspace_referral_codes (workspace_id, code, created_by)
		SELECT w.id, generate_referral_code(), w.owner_id
		FROM workspaces w
		WHERE w.id = $1
		ON CONFLICT (workspace_id) DO NOTHING
	`, wsID)
	if err != nil {
		return "", err
	}

	var code string
	err = s.db.QueryRow(ctx, `
		SELECT code
		FROM workspace_referral_codes
		WHERE workspace_id = $1
	`, wsID).Scan(&code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrReferralCodeNotFound
		}
		return "", err
	}

	return code, nil
}

func (s *ReferralService) expirePendingAttributions(ctx context.Context, wsID uuid.UUID) error {
	_, err := s.db.Exec(ctx, `
		UPDATE referral_attributions
		SET status = 'expired'
		WHERE referrer_workspace_id = $1
		  AND status = 'pending'
		  AND expires_at < now()
	`, wsID)
	return err
}
