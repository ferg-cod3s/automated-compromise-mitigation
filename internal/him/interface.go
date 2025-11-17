// Package him implements the Human-in-the-Middle (HIM) Manager.
//
// The HIM Manager orchestrates workflows where user intervention is required
// due to MFA, CAPTCHA, or policy restrictions. It maintains state across
// pause/resume cycles and ensures secure communication of user input.
package him

import (
	"context"
	"time"
)

// HIMManager manages Human-in-the-Middle workflows and state.
type HIMManager interface {
	// RequiresHIM determines if a rotation action requires HIM intervention.
	// Returns true if HIM is needed, along with the specific HIM type required.
	RequiresHIM(ctx context.Context, action RotationAction) (bool, HIMType, error)

	// PromptUser sends a prompt to the client and waits for user response.
	// This is a blocking call with a configurable timeout (default: 5 minutes).
	// Returns the user's response or an error if timeout occurs.
	PromptUser(ctx context.Context, prompt HIMPrompt) (*HIMResponse, error)

	// ResumeAutomation continues the automation workflow after receiving HIM input.
	// The response contains the user's input (e.g., TOTP code, CAPTCHA solution).
	ResumeAutomation(ctx context.Context, sessionID string, response *HIMResponse) error

	// GetSessionState retrieves the current state of a HIM session.
	GetSessionState(ctx context.Context, sessionID string) (*HIMSessionState, error)

	// CancelSession cancels an in-progress HIM session.
	CancelSession(ctx context.Context, sessionID string) error

	// ListActiveSessions returns all active HIM sessions.
	ListActiveSessions(ctx context.Context) ([]*HIMSessionState, error)
}

// RotationAction describes the action being performed that may require HIM.
type RotationAction struct {
	// CredentialID is the ID of the credential being rotated.
	CredentialID string

	// Site is the target website or service.
	Site string

	// ActionType describes the type of action (password_change, account_recovery, etc.).
	ActionType ActionType

	// Method is the intended rotation method (auto, API, manual).
	Method string

	// Timestamp is when the action was initiated.
	Timestamp time.Time
}

// ActionType describes the type of rotation action.
type ActionType string

const (
	// ActionPasswordChange indicates a password change operation.
	ActionPasswordChange ActionType = "password_change"

	// ActionAccountRecovery indicates an account recovery operation.
	ActionAccountRecovery ActionType = "account_recovery"

	// ActionEmailChange indicates an email address change operation.
	ActionEmailChange ActionType = "email_change"
)

// HIMPrompt represents a prompt sent to the user during HIM workflow.
type HIMPrompt struct {
	// SessionID uniquely identifies this HIM session.
	SessionID string

	// Type is the type of HIM required.
	Type HIMType

	// Site is the website or service being accessed.
	Site string

	// Message is the human-readable prompt message.
	Message string

	// InputType describes what kind of input is expected.
	InputType InputType

	// Timeout is how long to wait for user response.
	Timeout time.Duration

	// Options contains additional context or choices for the user.
	Options map[string]interface{}

	// CreatedAt is when this prompt was created.
	CreatedAt time.Time
}

// InputType describes the expected user input.
type InputType string

const (
	// InputTOTP indicates a 6-digit TOTP code is expected.
	InputTOTP InputType = "totp"

	// InputSMS indicates an SMS verification code is expected.
	InputSMS InputType = "sms"

	// InputCAPTCHA indicates CAPTCHA solution is expected.
	InputCAPTCHA InputType = "captcha"

	// InputConfirmation indicates yes/no confirmation is expected.
	InputConfirmation InputType = "confirmation"

	// InputText indicates arbitrary text input is expected.
	InputText InputType = "text"
)

// HIMResponse represents the user's response to a HIM prompt.
type HIMResponse struct {
	// SessionID links this response to the original prompt.
	SessionID string

	// Input contains the user's input (e.g., TOTP code, CAPTCHA text).
	Input string

	// Confirmed indicates if the user confirmed the action (for confirmation prompts).
	Confirmed bool

	// CancelRequested indicates if the user wants to cancel the operation.
	CancelRequested bool

	// RespondedAt is when the user responded.
	RespondedAt time.Time
}

// HIMSessionState represents the current state of a HIM session.
type HIMSessionState struct {
	// SessionID uniquely identifies this session.
	SessionID string

	// State is the current state of the session.
	State SessionState

	// RotationAction is the action that triggered this HIM session.
	RotationAction RotationAction

	// Prompt is the current prompt (if in AWAITING_INPUT state).
	Prompt *HIMPrompt

	// CreatedAt is when this session was created.
	CreatedAt time.Time

	// UpdatedAt is when this session was last updated.
	UpdatedAt time.Time

	// ExpiresAt is when this session will expire if no response is received.
	ExpiresAt time.Time

	// AttemptCount is the number of input attempts made.
	AttemptCount int

	// MaxAttempts is the maximum number of attempts allowed.
	MaxAttempts int
}

// HIMError represents an error during HIM operations.
type HIMError struct {
	// Code is the error code.
	Code HIMErrorCode

	// Message is a human-readable error message.
	Message string

	// Cause is the underlying error, if any.
	Cause error

	// SessionID is the session that encountered the error.
	SessionID string
}

func (e *HIMError) Error() string {
	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}
	return e.Message
}

func (e *HIMError) Unwrap() error {
	return e.Cause
}

// HIMErrorCode represents specific HIM error conditions.
type HIMErrorCode string

const (
	// ErrSessionNotFound indicates the session ID does not exist.
	ErrSessionNotFound HIMErrorCode = "SESSION_NOT_FOUND"

	// ErrSessionExpired indicates the session has expired.
	ErrSessionExpired HIMErrorCode = "SESSION_EXPIRED"

	// ErrInvalidInput indicates the user's input was invalid.
	ErrInvalidInput HIMErrorCode = "INVALID_INPUT"

	// ErrTimeout indicates the user did not respond within the timeout period.
	ErrTimeout HIMErrorCode = "TIMEOUT"

	// ErrCancelled indicates the session was cancelled by the user.
	ErrCancelled HIMErrorCode = "CANCELLED"

	// ErrMaxAttemptsExceeded indicates the user exceeded the maximum number of attempts.
	ErrMaxAttemptsExceeded HIMErrorCode = "MAX_ATTEMPTS_EXCEEDED"
)
