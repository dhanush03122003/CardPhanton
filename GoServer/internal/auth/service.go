package auth

import (
    "crypto/rand"
    "crypto/sha256"
    "crypto/subtle"
    "encoding/base64"
    "encoding/hex"
    "encoding/json"
    "net/http"

    "github.com/go-webauthn/webauthn/protocol"
    "github.com/go-webauthn/webauthn/webauthn"
    "github.com/google/uuid"
)

// WebAuthnService wraps the go-webauthn library
type WebAuthnService struct {
    webAuthn *webauthn.WebAuthn
    config   *Config
}

// Config holds WebAuthn configuration
type Config struct {
    RPID         string
    RPName       string
    RPOrigins    []string
    RPTopOrigins []string
}

// NewWebAuthnService creates a new WebAuthn service
func NewWebAuthnService(cfg *Config) (*WebAuthnService, error) {
    webAuthnInstance, err := webauthn.New(&webauthn.Config{
        RPID:          cfg.RPID,
        RPDisplayName: cfg.RPName,
        RPOrigins:     cfg.RPOrigins,
        RPTopOrigins:  cfg.RPTopOrigins,
        AuthenticatorSelection: protocol.AuthenticatorSelection{
			ResidentKey:        protocol.ResidentKeyRequirementPreferred, 
			RequireResidentKey: protocol.ResidentKeyNotRequired(),          
			UserVerification:   protocol.VerificationPreferred,
		},
    })
    if err != nil {
        return nil, err
    }

    return &WebAuthnService{
        webAuthn: webAuthnInstance,
        config:   cfg,
    }, nil
}

// User is a minimal implementation of webauthn.User
type User struct {
    ID          []byte
    Name        string
    DisplayName string
    Credentials []webauthn.Credential
}

// WebAuthnID returns the user's ID
func (u *User) WebAuthnID() []byte {
    return u.ID
}

// WebAuthnName returns the user's name
func (u *User) WebAuthnName() string {
    return u.Name
}

// WebAuthnDisplayName returns the user's display name
func (u *User) WebAuthnDisplayName() string {
    return u.DisplayName
}

// WebAuthnCredentials returns the user's credentials
func (u *User) WebAuthnCredentials() []webauthn.Credential {
    return u.Credentials
}

// NewUser creates a new WebAuthn user model
func NewUser(userID, username string) *User {
    userUUID, err := uuid.Parse(userID)
    if err != nil {
        userBytes := sha256.Sum256([]byte(userID))
        copy(userBytes[:], userUUID[:])
    }

    return &User{
        ID:          userUUID[:],
        Name:        username,
        DisplayName: username,
    }
}

// Credential represents a WebAuthn credential for filtering
type Credential struct {
    ID         string   `json:"id"`
    Transports []string `json:"transports,omitempty"`
}

// RegistrationData holds registration ceremony data
type RegistrationData struct {
    Options *protocol.CredentialCreation
    Session webauthn.SessionData
}

// LoginData holds login ceremony data
type LoginData struct {
    Options *protocol.CredentialAssertion
    Session webauthn.SessionData
}

// BeginRegistration starts a new registration ceremony
func (s *WebAuthnService) BeginRegistration(userID, username string, existingCredentials []Credential) (*RegistrationData, error) {
    user := NewUser(userID, username)

    excludedCredentials := make([]protocol.CredentialDescriptor, 0, len(existingCredentials))
    for _, cred := range existingCredentials {
        credID, err := base64.RawURLEncoding.DecodeString(cred.ID)
        if err != nil {
            continue
        }
        excludedCredentials = append(excludedCredentials, protocol.CredentialDescriptor{
            Type:         protocol.PublicKeyCredentialType,
            CredentialID: credID,
        })
    }

    opts, session, err := s.webAuthn.BeginRegistration(
        user,
        webauthn.WithExclusions(excludedCredentials),
    )
    if err != nil {
        return nil, err
    }

    return &RegistrationData{Options: opts, Session: *session}, nil
}

// FinishRegistration completes the registration ceremony
func (s *WebAuthnService) FinishRegistration(user *User, session webauthn.SessionData, w http.ResponseWriter, r *http.Request) (*webauthn.Credential, error) {
    return s.webAuthn.FinishRegistration(user, session, r)
}

// BeginAuthentication starts a new authentication ceremony
func (s *WebAuthnService) BeginAuthentication(userID string, allowCredentials []Credential) (*LoginData, error) {
    user := NewUser(userID, "")

    credentials := make([]protocol.CredentialDescriptor, 0, len(allowCredentials))
    for _, cred := range allowCredentials {
        credID, err := base64.RawURLEncoding.DecodeString(cred.ID)
        if err != nil {
            continue
        }
        var transports []protocol.AuthenticatorTransport
        if cred.Transports != nil {
            for _, t := range cred.Transports {
                transports = append(transports, protocol.AuthenticatorTransport(t))
            }
        } else {
            transports = []protocol.AuthenticatorTransport{protocol.Internal}
        }
        credentials = append(credentials, protocol.CredentialDescriptor{
            Type:         protocol.PublicKeyCredentialType,
            CredentialID: credID,
            Transport:    transports,
        })
    }

    opts, session, err := s.webAuthn.BeginLogin(
        user,
        webauthn.WithAllowedCredentials(credentials),
    )
    if err != nil {
        return nil, err
    }

    return &LoginData{Options: opts, Session: *session}, nil
}

// FinishAuthentication completes the authentication ceremony
func (s *WebAuthnService) FinishAuthentication(user *User, session webauthn.SessionData, w http.ResponseWriter, r *http.Request) (*webauthn.Credential, error) {
    return s.webAuthn.FinishLogin(user, session, r)
}

// CreateCredential completes registration by parsing the raw response data (bypassing http.Request)
func (s *WebAuthnService) CreateCredential(user *User, session webauthn.SessionData, response *protocol.ParsedCredentialCreationData) (*webauthn.Credential, error) {
    return s.webAuthn.CreateCredential(user, session, response)
}

// ValidateLogin completes authentication by parsing the raw response data (bypassing http.Request)
func (s *WebAuthnService) ValidateLogin(user *User, session webauthn.SessionData, response *protocol.ParsedCredentialAssertionData) (*webauthn.Credential, error) {
    return s.webAuthn.ValidateLogin(user, session, response)
}

func GenerateChallenge() string {
    bytes := make([]byte, 32)
    rand.Read(bytes)
    return base64.RawURLEncoding.EncodeToString(bytes)
}

func EncodeBase64URL(data []byte) string {
    return base64.RawURLEncoding.EncodeToString(data)
}

func DecodeBase64URL(data string) ([]byte, error) {
    return base64.RawURLEncoding.DecodeString(data)
}

func VerifyChallenge(expected, received string) bool {
    expectedBytes, err := base64.RawURLEncoding.DecodeString(expected)
    if err != nil {
        return false
    }
    receivedBytes, err := base64.RawURLEncoding.DecodeString(received)
    if err != nil {
        return false
    }
    return subtle.ConstantTimeCompare(expectedBytes, receivedBytes) == 1
}

func EncodeHex(data []byte) string {
    return hex.EncodeToString(data)
}

func DecodeHex(data string) ([]byte, error) {
    return hex.DecodeString(data)
}

type SessionData = webauthn.SessionData
type PublicKeyCredentialCreationOptions = protocol.CredentialCreation
type PublicKeyCredentialRequestOptions = protocol.CredentialAssertion

func ParseCredentialRequest(data json.RawMessage) (*protocol.PublicKeyCredential, error) {
    var cred protocol.PublicKeyCredential
    if err := json.Unmarshal(data, &cred); err != nil {
        return nil, err
    }
    return &cred, nil
}

// BeginDiscoverableLogin starts a new authentication ceremony without a known user (Conditional UI)
func (s *WebAuthnService) BeginDiscoverableLogin() (*LoginData, error) {
    opts, session, err := s.webAuthn.BeginDiscoverableLogin()
    if err != nil {
        return nil, err
    }

    return &LoginData{Options: opts, Session: *session}, nil
}

// BeginLoginDirect starts a login ceremony using a fully populated User object
func (s *WebAuthnService) BeginLoginDirect(user *User) (*LoginData, error) {
    opts, session, err := s.webAuthn.BeginLogin(user)
    if err != nil {
        return nil, err
    }

    return &LoginData{Options: opts, Session: *session}, nil
}
