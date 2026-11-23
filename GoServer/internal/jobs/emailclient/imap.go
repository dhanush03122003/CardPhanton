package emailclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	jobtemporal "webauthn-server/internal/jobs/temporal"

	"github.com/emersion/go-imap"
	imapclient "github.com/emersion/go-imap/client"
	"github.com/emersion/go-message"
	_ "github.com/emersion/go-message/charset"
	"github.com/emersion/go-message/mail"
)

type IMAPClient struct {
	Host     string
	Password string
}

func NewIMAPClient(host, password string) *IMAPClient {
	if host == "" {
		host = "imap.gmail.com:993"
	}
	return &IMAPClient{Host: host, Password: password}
}

func (c *IMAPClient) connect(email string) (*imapclient.Client, error) {
	serverName := strings.Split(c.Host, ":")[0]
	conn, err := imapclient.DialTLS(c.Host, &tls.Config{ServerName: serverName, MinVersion: tls.VersionTLS12})
	if err != nil {
		return nil, fmt.Errorf("connect to imap: %w", err)
	}
	if email == "" {
		conn.Logout()
		return nil, fmt.Errorf("imap email is required")
	}
	if err := conn.Login(email, c.Password); err != nil {
		conn.Logout()
		return nil, fmt.Errorf("imap login: %w", err)
	}
	return conn, nil
}

func (c *IMAPClient) FetchRecentMetadata(ctx context.Context, email, bankName, cardLast4 string) ([]jobtemporal.EmailMetadata, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	log.Printf("[IMAP] Connecting to inbox for %s (Bank substring: %q, CardLast4: %q)...", email, bankName, cardLast4)
	conn, err := c.connect(email)
	if err != nil {
		return nil, err
	}
	defer conn.Logout()

	doneCtx := make(chan struct{})
	defer close(doneCtx)
	go func() {
		select {
		case <-ctx.Done():
			log.Printf("[IMAP] Context cancelled, terminating connection for %s", email)
			_ = conn.Terminate()
		case <-doneCtx:
		}
	}()

	if _, err := conn.Select("INBOX", true); err != nil {
		return nil, fmt.Errorf("select inbox: %w", err)
	}

	sinceDate := time.Now().AddDate(0, 0, -30)
	criteria := imap.NewSearchCriteria()
	criteria.Since = sinceDate

	log.Printf("[IMAP] Searching INBOX for all emails SINCE=%s...", sinceDate.Format("02-Jan-2006"))
	allUIDs, err := conn.UidSearch(criteria)
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("search inbox: %w", err)
	}

	log.Printf("[IMAP] Found %d total email(s) in last 30 days. Scanning senders for substring %q...", len(allUIDs), bankName)
	if len(allUIDs) == 0 {
		return []jobtemporal.EmailMetadata{}, nil
	}

	envSet := new(imap.SeqSet)
	envSet.AddNum(allUIDs...)
	envChan := make(chan *imap.Message, 50)
	envErrChan := make(chan error, 1)
	go func() {
		envErrChan <- conn.UidFetch(envSet, []imap.FetchItem{imap.FetchEnvelope, imap.FetchUid}, envChan)
	}()

	bankKeyword := strings.ToLower(strings.Fields(strings.TrimSpace(bankName))[0])
	var bankMatchedUIDs []uint32
	for msg := range envChan {
		if msg == nil || msg.Envelope == nil {
			continue
		}
		for _, addr := range msg.Envelope.From {
			fullSender := strings.ToLower(fmt.Sprintf("%s %s@%s", addr.PersonalName, addr.MailboxName, addr.HostName))
			if strings.Contains(fullSender, bankKeyword) {
				bankMatchedUIDs = append(bankMatchedUIDs, msg.Uid)
				break
			}
		}
	}

	if err := <-envErrChan; err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("fetch envelopes: %w", err)
	}

	log.Printf("[IMAP] Substring sender scan matched %d email(s) from %q", len(bankMatchedUIDs), bankKeyword)
	if len(bankMatchedUIDs) == 0 {
		return []jobtemporal.EmailMetadata{}, nil
	}
	if len(bankMatchedUIDs) > 60 {
		bankMatchedUIDs = bankMatchedUIDs[len(bankMatchedUIDs)-60:]
	}

	bodySet := new(imap.SeqSet)
	bodySet.AddNum(bankMatchedUIDs...)
	section := &imap.BodySectionName{Peek: true, Partial: []int{0, 131072}}
	messages := make(chan *imap.Message, 10)
	fetchDone := make(chan error, 1)
	go func() {
		fetchDone <- conn.UidFetch(bodySet, []imap.FetchItem{imap.FetchEnvelope, imap.FetchUid, section.FetchItem()}, messages)
	}()

	byUID := make(map[uint32]*imap.Message, len(bankMatchedUIDs))
	for message := range messages {
		if message != nil {
			byUID[message.Uid] = message
		}
	}
	if fetchErr := <-fetchDone; fetchErr != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		return nil, fmt.Errorf("fetch bodies: %w", fetchErr)
	}

	cardLower := strings.ToLower(strings.TrimSpace(cardLast4))
	result := make([]jobtemporal.EmailMetadata, 0, 20)
	for i := len(bankMatchedUIDs) - 1; i >= 0 && len(result) < 20; i-- {
		uid := bankMatchedUIDs[i]
		msg := byUID[uid]
		if msg == nil || msg.Envelope == nil {
			continue
		}

		senderStr := "unknown"
		if len(msg.Envelope.From) > 0 {
			f := msg.Envelope.From[0]
			senderStr = fmt.Sprintf("%s <%s@%s>", f.PersonalName, f.MailboxName, f.HostName)
		}

		bodyLiteral := msg.GetBody(section)
		if bodyLiteral == nil {
			for _, lit := range msg.Body {
				bodyLiteral = lit
				break
			}
		}
		body, err := extractPlainText(bodyLiteral)
		if err != nil {
			log.Printf("[IMAP] Warning: failed to parse MIME body for UID %d: %v", uid, err)
		}
		subject := msg.Envelope.Subject
		if cardLower != "" && !strings.Contains(strings.ToLower(body), cardLower) && !strings.Contains(strings.ToLower(subject), cardLower) {
			log.Printf("[IMAP] Skipping bank email UID %d | From: %s | Subject: %q (did not contain card %q)", uid, senderStr, subject, cardLast4)
			continue
		}

		item := jobtemporal.EmailMetadata{
			Index: len(result), UID: msg.Uid, Date: msg.Envelope.Date.Format(time.RFC3339), Subject: subject, Snippet: snippet(body),
		}
		log.Printf("[IMAP] MATCHED [%02d] | UID: %d | From: %s | Date: %s | Subject: %q | Snippet: %q", item.Index, item.UID, senderStr, msg.Envelope.Date.Format("02-Jan 15:04"), item.Subject, item.Snippet)
		result = append(result, item)
	}

	log.Printf("[IMAP] Successfully prepared %d matching email(s) for LLM Router", len(result))
	return result, nil
}

func (c *IMAPClient) FetchFullBodyByUID(ctx context.Context, email string, uid uint32) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	conn, err := c.connect(email)
	if err != nil {
		return "", err
	}
	defer conn.Logout()
	if _, err := conn.Select("INBOX", true); err != nil {
		return "", fmt.Errorf("select inbox: %w", err)
	}

	set := new(imap.SeqSet)
	set.AddNum(uid)
	section := &imap.BodySectionName{Peek: true}
	messages := make(chan *imap.Message, 1)
	if err := conn.UidFetch(set, []imap.FetchItem{imap.FetchUid, section.FetchItem()}, messages); err != nil {
		return "", fmt.Errorf("fetch uid %d: %w", uid, err)
	}
	for msg := range messages {
		if msg != nil {
			bodyLiteral := msg.GetBody(section)
			if bodyLiteral == nil {
				for _, lit := range msg.Body {
					bodyLiteral = lit
					break
				}
			}
			return extractPlainText(bodyLiteral)
		}
	}
	return "", fmt.Errorf("uid %d not found", uid)
}

func extractPlainText(body imap.Literal) (string, error) {
	if body == nil {
		return "", nil
	}
	rawBytes, err := io.ReadAll(body)
	if err != nil && len(rawBytes) == 0 {
		return "", err
	}
	reader, err := mail.CreateReader(bytes.NewReader(rawBytes))
	if err != nil && !message.IsUnknownCharset(err) {
		return fallbackExtractHTML(rawBytes), err
	}
	if reader == nil {
		return fallbackExtractHTML(rawBytes), err
	}

	var htmlParts []string
	var plainParts []string
	for {
		part, err := reader.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil && !message.IsUnknownCharset(err) {
			break
		}
		if part == nil {
			break
		}
		switch h := part.Header.(type) {
		case *mail.InlineHeader:
			contentType, _, _ := h.ContentType()
			partBytes, _ := io.ReadAll(part.Body)
			if len(partBytes) == 0 {
				continue
			}
			switch strings.ToLower(contentType) {
			case "text/html":
				if cleaned := strings.TrimSpace(extractTextFromHTML(partBytes)); cleaned != "" {
					htmlParts = append(htmlParts, cleaned)
				}
			case "text/plain":
				if cleaned := cleanText(string(partBytes)); cleaned != "" {
					plainParts = append(plainParts, cleaned)
				}
			}
		case *mail.AttachmentHeader:
			continue
		}
	}
	if len(htmlParts) > 0 {
		return strings.Join(htmlParts, "\n\n"), nil
	}
	if len(plainParts) > 0 {
		return strings.Join(plainParts, "\n\n"), nil
	}
	return fallbackExtractHTML(rawBytes), nil
}

func fallbackExtractHTML(rawBytes []byte) string {
	bodyStart := bytes.Index(rawBytes, []byte("\r\n\r\n"))
	if bodyStart != -1 {
		rawBytes = rawBytes[bodyStart+4:]
	}
	return strings.TrimSpace(extractTextFromHTML(rawBytes))
}
