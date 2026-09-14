package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/taufiqgit/news-api/internal/domain/entity"
)

// Config adalah konfigurasi untuk provider [OI]-compatible.
type Config struct {
	BaseURL     string  // mis. https://api.openai.com/v1
	APIKey      string  // dari env, tidak pernah di-hardcode
	Model       string  // mis. gpt-4o-mini
	MaxTokens   int     // 0 = default provider
	Temperature float64 // ≤ 0.7
	Timeout     time.Duration
}

// oaiClient mengimplementasikan Rewriter untuk endpoint [OI]-compatible
// (Groq, OpenRouter, Together, Ollama, dsb.) via chat completions API.
type oaiClient struct {
	cfg    Config
	client *http.Client
}

// NewOpenAICompatible membuat rewriter berbasis endpoint [OI]-compatible
// (Groq, OpenRouter, Together, Ollama, dsb.) via chat completions API.
func NewOpenAICompatible(cfg Config) Rewriter {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	cfg.BaseURL = baseURL

	return &oaiClient{
		cfg:    cfg,
		client: &http.Client{Timeout: timeout},
	}
}

func (c *oaiClient) Rewrite(ctx context.Context, req entity.RewriteRequest) (*entity.RewriteResult, error) {
	// Coba hingga 2x bila respons JSON tidak valid.
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		res, err := c.call(ctx, req)
		if err != nil {
			return nil, err
		}
		result, err := parseResult(res)
		if err == nil {
			return result, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("ai: invalid response after retries: %w", lastErr)
}

func (c *oaiClient) call(ctx context.Context, req entity.RewriteRequest) (string, error) {
	body := chatRequest{
		Model: c.cfg.Model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: buildUserPrompt(req.Title, req.Content, req.SourceName, req.SourceURL, req.WebsiteName)},
		},
	}
	if c.cfg.MaxTokens > 0 {
		body.MaxTokens = c.cfg.MaxTokens
	}
	if c.cfg.Temperature != 0 {
		body.Temperature = c.cfg.Temperature
	}

	payload, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("ai: marshal request: %w", err)
	}

	url := c.cfg.BaseURL + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("ai: build request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.cfg.APIKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("ai: call %s: %w", url, err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1MB cap
	if err != nil {
		return "", fmt.Errorf("ai: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		// Potong body untuk error agar tidak membocorkan terlalu banyak.
		snippet := string(respBody)
		if len(snippet) > 512 {
			snippet = snippet[:512]
		}
		return "", fmt.Errorf("ai: %s returned status %d: %s", url, resp.StatusCode, snippet)
	}

	var out chatResponse
	if err := json.Unmarshal(respBody, &out); err != nil {
		return "", fmt.Errorf("ai: decode response: %w", err)
	}
	if len(out.Choices) == 0 {
		return "", fmt.Errorf("ai: empty choices in response")
	}
	return out.Choices[0].Message.Content, nil
}

// parseResult mem-parsing JSON hasil AI menjadi RewriteResult.
// Menerima body yang mungkin dibungkus markdown fences.
func parseResult(raw string) (*entity.RewriteResult, error) {
	jsonStr := stripFences(raw)

	var r entity.RewriteResult
	if err := json.Unmarshal([]byte(jsonStr), &r); err != nil {
		return nil, fmt.Errorf("ai: decode article JSON: %w", err)
	}
	if strings.TrimSpace(r.Title) == "" || strings.TrimSpace(r.Content) == "" {
		return nil, fmt.Errorf("ai: article missing title or content")
	}
	return &r, nil
}

// stripFences menghapus pembungkus markdown ```json ... ``` bila ada.
func stripFences(s string) string {
	s = strings.TrimSpace(s)
	lower := strings.ToLower(s)
	if strings.HasPrefix(lower, "```json") {
		s = strings.TrimPrefix(s, s[:7])
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
	}
	s = strings.TrimSuffix(strings.TrimSpace(s), "```")
	return strings.TrimSpace(s)
}

// --- Tipe request/response [OI]-compatible ---

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model       string        `json:"model"`
	Messages    []chatMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}
