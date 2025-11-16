// Package him implements the Human-in-the-Middle (HIM) workflow system.
//
// HIM handles scenarios where automated credential rotation requires human
// intervention, such as:
//   - MFA/2FA code entry (TOTP, SMS, Push)
//   - CAPTCHA solving
//   - Manual password changes (when no API is available)
//   - Terms of Service review and acceptance
package him

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// Service manages HIM sessions and user prompts.
type Service struct {
	sessions sync.Map // sessionID -> *Session
	timeout  time.Duration
}

// NewService creates a new HIM service with the specified default timeout.
func NewService(timeout time.Duration) *Service {
	if timeout == 0 {
		timeout = 5 * time.Minute // Default timeout
	}

	return &Service{
		timeout: timeout,
	}
}

// CreateSession creates a new HIM session for user interaction.
func (s *Service) CreateSession(ctx context.Context, req SessionRequest) (*Session, error) {
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	session := &Session{
		ID:              sessionID,
		Type:            req.Type,
		CredentialID:    req.CredentialID,
		Site:            req.Site,
		Prompt:          req.Prompt,
		ExpectedInput:   req.ExpectedInput,
		SecurityToken:   generateSecurityToken(),
		State:           StateInitialized,
		CreatedAt:       time.Now(),
		ExpiresAt:       time.Now().Add(s.timeout),
		AttemptCount:    0,
		MaxAttempts:     req.MaxAttempts,
		responseChannel: make(chan Response, 1),
	}

	if session.MaxAttempts == 0 {
		session.MaxAttempts = 3 // Default max attempts
	}

	s.sessions.Store(sessionID, session)

	// Start timeout goroutine
	go s.handleTimeout(ctx, sessionID)

	return session, nil
}

// GetSession retrieves a session by ID.
func (s *Service) GetSession(ctx context.Context, sessionID string) (*Session, error) {
	value, ok := s.sessions.Load(sessionID)
	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	session, ok := value.(*Session)
	if !ok {
		return nil, fmt.Errorf("invalid session data")
	}

	return session, nil
}

// SubmitResponse submits a user response to a HIM session.
func (s *Service) SubmitResponse(ctx context.Context, sessionID string, response Response) error {
	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	// Verify security token to prevent CSRF
	if response.SecurityToken != session.SecurityToken {
		return fmt.Errorf("invalid security token")
	}

	// Check if session has expired
	if time.Now().After(session.ExpiresAt) {
		session.State = StateTimeout
		return fmt.Errorf("session expired")
	}

	// Check if max attempts exceeded
	if session.AttemptCount >= session.MaxAttempts {
		session.State = StateFailed
		return fmt.Errorf("maximum attempts exceeded")
	}

	session.AttemptCount++
	session.LastUpdated = time.Now()
	session.State = StateProcessing

	// Send response to waiting channel
	select {
	case session.responseChannel <- response:
		session.State = StateCompleted
		session.CompletedAt = time.Now()
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout sending response")
	}
}

// WaitForResponse waits for a user response to a HIM session.
// Returns the response or an error if timeout occurs.
func (s *Service) WaitForResponse(ctx context.Context, sessionID string) (Response, error) {
	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return Response{}, err
	}

	select {
	case response := <-session.responseChannel:
		return response, nil
	case <-time.After(s.timeout):
		session.State = StateTimeout
		return Response{}, fmt.Errorf("timeout waiting for user response")
	case <-ctx.Done():
		session.State = StateCancelled
		return Response{}, ctx.Err()
	}
}

// CancelSession cancels an active HIM session.
func (s *Service) CancelSession(ctx context.Context, sessionID string) error {
	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return err
	}

	session.State = StateCancelled
	session.CompletedAt = time.Now()

	// Close response channel to unblock any waiters
	close(session.responseChannel)

	return nil
}

// ListActiveSessions returns all active (non-completed) HIM sessions.
func (s *Service) ListActiveSessions(ctx context.Context) ([]*Session, error) {
	var sessions []*Session

	s.sessions.Range(func(key, value interface{}) bool {
		session, ok := value.(*Session)
		if !ok {
			return true // Continue iteration
		}

		// Include only active sessions
		if session.State != StateCompleted && session.State != StateFailed &&
			session.State != StateCancelled && session.State != StateTimeout {
			sessions = append(sessions, session)
		}

		return true // Continue iteration
	})

	return sessions, nil
}

// CleanupExpiredSessions removes expired sessions from memory.
func (s *Service) CleanupExpiredSessions(ctx context.Context) {
	var toDelete []string

	s.sessions.Range(func(key, value interface{}) bool {
		sessionID, ok := key.(string)
		if !ok {
			return true
		}

		session, ok := value.(*Session)
		if !ok {
			return true
		}

		// Delete completed or expired sessions older than 1 hour
		if (session.State == StateCompleted || session.State == StateFailed ||
			session.State == StateCancelled || session.State == StateTimeout) &&
			time.Since(session.CompletedAt) > time.Hour {
			toDelete = append(toDelete, sessionID)
		}

		return true
	})

	for _, sessionID := range toDelete {
		s.sessions.Delete(sessionID)
	}
}

// handleTimeout monitors a session and marks it as timed out if it expires.
func (s *Service) handleTimeout(ctx context.Context, sessionID string) {
	session, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return
	}

	// Wait until expiration
	time.Sleep(time.Until(session.ExpiresAt))

	// Check if still pending
	currentSession, err := s.GetSession(ctx, sessionID)
	if err != nil {
		return
	}

	if currentSession.State == StatePending || currentSession.State == StateInitialized {
		currentSession.State = StateTimeout
		currentSession.CompletedAt = time.Now()
	}
}

// Helper functions

func generateSessionID() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func generateSecurityToken() string {
	bytes := make([]byte, 32)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
