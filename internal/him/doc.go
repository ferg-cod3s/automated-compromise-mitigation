// Package him implements the Human-in-the-Middle (HIM) Manager.
//
// The HIM Manager orchestrates workflows where user intervention is required
// due to multi-factor authentication (MFA), CAPTCHA challenges, or Terms of
// Service restrictions that prohibit full automation.
//
// # State Machine
//
// HIM sessions follow a state machine with these states:
//
//   - CREATED: Session initialized, not yet awaiting input
//   - AWAITING_INPUT: Waiting for user to respond to prompt
//   - VALIDATING: User input received, validating response
//   - RESUMING: Validation succeeded, resuming automation
//   - COMPLETED: Session completed successfully
//   - FAILED: Session failed (invalid input, error)
//   - CANCELLED: User cancelled the session
//   - EXPIRED: Session timed out waiting for response
//
// # HIM Trigger Conditions
//
// HIM workflows are triggered when:
//
//   - MFA (TOTP): Site requires 6-digit TOTP code
//   - MFA (SMS): Site requires SMS verification code
//   - MFA (Push): Site requires push notification approval
//   - CAPTCHA: Site displays CAPTCHA challenge
//   - ToS Violation: ACVS detects automation is prohibited
//   - API Unavailable: No documented API for target service
//
// # Security Considerations
//
//   - No Password Capture: HIM prompts NEVER ask for master password
//   - Timeout Enforcement: Limited time to respond (prevents session hijacking)
//   - Secure Input: All user input transmitted over mTLS, never logged plaintext
//   - Context Preservation: State maintained across pause/resume cycles
//
// # Example Usage
//
//	ctx := context.Background()
//	himMgr := him.NewManager()
//
//	// Check if HIM is required for a rotation
//	action := him.RotationAction{
//	    CredentialID: "cred-123",
//	    Site:         "github.com",
//	    ActionType:   him.ActionPasswordChange,
//	}
//	required, himType, _ := himMgr.RequiresHIM(ctx, action)
//	if required {
//	    log.Printf("HIM required: %s", himType)
//	}
//
//	// Prompt user for MFA code
//	prompt := him.HIMPrompt{
//	    SessionID: "session-456",
//	    Type:      him.HIMMFA,
//	    Site:      "github.com",
//	    Message:   "Enter your 6-digit TOTP code",
//	    InputType: him.InputTOTP,
//	    Timeout:   5 * time.Minute,
//	}
//	response, err := himMgr.PromptUser(ctx, prompt)
//	if err != nil {
//	    log.Fatalf("HIM prompt failed: %v", err)
//	}
//
//	// Resume automation with user's input
//	err = himMgr.ResumeAutomation(ctx, prompt.SessionID, response)
//	if err != nil {
//	    log.Fatalf("Resume failed: %v", err)
//	}
//
// # Phase I Implementation
//
// Phase I focuses on:
//   - Basic MFA (TOTP) support
//   - Session state management
//   - gRPC streaming for real-time prompts
//   - Timeout handling
//
// Future phases will add:
//   - CAPTCHA solving integration
//   - Push notification support
//   - SMS code handling
//   - Multi-step HIM workflows
package him
