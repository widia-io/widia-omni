package domain

import (
	"time"

	"github.com/google/uuid"
)

type WorkspaceRole string

const (
	RoleOwner  WorkspaceRole = "owner"
	RoleAdmin  WorkspaceRole = "admin"
	RoleMember WorkspaceRole = "member"
	RoleViewer WorkspaceRole = "viewer"
)

type Workspace struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	OwnerID   uuid.UUID `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type WorkspaceMember struct {
	WorkspaceID uuid.UUID     `json:"workspace_id"`
	UserID      uuid.UUID     `json:"user_id"`
	Role        WorkspaceRole `json:"role"`
	JoinedAt    time.Time     `json:"joined_at"`
}

type WorkspaceListItem struct {
	ID          uuid.UUID     `json:"id"`
	Name        string        `json:"name"`
	Slug        string        `json:"slug"`
	Role        WorkspaceRole `json:"role"`
	IsDefault   bool          `json:"is_default"`
	MemberCount int           `json:"member_count"`
}

type WorkspaceMemberSummary struct {
	UserID      uuid.UUID     `json:"user_id"`
	DisplayName string        `json:"display_name"`
	Email       string        `json:"email"`
	Role        WorkspaceRole `json:"role"`
	JoinedAt    time.Time     `json:"joined_at"`
}

type WorkspaceInvite struct {
	ID          uuid.UUID     `json:"id"`
	WorkspaceID uuid.UUID     `json:"workspace_id"`
	Email       string        `json:"email"`
	Role        WorkspaceRole `json:"role"`
	InvitedBy   uuid.UUID     `json:"invited_by"`
	ExpiresAt   time.Time     `json:"expires_at"`
	AcceptedAt  *time.Time    `json:"accepted_at,omitempty"`
	AcceptedBy  *uuid.UUID    `json:"accepted_by,omitempty"`
	RevokedAt   *time.Time    `json:"revoked_at,omitempty"`
	RevokedBy   *uuid.UUID    `json:"revoked_by,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
}

type WorkspaceInviteWithURL struct {
	InviteURL string          `json:"invite_url"`
	Invite    WorkspaceInvite `json:"invite"`
}

func (r WorkspaceRole) CanManage() bool {
	return r == RoleOwner || r == RoleAdmin
}
