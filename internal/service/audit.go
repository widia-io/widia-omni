package service

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditService struct {
	db *pgxpool.Pool
}

func NewAuditService(db *pgxpool.Pool) *AuditService {
	return &AuditService{db: db}
}

func (s *AuditService) Log(ctx context.Context, wsID, userID uuid.UUID, action, entityType string, entityID *uuid.UUID, metadata json.RawMessage, ip, ua *string) error {
	_, err := s.db.Exec(ctx, `
		SELECT insert_audit_log($1, $2, $3, $4, $5, $6, $7::inet, $8)
	`, wsID, userID, action, entityType, entityID, metadata, ip, ua)
	return err
}

func (s *AuditService) LogAction(ctx context.Context, wsID, userID uuid.UUID, action string, metadata map[string]any, ip, ua *string) error {
	raw, err := marshalAuditMetadata(metadata)
	if err != nil {
		return err
	}
	return s.Log(ctx, wsID, userID, action, "onboarding", nil, raw, ip, ua)
}

func (s *AuditService) LogActionOnce(ctx context.Context, wsID, userID uuid.UUID, action string, metadata map[string]any, ip, ua *string) error {
	raw, err := marshalAuditMetadata(metadata)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(ctx, `
		INSERT INTO audit_log (workspace_id, user_id, action, entity_type, entity_id, metadata, ip_address, user_agent)
		SELECT $1, $2, $3, 'onboarding', NULL, $4, $5::inet, $6
		WHERE NOT EXISTS (
			SELECT 1
			FROM audit_log
			WHERE workspace_id = $1 AND user_id = $2 AND action = $3
		)
	`, wsID, userID, action, raw, ip, ua)
	return err
}

func marshalAuditMetadata(metadata map[string]any) (json.RawMessage, error) {
	if len(metadata) == 0 {
		return nil, nil
	}

	raw, err := json.Marshal(metadata)
	if err != nil {
		return nil, err
	}
	return raw, nil
}
