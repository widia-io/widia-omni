package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/hibiken/asynq"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/widia-io/widia-omni/internal/domain"
)

var (
	ErrWorkspaceNotFound     = errors.New("workspace not found")
	ErrInsufficientPerms     = errors.New("insufficient permissions")
	ErrInviteNotFound        = errors.New("invite not found")
	ErrInviteExpired         = errors.New("invite expired")
	ErrInviteRevoked         = errors.New("invite revoked")
	ErrInviteAccepted        = errors.New("invite already accepted")
	ErrInviteEmailMismatch   = errors.New("invite email does not match authenticated user")
	ErrFamilyNotEnabled      = errors.New("family plan is not enabled")
	ErrMembersLimitReached   = errors.New("members limit reached for this workspace")
	ErrOwnerRemovalForbidden = errors.New("cannot remove workspace owner")
	ErrSelfRemovalForbidden  = errors.New("cannot remove yourself from workspace")
	ErrMemberNotFound        = errors.New("member not found")
)

type WorkspaceService struct {
	db     *pgxpool.Pool
	rdb    *redis.Client
	queue  *asynq.Client
	appURL string
	logger zerolog.Logger
}

type sendNotificationPayload struct {
	Email   string `json:"email"`
	Subject string `json:"subject"`
	HTML    string `json:"html"`
}

func NewWorkspaceService(db *pgxpool.Pool, rdb *redis.Client, queue *asynq.Client, appURL string, logger zerolog.Logger) *WorkspaceService {
	return &WorkspaceService{
		db:     db,
		rdb:    rdb,
		queue:  queue,
		appURL: strings.TrimRight(appURL, "/"),
		logger: logger,
	}
}

func (s *WorkspaceService) GetWorkspace(ctx context.Context, wsID, userID uuid.UUID) (*domain.Workspace, error) {
	var w domain.Workspace
	err := s.db.QueryRow(ctx, `
		SELECT w.id, w.name, w.slug, w.owner_id, w.created_at, w.updated_at
		FROM workspaces w
		JOIN workspace_members wm ON wm.workspace_id = w.id
		WHERE w.id = $1 AND wm.user_id = $2
	`, wsID, userID).Scan(&w.ID, &w.Name, &w.Slug, &w.OwnerID, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrWorkspaceNotFound
		}
		return nil, err
	}
	return &w, nil
}

type UpdateWorkspaceRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

func (s *WorkspaceService) UpdateWorkspace(ctx context.Context, wsID uuid.UUID, req UpdateWorkspaceRequest) (*domain.Workspace, error) {
	var w domain.Workspace
	err := s.db.QueryRow(ctx, `
		UPDATE workspaces
		SET name = COALESCE(NULLIF($2, ''), name),
			slug = COALESCE(NULLIF($3, ''), slug)
		WHERE id = $1
		RETURNING id, name, slug, owner_id, created_at, updated_at
	`, wsID, strings.TrimSpace(req.Name), strings.TrimSpace(req.Slug)).Scan(&w.ID, &w.Name, &w.Slug, &w.OwnerID, &w.CreatedAt, &w.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrWorkspaceNotFound
		}
		return nil, err
	}
	return &w, nil
}

type WorkspaceUsage struct {
	Counters *domain.WorkspaceCounter  `json:"counters"`
	Limits   *domain.EntitlementLimits `json:"limits"`
}

func (s *WorkspaceService) GetUsage(ctx context.Context, wsID uuid.UUID) (*WorkspaceUsage, error) {
	var c domain.WorkspaceCounter
	err := s.db.QueryRow(ctx, `
		SELECT workspace_id, areas_count, goals_count, habits_count, members_count,
			   tasks_created_today, tasks_today_date, transactions_month_count,
			   transactions_month, storage_bytes_used, updated_at
		FROM workspace_counters WHERE workspace_id = $1
	`, wsID).Scan(
		&c.WorkspaceID, &c.AreasCount, &c.GoalsCount, &c.HabitsCount, &c.MembersCount,
		&c.TasksCreatedToday, &c.TasksTodayDate, &c.TransactionsMonthCount,
		&c.TransactionsMonth, &c.StorageBytesUsed, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	var limitsRaw json.RawMessage
	err = s.db.QueryRow(ctx, `
		SELECT limits FROM workspace_entitlements
		WHERE workspace_id = $1 AND is_current = true
	`, wsID).Scan(&limitsRaw)
	if err != nil {
		return nil, err
	}

	limits, err := domain.ParseLimits(limitsRaw)
	if err != nil {
		return nil, err
	}

	return &WorkspaceUsage{Counters: &c, Limits: limits}, nil
}

func (s *WorkspaceService) ListWorkspaces(ctx context.Context, userID uuid.UUID) ([]domain.WorkspaceListItem, error) {
	rows, err := s.db.Query(ctx, `
		SELECT w.id, w.name, w.slug, wm.role,
			   (up.default_workspace_id = w.id) AS is_default,
			   COALESCE(wc.members_count, 1) AS member_count
		FROM workspace_members wm
		JOIN workspaces w ON w.id = wm.workspace_id
		JOIN user_profiles up ON up.id = wm.user_id
		LEFT JOIN workspace_counters wc ON wc.workspace_id = w.id
		WHERE wm.user_id = $1
		ORDER BY is_default DESC, w.created_at ASC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]domain.WorkspaceListItem, 0)
	for rows.Next() {
		var it domain.WorkspaceListItem
		if err := rows.Scan(&it.ID, &it.Name, &it.Slug, &it.Role, &it.IsDefault, &it.MemberCount); err != nil {
			return nil, err
		}
		items = append(items, it)
	}
	return items, rows.Err()
}

func (s *WorkspaceService) SwitchWorkspace(ctx context.Context, userID, wsID uuid.UUID) error {
	var exists bool
	err := s.db.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM workspace_members
			WHERE workspace_id = $1 AND user_id = $2
		)
	`, wsID, userID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return ErrWorkspaceNotFound
	}

	tag, err := s.db.Exec(ctx, `
		UPDATE user_profiles
		SET default_workspace_id = $2
		WHERE id = $1
	`, userID, wsID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrWorkspaceNotFound
	}

	s.invalidateTenantCache(ctx, userID)
	return nil
}

func (s *WorkspaceService) ListMembers(ctx context.Context, wsID uuid.UUID) ([]domain.WorkspaceMemberSummary, error) {
	rows, err := s.db.Query(ctx, `
		SELECT wm.user_id, up.display_name, up.email, wm.role, wm.joined_at
		FROM workspace_members wm
		JOIN user_profiles up ON up.id = wm.user_id
		WHERE wm.workspace_id = $1
		ORDER BY wm.joined_at ASC
	`, wsID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]domain.WorkspaceMemberSummary, 0)
	for rows.Next() {
		var m domain.WorkspaceMemberSummary
		if err := rows.Scan(&m.UserID, &m.DisplayName, &m.Email, &m.Role, &m.JoinedAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

func (s *WorkspaceService) RemoveMember(ctx context.Context, wsID, actorUserID uuid.UUID, actorRole domain.WorkspaceRole, targetUserID uuid.UUID) error {
	if !actorRole.CanManage() {
		return ErrInsufficientPerms
	}
	if actorUserID == targetUserID {
		return ErrSelfRemovalForbidden
	}

	var targetRole domain.WorkspaceRole
	err := s.db.QueryRow(ctx, `
		SELECT role
		FROM workspace_members
		WHERE workspace_id = $1 AND user_id = $2
	`, wsID, targetUserID).Scan(&targetRole)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrMemberNotFound
		}
		return err
	}
	if targetRole == domain.RoleOwner {
		return ErrOwnerRemovalForbidden
	}

	tag, err := s.db.Exec(ctx, `
		DELETE FROM workspace_members
		WHERE workspace_id = $1 AND user_id = $2
	`, wsID, targetUserID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrMemberNotFound
	}

	_, _ = s.db.Exec(ctx, `
		UPDATE user_profiles
		SET default_workspace_id = (
			SELECT wm.workspace_id
			FROM workspace_members wm
			WHERE wm.user_id = $1
			ORDER BY wm.joined_at ASC
			LIMIT 1
		)
		WHERE id = $1 AND default_workspace_id = $2
	`, targetUserID, wsID)

	s.invalidateTenantCache(ctx, targetUserID)
	return nil
}

type CreateWorkspaceInviteRequest struct {
	Email string               `json:"email"`
	Role  domain.WorkspaceRole `json:"role"`
}

func (s *WorkspaceService) CreateInvite(ctx context.Context, wsID, actorUserID uuid.UUID, actorRole domain.WorkspaceRole, req CreateWorkspaceInviteRequest) (*domain.WorkspaceInviteWithURL, error) {
	if !actorRole.CanManage() {
		return nil, ErrInsufficientPerms
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	if email == "" || !strings.Contains(email, "@") {
		return nil, fmt.Errorf("invalid invite email")
	}

	role := req.Role
	if role == "" {
		role = domain.RoleMember
	}
	if role != domain.RoleMember {
		return nil, fmt.Errorf("only member role is allowed for invites")
	}

	limits, membersCount, err := s.getFamilyUsage(ctx, wsID)
	if err != nil {
		return nil, err
	}
	if !limits.FamilyEnabled {
		return nil, ErrFamilyNotEnabled
	}
	if !limits.CanAddMember(membersCount) {
		return nil, ErrMembersLimitReached
	}

	rawToken, tokenHash, err := generateInviteToken()
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().UTC().Add(7 * 24 * time.Hour)

	var invite domain.WorkspaceInvite
	err = s.db.QueryRow(ctx, `
		INSERT INTO workspace_invites (workspace_id, email, role, token_hash, invited_by, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, workspace_id, email, role, invited_by, expires_at,
				 accepted_at, accepted_by, revoked_at, revoked_by, created_at
	`, wsID, email, role, tokenHash, actorUserID, expiresAt).Scan(
		&invite.ID, &invite.WorkspaceID, &invite.Email, &invite.Role, &invite.InvitedBy,
		&invite.ExpiresAt, &invite.AcceptedAt, &invite.AcceptedBy, &invite.RevokedAt,
		&invite.RevokedBy, &invite.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	inviteURL := fmt.Sprintf("%s/invite/%s", s.appURL, rawToken)
	s.enqueueInviteEmail(ctx, email, inviteURL)

	return &domain.WorkspaceInviteWithURL{
		InviteURL: inviteURL,
		Invite:    invite,
	}, nil
}

func (s *WorkspaceService) ListInvites(ctx context.Context, wsID uuid.UUID, limit, offset int) ([]domain.WorkspaceInvite, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	rows, err := s.db.Query(ctx, `
		SELECT id, workspace_id, email, role, invited_by, expires_at,
			   accepted_at, accepted_by, revoked_at, revoked_by, created_at
		FROM workspace_invites
		WHERE workspace_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, wsID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	invites := make([]domain.WorkspaceInvite, 0)
	for rows.Next() {
		var inv domain.WorkspaceInvite
		if err := rows.Scan(
			&inv.ID, &inv.WorkspaceID, &inv.Email, &inv.Role, &inv.InvitedBy,
			&inv.ExpiresAt, &inv.AcceptedAt, &inv.AcceptedBy, &inv.RevokedAt,
			&inv.RevokedBy, &inv.CreatedAt,
		); err != nil {
			return nil, err
		}
		invites = append(invites, inv)
	}
	return invites, rows.Err()
}

func (s *WorkspaceService) RevokeInvite(ctx context.Context, wsID, inviteID, actorUserID uuid.UUID, actorRole domain.WorkspaceRole) error {
	if !actorRole.CanManage() {
		return ErrInsufficientPerms
	}

	tag, err := s.db.Exec(ctx, `
		UPDATE workspace_invites
		SET revoked_at = now(), revoked_by = $3
		WHERE id = $1
		  AND workspace_id = $2
		  AND accepted_at IS NULL
		  AND revoked_at IS NULL
	`, inviteID, wsID, actorUserID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrInviteNotFound
	}
	return nil
}

func (s *WorkspaceService) AcceptInvite(ctx context.Context, userID uuid.UUID, token string) (uuid.UUID, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return uuid.Nil, ErrInviteNotFound
	}
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return uuid.Nil, err
	}
	defer tx.Rollback(ctx)

	var inviteID uuid.UUID
	var wsID uuid.UUID
	var inviteEmail string
	var role domain.WorkspaceRole
	var expiresAt time.Time
	var acceptedAt *time.Time
	var revokedAt *time.Time
	err = tx.QueryRow(ctx, `
		SELECT id, workspace_id, email, role, expires_at, accepted_at, revoked_at
		FROM workspace_invites
		WHERE token_hash = $1
		FOR UPDATE
	`, tokenHash).Scan(&inviteID, &wsID, &inviteEmail, &role, &expiresAt, &acceptedAt, &revokedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrInviteNotFound
		}
		return uuid.Nil, err
	}
	if acceptedAt != nil {
		return uuid.Nil, ErrInviteAccepted
	}
	if revokedAt != nil {
		return uuid.Nil, ErrInviteRevoked
	}
	if time.Now().UTC().After(expiresAt) {
		return uuid.Nil, ErrInviteExpired
	}

	var userEmail string
	err = tx.QueryRow(ctx, `SELECT email FROM user_profiles WHERE id = $1`, userID).Scan(&userEmail)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, ErrWorkspaceNotFound
		}
		return uuid.Nil, err
	}
	if !strings.EqualFold(strings.TrimSpace(userEmail), strings.TrimSpace(inviteEmail)) {
		return uuid.Nil, ErrInviteEmailMismatch
	}

	var limitsRaw json.RawMessage
	var membersCount int
	err = tx.QueryRow(ctx, `
		SELECT we.limits, wc.members_count
		FROM workspace_entitlements we
		JOIN workspace_counters wc ON wc.workspace_id = we.workspace_id
		WHERE we.workspace_id = $1 AND we.is_current = true
		FOR UPDATE OF wc
	`, wsID).Scan(&limitsRaw, &membersCount)
	if err != nil {
		return uuid.Nil, err
	}
	limits, err := domain.ParseLimits(limitsRaw)
	if err != nil {
		return uuid.Nil, err
	}
	if !limits.FamilyEnabled {
		return uuid.Nil, ErrFamilyNotEnabled
	}
	if !limits.CanAddMember(membersCount) {
		return uuid.Nil, ErrMembersLimitReached
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO workspace_members (workspace_id, user_id, role)
		VALUES ($1, $2, $3)
		ON CONFLICT (workspace_id, user_id) DO NOTHING
	`, wsID, userID, role)
	if err != nil {
		return uuid.Nil, err
	}

	_, err = tx.Exec(ctx, `
		UPDATE workspace_invites
		SET accepted_at = now(), accepted_by = $2
		WHERE id = $1
	`, inviteID, userID)
	if err != nil {
		return uuid.Nil, err
	}

	_, err = tx.Exec(ctx, `
		UPDATE user_profiles
		SET default_workspace_id = $2,
			onboarding_completed = true
		WHERE id = $1
	`, userID, wsID)
	if err != nil {
		return uuid.Nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return uuid.Nil, err
	}

	s.invalidateTenantCache(ctx, userID)
	return wsID, nil
}

func (s *WorkspaceService) getFamilyUsage(ctx context.Context, wsID uuid.UUID) (*domain.EntitlementLimits, int, error) {
	var limitsRaw json.RawMessage
	var membersCount int
	err := s.db.QueryRow(ctx, `
		SELECT we.limits, wc.members_count
		FROM workspace_entitlements we
		JOIN workspace_counters wc ON wc.workspace_id = we.workspace_id
		WHERE we.workspace_id = $1 AND we.is_current = true
	`, wsID).Scan(&limitsRaw, &membersCount)
	if err != nil {
		return nil, 0, err
	}

	limits, err := domain.ParseLimits(limitsRaw)
	if err != nil {
		return nil, 0, err
	}
	return limits, membersCount, nil
}

func (s *WorkspaceService) enqueueInviteEmail(ctx context.Context, toEmail, inviteURL string) {
	if s.queue == nil {
		return
	}

	payload, err := json.Marshal(sendNotificationPayload{
		Email:   toEmail,
		Subject: "Convite para entrar no workspace",
		HTML:    fmt.Sprintf("<p>Voce foi convidado para um workspace no Widia Omni.</p><p>Aceite o convite em: %s</p>", inviteURL),
	})
	if err != nil {
		s.logger.Warn().Err(err).Str("email", toEmail).Msg("failed to marshal invite email payload")
		return
	}

	task := asynq.NewTask(
		"notification:send",
		payload,
		asynq.MaxRetry(3),
		asynq.Timeout(5*time.Minute),
	)
	if _, err := s.queue.EnqueueContext(ctx, task, asynq.Queue("default")); err != nil {
		s.logger.Warn().Err(err).Str("email", toEmail).Msg("failed to enqueue invite email")
	}
}

func (s *WorkspaceService) invalidateTenantCache(ctx context.Context, userID uuid.UUID) {
	if s.rdb == nil {
		return
	}
	cacheKey := fmt.Sprintf("tenant:%s", userID.String())
	_ = s.rdb.Del(ctx, cacheKey).Err()
}

func generateInviteToken() (string, string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", err
	}
	raw := hex.EncodeToString(buf)
	h := sha256.Sum256([]byte(raw))
	return raw, hex.EncodeToString(h[:]), nil
}
