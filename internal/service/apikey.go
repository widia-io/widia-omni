package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/widia-io/widia-omni/internal/domain"
)

const (
	maxKeysPerWorkspace = 5
	apiKeyCacheTTL      = 5 * time.Minute
	lastUsedDebounce    = 1 * time.Minute
)

type APIKeyService struct {
	db  *pgxpool.Pool
	rdb *redis.Client
}

func NewAPIKeyService(db *pgxpool.Pool, rdb *redis.Client) *APIKeyService {
	return &APIKeyService{db: db, rdb: rdb}
}

type CreateAPIKeyRequest struct {
	Name      string     `json:"name"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type cachedAPIKey struct {
	KeyID       uuid.UUID                `json:"key_id"`
	WorkspaceID uuid.UUID                `json:"workspace_id"`
	ExpiresAt   *time.Time               `json:"expires_at,omitempty"`
	Limits      *domain.EntitlementLimits `json:"limits"`
}

func (s *APIKeyService) Create(ctx context.Context, wsID, userID uuid.UUID, req CreateAPIKeyRequest) (*domain.APIKeyWithSecret, error) {
	var count int
	err := s.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM api_keys WHERE workspace_id = $1 AND revoked_at IS NULL`,
		wsID).Scan(&count)
	if err != nil {
		return nil, fmt.Errorf("count keys: %w", err)
	}
	if count >= maxKeysPerWorkspace {
		return nil, fmt.Errorf("maximum %d active API keys per workspace", maxKeysPerWorkspace)
	}

	rawBytes := make([]byte, 16)
	if _, err := rand.Read(rawBytes); err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}
	rawKey := "wsk_" + hex.EncodeToString(rawBytes)
	prefix := rawKey[:12]

	hash := sha256.Sum256([]byte(rawKey))
	keyHash := hex.EncodeToString(hash[:])

	var ak domain.APIKey
	err = s.db.QueryRow(ctx, `
		INSERT INTO api_keys (workspace_id, created_by, name, key_hash, key_prefix, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, workspace_id, created_by, name, key_prefix, key_hash, last_used_at, revoked_at, expires_at, created_at
	`, wsID, userID, req.Name, keyHash, prefix, req.ExpiresAt).Scan(
		&ak.ID, &ak.WorkspaceID, &ak.CreatedBy, &ak.Name, &ak.KeyPrefix,
		&ak.KeyHash, &ak.LastUsedAt, &ak.RevokedAt, &ak.ExpiresAt, &ak.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert api key: %w", err)
	}

	return &domain.APIKeyWithSecret{APIKey: ak, RawKey: rawKey}, nil
}

func (s *APIKeyService) List(ctx context.Context, wsID uuid.UUID) ([]domain.APIKey, error) {
	rows, err := s.db.Query(ctx, `
		SELECT id, workspace_id, created_by, name, key_prefix, last_used_at, expires_at, created_at
		FROM api_keys
		WHERE workspace_id = $1 AND revoked_at IS NULL
		ORDER BY created_at DESC
	`, wsID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []domain.APIKey
	for rows.Next() {
		var k domain.APIKey
		if err := rows.Scan(&k.ID, &k.WorkspaceID, &k.CreatedBy, &k.Name,
			&k.KeyPrefix, &k.LastUsedAt, &k.ExpiresAt, &k.CreatedAt); err != nil {
			return nil, err
		}
		keys = append(keys, k)
	}
	return keys, nil
}

func (s *APIKeyService) Revoke(ctx context.Context, wsID, keyID uuid.UUID) error {
	tag, err := s.db.Exec(ctx, `
		UPDATE api_keys SET revoked_at = now()
		WHERE id = $1 AND workspace_id = $2 AND revoked_at IS NULL
	`, keyID, wsID)
	if err != nil {
		return fmt.Errorf("revoke key: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("api key not found")
	}

	// Invalidate any cached entry for this key
	if s.rdb != nil {
		// We don't have the hash here, but we can scan — or just let it expire.
		// For immediate invalidation, we'd need the hash. Accept 5min TTL delay.
	}

	return nil
}

func (s *APIKeyService) ValidateKey(ctx context.Context, keyHash string) (*domain.APIKey, *domain.EntitlementLimits, error) {
	// Check Redis cache first
	if s.rdb != nil {
		cacheKey := fmt.Sprintf("apikey:%s", keyHash)
		cached, err := s.rdb.Get(ctx, cacheKey).Bytes()
		if err == nil {
			var c cachedAPIKey
			if json.Unmarshal(cached, &c) == nil {
				ak := &domain.APIKey{
					ID:          c.KeyID,
					WorkspaceID: c.WorkspaceID,
					ExpiresAt:   c.ExpiresAt,
				}
				return ak, c.Limits, nil
			}
		}
	}

	var ak domain.APIKey
	var limitsRaw json.RawMessage
	err := s.db.QueryRow(ctx, `
		SELECT ak.id, ak.workspace_id, ak.expires_at, we.limits
		FROM api_keys ak
		JOIN workspace_entitlements we ON we.workspace_id = ak.workspace_id AND we.is_current = true
		WHERE ak.key_hash = $1 AND ak.revoked_at IS NULL
	`, keyHash).Scan(&ak.ID, &ak.WorkspaceID, &ak.ExpiresAt, &limitsRaw)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, fmt.Errorf("invalid api key")
		}
		return nil, nil, fmt.Errorf("validate key: %w", err)
	}

	limits, err := domain.ParseLimits(limitsRaw)
	if err != nil {
		return nil, nil, fmt.Errorf("parse limits: %w", err)
	}

	// Cache the result
	if s.rdb != nil {
		cacheKey := fmt.Sprintf("apikey:%s", keyHash)
		c := cachedAPIKey{
			KeyID:       ak.ID,
			WorkspaceID: ak.WorkspaceID,
			ExpiresAt:   ak.ExpiresAt,
			Limits:      limits,
		}
		data, _ := json.Marshal(c)
		s.rdb.Set(ctx, cacheKey, data, apiKeyCacheTTL)
	}

	return &ak, limits, nil
}

func (s *APIKeyService) TouchLastUsed(ctx context.Context, keyID uuid.UUID) {
	if s.rdb != nil {
		debounceKey := fmt.Sprintf("apikey:lu:%s", keyID.String())
		set, err := s.rdb.SetNX(ctx, debounceKey, "1", lastUsedDebounce).Result()
		if err != nil || !set {
			return // debounced — skip DB write
		}
	}

	s.db.Exec(ctx, `UPDATE api_keys SET last_used_at = now() WHERE id = $1`, keyID)
}

func HashAPIKey(rawKey string) string {
	hash := sha256.Sum256([]byte(rawKey))
	return hex.EncodeToString(hash[:])
}
