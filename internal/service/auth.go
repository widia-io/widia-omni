package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type AuthService struct {
	supabaseURL string
	serviceKey  string
	httpClient  *http.Client
}

func NewAuthService(supabaseURL, serviceKey string) *AuthService {
	return &AuthService{
		supabaseURL: supabaseURL,
		serviceKey:  serviceKey,
		httpClient:  &http.Client{},
	}
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Data     map[string]interface{} `json:"data,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type AuthResponse struct {
	AccessToken  string          `json:"access_token,omitempty"`
	RefreshToken string          `json:"refresh_token,omitempty"`
	User         json.RawMessage `json:"user,omitempty"`
	Error        string          `json:"error,omitempty"`
	Message      string          `json:"msg,omitempty"`
}

func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (*AuthResponse, error) {
	return s.postAuth(ctx, "/auth/v1/signup", req)
}

func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*AuthResponse, error) {
	body := map[string]string{"email": req.Email, "password": req.Password}
	return s.postAuth(ctx, "/auth/v1/token?grant_type=password", body)
}

func (s *AuthService) Refresh(ctx context.Context, req RefreshRequest) (*AuthResponse, error) {
	body := map[string]string{"refresh_token": req.RefreshToken}
	return s.postAuth(ctx, "/auth/v1/token?grant_type=refresh_token", body)
}

func (s *AuthService) Logout(ctx context.Context, accessToken string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.supabaseURL+"/auth/v1/logout", nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("apikey", s.serviceKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	body := map[string]string{"email": email}
	_, err := s.postAuth(ctx, "/auth/v1/recover", body)
	return err
}

func (s *AuthService) ResetPassword(ctx context.Context, accessToken, newPassword string) error {
	body := map[string]string{"password": newPassword}
	data, _ := json.Marshal(body)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, s.supabaseURL+"/auth/v1/user", bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("apikey", s.serviceKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, tokenHash, verType string) error {
	url := fmt.Sprintf("%s/auth/v1/verify?token=%s&type=%s", s.supabaseURL, tokenHash, verType)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("apikey", s.serviceKey)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (s *AuthService) postAuth(ctx context.Context, path string, body interface{}) (*AuthResponse, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.supabaseURL+path, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("apikey", s.serviceKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var authResp AuthResponse
	if err := json.Unmarshal(respBody, &authResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if resp.StatusCode >= 400 {
		msg := authResp.Error
		if msg == "" {
			msg = authResp.Message
		}
		return nil, fmt.Errorf("auth error (%d): %s", resp.StatusCode, msg)
	}

	return &authResp, nil
}
