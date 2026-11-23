package llm

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"text/template"

	jobtemporal "webauthn-server/internal/jobs/temporal"
)

type Config struct {
	Provider, GeminiAPIKey, OllamaBaseURL, ModelName string
	HTTPClient                                       *http.Client
}
type LLMClient interface {
	RouteEmails(context.Context, []jobtemporal.EmailMetadata) ([]int, error)
	ExtractCardLimits(context.Context, string) (*jobtemporal.CardEvent, error)
}
type Client struct {
	config     Config
	httpClient *http.Client
}

//go:embed prompts/*.tmpl
var promptFiles embed.FS

var promptTemplates = template.Must(template.ParseFS(promptFiles, "prompts/*.tmpl"))

type categorizedRouterOutput struct {
	TransactionAlert *int `json:"transaction_alert"`
	MonthlyStatement *int `json:"monthly_statement"`
	PaymentReminder  *int `json:"payment_reminder"`
}

func NewClient(config Config) *Client {
	hc := config.HTTPClient
	if hc == nil {
		hc = http.DefaultClient
	}
	if config.OllamaBaseURL == "" {
		config.OllamaBaseURL = "http://localhost:11434"
	}
	return &Client{config: config, httpClient: hc}
}
func (c *Client) RouteEmails(ctx context.Context, emails []jobtemporal.EmailMetadata) ([]int, error) {
	payload, err := json.Marshal(emails)
	if err != nil {
		return nil, err
	}
	instruction, err := renderPrompt("routing.tmpl", nil)
	if err != nil {
		return nil, err
	}

	var categorized categorizedRouterOutput
	if err := c.completeJSON(ctx, instruction, string(payload), &categorized); err != nil {
		return nil, err
	}
	log.Printf("[LLM] Router classification: transaction_alert=%v monthly_statement=%v payment_reminder=%v", categorized.TransactionAlert, categorized.MonthlyStatement, categorized.PaymentReminder)

	// Collect non-nil indices in a deterministic order (max 3 total, 1 per category)
	seen := make(map[int]bool)
	var selected []int
	for _, ptr := range []*int{categorized.TransactionAlert, categorized.MonthlyStatement, categorized.PaymentReminder} {
		if ptr != nil && *ptr >= 0 && *ptr < len(emails) && !seen[*ptr] {
			seen[*ptr] = true
			selected = append(selected, *ptr)
		}
	}
	if len(selected) == 0 {
		return nil, fmt.Errorf("LLM did not select any matching email indices")
	}

	return selected, nil
}
func (c *Client) ExtractCardLimits(ctx context.Context, body string) (*jobtemporal.CardEvent, error) {
	instruction, err := renderPrompt("extract_card_limits.tmpl", nil)
	if err != nil {
		return nil, err
	}
	var event *jobtemporal.CardEvent
	if err := c.completeJSON(ctx, instruction, body, &event); err != nil {
		return nil, err
	}
	return event, nil
}

func renderPrompt(name string, data any) (string, error) {
	var rendered bytes.Buffer
	if err := promptTemplates.ExecuteTemplate(&rendered, name, data); err != nil {
		return "", fmt.Errorf("render prompt %q: %w", name, err)
	}
	return strings.TrimSpace(rendered.String()), nil
}

func (c *Client) completeJSON(ctx context.Context, system, user string, target any) error {
	if strings.EqualFold(c.config.Provider, "ollama") {
		return c.ollama(ctx, system, user, target)
	}
	if strings.EqualFold(c.config.Provider, "gemini") {
		return c.gemini(ctx, system, user, target)
	}
	return fmt.Errorf("unsupported llm provider %q", c.config.Provider)
}
func (c *Client) ollama(ctx context.Context, system, user string, target any) error {
	request := map[string]any{"model": c.config.ModelName, "prompt": system + "\n\n" + user, "stream": false, "format": "json"}
	var response struct {
		Response string `json:"response"`
	}
	if err := c.postJSON(ctx, strings.TrimRight(c.config.OllamaBaseURL, "/")+"/api/generate", request, &response); err != nil {
		return err
	}
	return decodeStrict(response.Response, target)
}
func (c *Client) gemini(ctx context.Context, system, user string, target any) error {
	endpoint := "https://generativelanguage.googleapis.com/v1beta/models/" + c.config.ModelName + ":generateContent?key=" + c.config.GeminiAPIKey
	request := map[string]any{
		"system_instruction": map[string]any{"parts": []map[string]string{{"text": system}}},
		"contents":           []map[string]any{{"role": "user", "parts": []map[string]string{{"text": user}}}},
		"generationConfig":   map[string]string{"responseMimeType": "application/json"},
	}
	var response struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := c.postJSON(ctx, endpoint, request, &response); err != nil {
		return err
	}
	if len(response.Candidates) == 0 || len(response.Candidates[0].Content.Parts) == 0 {
		return fmt.Errorf("gemini returned no candidates")
	}
	return decodeStrict(response.Candidates[0].Content.Parts[0].Text, target)
}
func (c *Client) postJSON(ctx context.Context, endpoint string, request, response any) error {
	data, err := json.Marshal(request)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return fmt.Errorf("llm request failed with status %d: %s", res.StatusCode, strings.TrimSpace(string(raw)))
	}
	if err := json.Unmarshal(raw, response); err != nil {
		return fmt.Errorf("decode llm response: %w", err)
	}
	return nil
}
func decodeStrict(value string, target any) error {
	value = strings.TrimSpace(value)
	if start := strings.Index(value, "{"); start > 0 {
		value = value[start:]
	}
	if end := strings.LastIndex(value, "}"); end >= 0 {
		value = value[:end+1]
	}
	decoder := json.NewDecoder(strings.NewReader(value))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}
