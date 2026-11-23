package emailclient

import (
	"context"
	"encoding/base64"
	"fmt"
	"hash/fnv"
	"log"
	"net/mail"
	"strings"
	"sync"
	"time"

	jobtemporal "webauthn-server/internal/jobs/temporal"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
)

// RefreshTokenFunc is the signature for your existing function that fetches a user's refresh token from DB.
type RefreshTokenFunc func(ctx context.Context, email string) (string, error)

// GoogleClient is the Gmail API implementation of EmailClient.
type GoogleClient struct {
	oauthConfig     *oauth2.Config
	getRefreshToken RefreshTokenFunc
	uidToGmailID    sync.Map // Bridges uint32 UID -> Gmail string Message ID
}

// NewGoogleClient initializes the Gmail API client using your OAuth credentials
// and your existing function that retrieves the user's refresh token.
func NewGoogleClient(clientID, clientSecret, redirectURL string, getRefreshToken RefreshTokenFunc) *GoogleClient {
	return &GoogleClient{
		oauthConfig: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       []string{gmail.GmailReadonlyScope},
			Endpoint:     google.Endpoint,
		},
		getRefreshToken: getRefreshToken,
	}
}

// getService creates an authenticated *gmail.Service for the given user email
// by fetching their refresh token and automatically minting a fresh access token.
func (c *GoogleClient) getService(ctx context.Context, email string) (*gmail.Service, error) {
	if c.getRefreshToken == nil {
		return nil, fmt.Errorf("getRefreshToken function is not configured")
	}

	refreshToken, err := c.getRefreshToken(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("fetch refresh token for %s: %w", email, err)
	}
	if refreshToken == "" {
		return nil, fmt.Errorf("no refresh token found for %s (user consent required)", email)
	}

	// oauth2.TokenSource automatically exchanges the RefreshToken for a new AccessToken
	tokenSource := c.oauthConfig.TokenSource(ctx, &oauth2.Token{
		RefreshToken: refreshToken,
	})

	svc, err := gmail.NewService(ctx, option.WithTokenSource(tokenSource))
	if err != nil {
		return nil, fmt.Errorf("create gmail service: %w", err)
	}
	return svc, nil
}

func (c *GoogleClient) FetchRecentMetadata(ctx context.Context, email, bankName, cardLast4 string) ([]jobtemporal.EmailMetadata, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	svc, err := c.getService(ctx, email)
	if err != nil {
		return nil, err
	}

	// Search recent mail broadly, then apply sender and card matching locally.
	bankFields := strings.Fields(strings.TrimSpace(bankName))
	bankKeyword := ""
	if len(bankFields) > 0 {
		bankKeyword = strings.ToLower(bankFields[0])
	}
	cardTrimmed := strings.TrimSpace(cardLast4)
queryParts := []string{
		"newer_than:30d",
		"label:INBOX",      // Must currently be in the inbox (not archived)
		"category:primary", // Must be in the Primary tab (not Promotions/Updates)
	}

	query := strings.Join(queryParts, " ")

	log.Printf("[Gmail API] Searching inbox for %s with query: %q", email, query)

	// Gmail returns messages ordered from newest to oldest by default
	listResp, err := svc.Users.Messages.List("me").
		Q(query).
		MaxResults(100).
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("gmail list messages: %w", err)
	}

	if len(listResp.Messages) == 0 {
		log.Printf("[Gmail API] No matching emails found for query: %q", query)
		return []jobtemporal.EmailMetadata{}, nil
	}

	cardLower := strings.ToLower(cardTrimmed)
	result := make([]jobtemporal.EmailMetadata, 0, len(listResp.Messages))

	for _, msgSummary := range listResp.Messages {
		msg, err := svc.Users.Messages.Get("me", msgSummary.Id).
			Format("full").
			Context(ctx).
			Do()
		if err != nil {
			log.Printf("[Gmail API] Warning: failed to fetch message %s: %v", msgSummary.Id, err)
			continue
		}

		subject := getHeader(msg.Payload.Headers, "Subject")
		from := getHeader(msg.Payload.Headers, "From")
		dateStr := formatGmailDate(msg)

		bodyText := extractBodyFromPayload(msg.Payload)
		if bodyText == "" {
			bodyText = msg.Snippet
		}

		if bankKeyword != "" && !strings.Contains(strings.ToLower(from), bankKeyword) {
			continue
		}

		// Double-check cardLast4 substring match in body or subject
		if cardLower != "" &&
			!strings.Contains(strings.ToLower(bodyText), cardLower) &&
			!strings.Contains(strings.ToLower(subject), cardLower) {
			continue
		}

		// Convert Gmail string ID to a deterministic uint32 and store in map for FetchFullBodyByUID
		uid := hashGmailIDToUint32(msg.Id)
		c.uidToGmailID.Store(uid, msg.Id)

		item := jobtemporal.EmailMetadata{
			Index:   len(result),
			UID:     uid,
			Date:    dateStr,
			Subject: subject,
			Snippet: snippet(bodyText),
		}

		log.Printf("[Gmail API] MATCHED [%02d] | UID: %d (GmailID: %s) | From: %s | Date: %s | Subject: %q",
			item.Index, item.UID, msg.Id, from, item.Date, item.Subject)

		result = append(result, item)
	}

	return result, nil
}

func (c *GoogleClient) FetchFullBodyByUID(ctx context.Context, email string, uid uint32) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	val, ok := c.uidToGmailID.Load(uid)
	if !ok {
		return "", fmt.Errorf("gmail message ID not found in cache for uid %d", uid)
	}
	gmailMsgID := val.(string)

	svc, err := c.getService(ctx, email)
	if err != nil {
		return "", err
	}

	msg, err := svc.Users.Messages.Get("me", gmailMsgID).
		Format("full").
		Context(ctx).
		Do()
	if err != nil {
		return "", fmt.Errorf("gmail get full message %s (uid %d): %w", gmailMsgID, uid, err)
	}

	bodyText := extractBodyFromPayload(msg.Payload)
	if bodyText == "" {
		return msg.Snippet, nil
	}
	return bodyText, nil
}

// -----------------------------------------------------------------------------
// Helper Functions: MIME Tree Traversal, Base64URL Decoding, & HTML Stripping
// -----------------------------------------------------------------------------

// extractBodyFromPayload walks the Gmail MIME tree, preferring text/html over text/plain.
func extractBodyFromPayload(part *gmail.MessagePart) string {
	if part == nil {
		return ""
	}

	var htmlParts []string
	var plainParts []string

	var walkParts func(p *gmail.MessagePart)
	walkParts = func(p *gmail.MessagePart) {
		if p == nil {
			return
		}

		// Skip attachments
		if p.Filename != "" {
			return
		}

		if p.Body != nil && p.Body.Data != "" {
			decoded, err := decodeBase64URL(p.Body.Data)
			if err == nil && len(decoded) > 0 {
				switch strings.ToLower(p.MimeType) {
				case "text/html":
					cleaned := strings.TrimSpace(extractTextFromHTML(decoded))
					if cleaned != "" {
						htmlParts = append(htmlParts, cleaned)
					}
				case "text/plain":
					cleaned := cleanText(string(decoded))
					if cleaned != "" {
						plainParts = append(plainParts, cleaned)
					}
				}
			}
		}

		for _, subPart := range p.Parts {
			walkParts(subPart)
		}
	}

	walkParts(part)

	if len(htmlParts) > 0 {
		return strings.Join(htmlParts, "\n\n")
	}
	if len(plainParts) > 0 {
		return strings.Join(plainParts, "\n\n")
	}
	return ""
}

// decodeBase64URL decodes Gmail's URL-safe base64 payload data.
func decodeBase64URL(data string) ([]byte, error) {
	// Gmail uses RawURLEncoding (no padding), fallback to URLEncoding if padded
	if decoded, err := base64.RawURLEncoding.DecodeString(data); err == nil {
		return decoded, nil
	}
	return base64.URLEncoding.DecodeString(data)
}

func getHeader(headers []*gmail.MessagePartHeader, name string) string {
	for _, h := range headers {
		if strings.EqualFold(h.Name, name) {
			return h.Value
		}
	}
	return ""
}

func formatGmailDate(msg *gmail.Message) string {
	if msg.InternalDate > 0 {
		return time.UnixMilli(msg.InternalDate).Format(time.RFC3339)
	}
	rawDate := getHeader(msg.Payload.Headers, "Date")
	if parsed, err := mail.ParseDate(rawDate); err == nil {
		return parsed.Format(time.RFC3339)
	}
	return time.Now().Format(time.RFC3339)
}

func hashGmailIDToUint32(gmailID string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(gmailID))
	return h.Sum32()
}
