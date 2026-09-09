package auth

// ErrorMessage is an API error message owned by the auth handlers.
type ErrorMessage string

const (
	ErrMsgUsernameRequired      ErrorMessage = "Username required"
	ErrMsgDatabase              ErrorMessage = "Database error"
	ErrMsgUsernameTaken         ErrorMessage = "Username is already taken. Please log in or choose a different username."
	ErrMsgSaveSession           ErrorMessage = "Failed to save session cookie"
	ErrMsgSaveAuthenticator     ErrorMessage = "Failed to save authenticator"
	ErrMsgRegistrationSession   ErrorMessage = "Session expired or missing. Please restart registration."
	ErrMsgAuthenticationSession ErrorMessage = "Session expired or missing. Please try again."
	ErrMsgInvalidUserID         ErrorMessage = "Invalid User ID in session"
	ErrMsgInvalidCredentialID   ErrorMessage = "Invalid credential ID encoding"
	ErrMsgPasskeyLookup         ErrorMessage = "Database error while looking up passkey"
	ErrMsgParseWebAuthnResponse ErrorMessage = "Failed to parse WebAuthn response: "
	ErrMsgBiometricVerification ErrorMessage = "Biometric verification failed: "
	ErrMsgInvalidResponseFormat ErrorMessage = "Invalid response format: "
	ErrMsgAuthenticationFailed  ErrorMessage = "Authentication failed: "
	ErrMsgPasskeyNotRecognized  ErrorMessage = "Passkey not recognized"
	ErrMsgUserNotFound          ErrorMessage = "User not found"
	ErrMsgPendingApproval       ErrorMessage = "Your account is pending admin approval."
	ErrMsgSuspended             ErrorMessage = "Your account is suspended. Please contact an administrator."
	ErrMsgGenerateToken         ErrorMessage = "Failed to generate token"
	ErrMsgUnauthorized          ErrorMessage = "Unauthorized"
	ErrMsgAuthenticatorID       ErrorMessage = "Authenticator ID required"
	ErrMsgAuthenticatorNotFound ErrorMessage = "Authenticator not found"
	ErrMsgUserDeleted           ErrorMessage = "User no longer exists"
	ErrMsgTokenRequired         ErrorMessage = "Token required"
	ErrMsgInvalidToken          ErrorMessage = "Invalid token"
	ErrMsgInvalidTokenClaims    ErrorMessage = "Invalid token claims"
)
