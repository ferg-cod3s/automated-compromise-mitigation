package him

import "time"

// Session represents a Human-in-the-Middle interaction session.
type Session struct {
	// ID is the unique session identifier.
	ID string

	// Type indicates the type of HIM required.
	Type HIMType

	// CredentialID is the ID of the credential requiring intervention.
	CredentialID string

	// Site is the website/service associated with this HIM session.
	Site string

	// Prompt is the message shown to the user.
	Prompt string

	// ExpectedInput describes what input the user should provide.
	ExpectedInput string

	// SecurityToken prevents CSRF attacks.
	SecurityToken string

	// State is the current state of the session.
	State SessionState

	// CreatedAt is when the session was created.
	CreatedAt time.Time

	// ExpiresAt is when the session will expire.
	ExpiresAt time.Time

	// LastUpdated is when the session was last modified.
	LastUpdated time.Time

	// CompletedAt is when the session completed (success or failure).
	CompletedAt time.Time

	// AttemptCount tracks how many times the user has tried.
	AttemptCount int

	// MaxAttempts is the maximum allowed attempts.
	MaxAttempts int

	// responseChannel is used internally to communicate responses.
	responseChannel chan Response
}

// SessionRequest contains parameters for creating a new HIM session.
type SessionRequest struct {
	// Type indicates the type of HIM required.
	Type HIMType

	// CredentialID is the credential requiring intervention.
	CredentialID string

	// Site is the website/service name.
	Site string

	// Prompt is the message to show the user.
	Prompt string

	// ExpectedInput describes what input is expected.
	ExpectedInput string

	// MaxAttempts limits retry attempts (default: 3).
	MaxAttempts int

	// Timeout overrides the default session timeout.
	Timeout time.Duration
}

// Response contains the user's response to a HIM prompt.
type Response struct {
	// SessionID is the session this response belongs to.
	SessionID string

	// SecurityToken must match the session's token.
	SecurityToken string

	// Data contains the user's input.
	Data ResponseData

	// Timestamp is when the response was submitted.
	Timestamp time.Time
}

// ResponseData holds the actual user input in typed form.
type ResponseData struct {
	// TextInput for codes, passwords, etc.
	TextInput string

	// BooleanInput for yes/no questions.
	BooleanInput bool

	// ChoiceInput for multiple choice (index of selected option).
	ChoiceInput int

	// FileInput for file uploads (e.g., screenshots).
	FileInput []byte
}

// HIMType indicates the type of human intervention required.
type HIMType string

const (
	// HIMMFA indicates multi-factor authentication is required.
	HIMMFA HIMType = "mfa"

	// HIMTOTP indicates TOTP code entry is required.
	HIMTOTP HIMType = "totp"

	// HIMSMS indicates SMS code entry is required.
	HIMSMS HIMType = "sms"

	// HIMPush indicates push notification approval is required.
	HIMPush HIMType = "push"

	// HIMEmail indicates email verification is required.
	HIMEmail HIMType = "email"

	// HIMCAPTCHA indicates CAPTCHA solving is required.
	HIMCAPTCHA HIMType = "captcha"

	// HIMManualRotation indicates manual password change is required.
	HIMManualRotation HIMType = "manual_rotation"

	// HIMToSReview indicates Terms of Service review is required.
	HIMToSReview HIMType = "tos_review"

	// HIMBiometric indicates biometric authentication is required.
	HIMBiometric HIMType = "biometric"

	// HIMSecurityKey indicates hardware security key is required.
	HIMSecurityKey HIMType = "security_key"
)

// SessionState indicates the current state of a HIM session.
type SessionState string

const (
	// StateInitialized indicates the session was just created.
	StateInitialized SessionState = "initialized"

	// StatePending indicates waiting for user response.
	StatePending SessionState = "pending"

	// StateProcessing indicates processing user response.
	StateProcessing SessionState = "processing"

	// StateCompleted indicates successful completion.
	StateCompleted SessionState = "completed"

	// StateFailed indicates the session failed.
	StateFailed SessionState = "failed"

	// StateCancelled indicates the session was cancelled.
	StateCancelled SessionState = "cancelled"

	// StateTimeout indicates the session expired.
	StateTimeout SessionState = "timeout"
)
