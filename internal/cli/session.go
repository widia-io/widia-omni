package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Session struct {
	BaseURL      string `json:"base_url"`
	UserID       string `json:"user_id"`
	UserEmail    string `json:"user_email"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type SessionStore struct {
	mu      sync.RWMutex
	path    string
	session Session
}

func DefaultSessionPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "widia", "cli", "session.json"), nil
}

func NewSessionStore(path string) (*SessionStore, error) {
	store := &SessionStore{path: path}
	if _, err := os.Stat(path); err == nil {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		if len(data) > 0 {
			if err := json.Unmarshal(data, &store.session); err != nil {
				return nil, err
			}
		}
	}
	return store, nil
}

func (s *SessionStore) BaseURL() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.session.BaseURL
}

func (s *SessionStore) SetBaseURL(url string) {
	s.mu.Lock()
	s.session.BaseURL = strings.TrimRight(url, "/")
	s.mu.Unlock()
	_ = s.save()
}

func (s *SessionStore) IsAuthenticated() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.session.AccessToken != "" && s.session.RefreshToken != ""
}

func (s *SessionStore) AccessToken() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.session.AccessToken
}

func (s *SessionStore) RefreshToken() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.session.RefreshToken
}

func (s *SessionStore) UserEmail() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.session.UserEmail
}

func (s *SessionStore) SetAuth(accessToken, refreshToken, userID, email string) error {
	s.mu.Lock()
	s.session.AccessToken = strings.TrimSpace(accessToken)
	s.session.RefreshToken = strings.TrimSpace(refreshToken)
	s.session.UserID = strings.TrimSpace(userID)
	s.session.UserEmail = strings.TrimSpace(email)
	s.mu.Unlock()
	return s.save()
}

func (s *SessionStore) Clear() error {
	s.mu.Lock()
	s.session.AccessToken = ""
	s.session.RefreshToken = ""
	s.session.UserID = ""
	s.session.UserEmail = ""
	s.mu.Unlock()
	return s.save()
}

func (s *SessionStore) Snapshot() Session {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.session
}

func (s *SessionStore) save() error {
	_ = os.MkdirAll(filepath.Dir(s.path), 0o700)
	data, err := json.MarshalIndent(s.session, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o600)
}

func (s *SessionStore) DeleteStorage() error {
	if err := os.Remove(s.path); err != nil && !os.IsNotExist(err) {
		return err
	}
	s.mu.Lock()
	s.session = Session{}
	s.mu.Unlock()
	return nil
}

func (s *SessionStore) String() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.session.UserEmail != "" {
		return fmt.Sprintf("%s (%s)", s.session.UserEmail, s.session.UserID)
	}
	if s.session.UserID != "" {
		return s.session.UserID
	}
	return "sem sessão"
}
