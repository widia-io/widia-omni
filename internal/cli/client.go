package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultRequestTimeout = 25 * time.Second

var (
	ErrNotAuthenticated = errors.New("não autenticado")
	ErrNoRefreshToken   = errors.New("não foi possível renovar sessão")
)

type Client struct {
	httpClient *http.Client
	session    *SessionStore
}

func NewClient(session *SessionStore) *Client {
	return &Client{
		session: session,
		httpClient: &http.Client{
			Timeout: defaultRequestTimeout,
		},
	}
}

type APIError struct {
	Status  int
	Message string
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s (%d)", e.Message, e.Status)
	}
	return fmt.Sprintf("erro da API (%d)", e.Status)
}

type authResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	User         struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	} `json:"user"`
	Error   string `json:"error"`
	Message string `json:"msg"`
}

type UserProfile struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Email       string `json:"email"`
}

type Workspace struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type WorkspaceListItem struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Role        string `json:"role"`
	IsDefault   bool   `json:"is_default"`
	MemberCount int    `json:"member_count"`
}

type WorkspaceCounters struct {
	AreasCount        int `json:"areas_count"`
	GoalsCount        int `json:"goals_count"`
	HabitsCount       int `json:"habits_count"`
	ProjectsCount     int `json:"projects_count"`
	MembersCount      int `json:"members_count"`
	TasksCreatedToday int `json:"tasks_created_today"`
	TransactionsMonth int `json:"transactions_month_count"`
	StorageBytesUsed  int `json:"storage_bytes_used"`
}

type EntitlementLimits struct {
	MaxAreas        int `json:"max_areas"`
	MaxGoals        int `json:"max_goals"`
	MaxHabits       int `json:"max_habits"`
	MaxProjects     int `json:"max_projects"`
	MaxMembers      int `json:"max_members"`
	MaxTasksPerDay  int `json:"max_tasks_per_day"`
	MaxTransactions int `json:"max_transactions_per_month"`
}

type WorkspaceUsage struct {
	Counters WorkspaceCounters `json:"counters"`
	Limits   EntitlementLimits `json:"limits"`
}

type Area struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Slug     string  `json:"slug"`
	Icon     string  `json:"icon"`
	Color    string  `json:"color"`
	Weight   float64 `json:"weight"`
	IsActive bool    `json:"is_active"`
}

type Task struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	IsCompleted bool   `json:"is_completed"`
	IsFocus     bool   `json:"is_focus"`
	Priority    string `json:"priority"`
}

type Project struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type Label struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type taskUpdateRequest struct {
	Position int `json:"position"`
}

type TaskCreateRequest struct {
	Title           string   `json:"title"`
	Description     *string  `json:"description,omitempty"`
	Priority        string   `json:"priority"`
	AreaID          *string  `json:"area_id,omitempty"`
	ProjectID       *string  `json:"project_id,omitempty"`
	DueDate         *string  `json:"due_date,omitempty"`
	DurationMinutes *int     `json:"duration_minutes,omitempty"`
	LabelIDs        []string `json:"label_ids,omitempty"`
}

type areaCreateRequest struct {
	Name      string  `json:"name"`
	Slug      string  `json:"slug"`
	Icon      string  `json:"icon"`
	Color     string  `json:"color"`
	Weight    float64 `json:"weight"`
	SortOrder int     `json:"sort_order"`
}

type workspaceSwitchRequest struct {
	WorkspaceID string `json:"workspace_id"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (c *Client) baseURL() string {
	baseURL := strings.TrimSpace(c.session.BaseURL())
	if baseURL == "" {
		return "http://localhost:8080"
	}
	return baseURL
}

func (c *Client) Login(ctx context.Context, email, password string) error {
	req := map[string]string{"email": email, "password": password}
	var resp authResponse
	if err := c.request(ctx, http.MethodPost, "/auth/login", req, &resp, false); err != nil {
		return err
	}
	if resp.AccessToken == "" {
		return errors.New("resposta de login sem token")
	}
	refreshToken := strings.TrimSpace(resp.RefreshToken)
	if refreshToken == "" {
		refreshToken = c.session.RefreshToken()
	}
	return c.session.SetAuth(resp.AccessToken, refreshToken, resp.User.ID, resp.User.Email)
}

func (c *Client) Refresh(ctx context.Context) error {
	refreshToken := c.session.RefreshToken()
	if refreshToken == "" {
		return ErrNoRefreshToken
	}

	req := refreshRequest{RefreshToken: refreshToken}
	var resp authResponse
	if err := c.request(ctx, http.MethodPost, "/auth/refresh", req, &resp, false); err != nil {
		return err
	}
	if resp.AccessToken == "" {
		return errors.New("resposta de refresh sem token")
	}
	nextRefresh := resp.RefreshToken
	if nextRefresh == "" {
		nextRefresh = refreshToken
	}
	return c.session.SetAuth(resp.AccessToken, nextRefresh, c.session.Snapshot().UserID, c.session.UserEmail())
}

func (c *Client) Logout(ctx context.Context) error {
	_ = c.request(ctx, http.MethodPost, "/auth/logout", nil, nil, true)
	return c.session.Clear()
}

func (c *Client) Me(ctx context.Context) (*UserProfile, error) {
	var out UserProfile
	err := c.request(ctx, http.MethodGet, "/api/v1/me", nil, &out, true)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) CurrentWorkspace(ctx context.Context) (*Workspace, error) {
	var out Workspace
	err := c.request(ctx, http.MethodGet, "/api/v1/workspace", nil, &out, true)
	if err != nil {
		return nil, err
	}
	return &out, nil
}

func (c *Client) ListWorkspaces(ctx context.Context) ([]WorkspaceListItem, error) {
	var out []WorkspaceListItem
	err := c.request(ctx, http.MethodGet, "/api/v1/workspaces", nil, &out, true)
	return out, err
}

func (c *Client) SwitchWorkspace(ctx context.Context, workspaceID string) error {
	req := workspaceSwitchRequest{WorkspaceID: workspaceID}
	return c.request(ctx, http.MethodPost, "/api/v1/workspace/switch", req, nil, true)
}

func (c *Client) GetWorkspaceUsage(ctx context.Context) (*WorkspaceUsage, error) {
	var out WorkspaceUsage
	err := c.request(ctx, http.MethodGet, "/api/v1/workspace/usage", nil, &out, true)
	return &out, err
}

func (c *Client) ListTasks(ctx context.Context, completed *bool) ([]Task, error) {
	path := "/api/v1/tasks"
	if completed != nil {
		if *completed {
			path = "/api/v1/tasks?is_completed=true"
		} else {
			path = "/api/v1/tasks?is_completed=false"
		}
	}
	var out []Task
	err := c.request(ctx, http.MethodGet, path, nil, &out, true)
	return out, err
}

func (c *Client) CreateTask(ctx context.Context, req TaskCreateRequest) (*Task, error) {
	if req.Priority == "" {
		req.Priority = "medium"
	}
	if req.Description != nil {
		description := strings.TrimSpace(*req.Description)
		if description == "" {
			req.Description = nil
		} else {
			req.Description = &description
		}
	}
	if strings.TrimSpace(req.Title) == "" {
		return nil, errors.New("título é obrigatório")
	}
	var out Task
	err := c.request(ctx, http.MethodPost, "/api/v1/tasks", req, &out, true)
	return &out, err
}

func (c *Client) CompleteTask(ctx context.Context, id string) (*Task, error) {
	path := "/api/v1/tasks/" + id + "/complete"
	var out Task
	err := c.request(ctx, http.MethodPatch, path, taskUpdateRequest{}, &out, true)
	return &out, err
}

func (c *Client) ReopenTask(ctx context.Context, id string) (*Task, error) {
	path := "/api/v1/tasks/" + id + "/reopen"
	var out Task
	err := c.request(ctx, http.MethodPatch, path, taskUpdateRequest{}, &out, true)
	return &out, err
}

func (c *Client) ListAreas(ctx context.Context) ([]Area, error) {
	var out []Area
	err := c.request(ctx, http.MethodGet, "/api/v1/areas", nil, &out, true)
	return out, err
}

func (c *Client) CreateArea(ctx context.Context, name, slug string) (*Area, error) {
	req := areaCreateRequest{
		Name:      name,
		Slug:      slug,
		Icon:      "star",
		Color:     "blue",
		Weight:    1,
		SortOrder: 0,
	}
	var out Area
	err := c.request(ctx, http.MethodPost, "/api/v1/areas", req, &out, true)
	return &out, err
}

func (c *Client) ListProjects(ctx context.Context) ([]Project, error) {
	var out []Project
	err := c.request(ctx, http.MethodGet, "/api/v1/projects", nil, &out, true)
	return out, err
}

func (c *Client) ListLabels(ctx context.Context) ([]Label, error) {
	var out []Label
	err := c.request(ctx, http.MethodGet, "/api/v1/labels", nil, &out, true)
	return out, err
}

func (c *Client) request(ctx context.Context, method, path string, payload any, out any, withAuth bool) error {
	return c.requestWithRetry(ctx, method, path, payload, out, withAuth, true)
}

func (c *Client) requestWithRetry(ctx context.Context, method, path string, payload any, out any, withAuth bool, canRetry bool) error {
	req, err := c.newRequest(ctx, method, path, payload, withAuth)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized && withAuth && canRetry {
		if c.session.RefreshToken() == "" {
			c.session.Clear()
			return ErrNotAuthenticated
		}
		if err := c.Refresh(ctx); err != nil {
			c.session.Clear()
			return ErrNotAuthenticated
		}
		return c.requestWithRetry(ctx, method, path, payload, out, withAuth, false)
	}

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}
	if resp.StatusCode >= http.StatusBadRequest {
		var parsed struct {
			Error   string `json:"error"`
			Message string `json:"message"`
		}
		raw, _ := io.ReadAll(resp.Body)
		if len(raw) > 0 {
			_ = json.Unmarshal(raw, &parsed)
		}
		msg := strings.TrimSpace(parsed.Error)
		if msg == "" {
			msg = strings.TrimSpace(parsed.Message)
		}
		if msg == "" {
			msg = strings.TrimSpace(string(raw))
		}
		if msg == "" {
			msg = resp.Status
		}
		return &APIError{Status: resp.StatusCode, Message: msg}
	}

	if out == nil {
		return nil
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if len(body) == 0 {
		return nil
	}
	return json.Unmarshal(body, out)
}

func (c *Client) newRequest(ctx context.Context, method, path string, payload any, withAuth bool) (*http.Request, error) {
	url := c.baseURL() + path

	var body io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if withAuth {
		token := c.session.AccessToken()
		if token == "" {
			return nil, ErrNotAuthenticated
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req, nil
}
