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
