package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/urmzd/zoro/api/internal/model"
)

type OllamaClient struct {
	host           string
	model          string
	embeddingModel string
	httpClient     *http.Client
}

func NewOllamaClient(host, model, embeddingModel string) *OllamaClient {
	return &OllamaClient{
		host:           host,
		model:          model,
		embeddingModel: embeddingModel,
		httpClient: &http.Client{
			Timeout: 300 * time.Second,
		},
	}
}

type ollamaGenerateRequest struct {
	Model   string         `json:"model"`
	Prompt  string         `json:"prompt"`
	Stream  bool           `json:"stream"`
	Format  map[string]any `json:"format,omitempty"`
	Options map[string]any `json:"options,omitempty"`
}

type ollamaGenerateResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

type ollamaEmbedRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type ollamaEmbedResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
}

type extractedData struct {
	Entities []extractedEntity   `json:"entities"`
	Relations []extractedRelation `json:"relations"`
}

type extractedEntity struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Summary string `json:"summary"`
}

type extractedRelation struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Type   string `json:"type"`
	Fact   string `json:"fact"`
}

func (c *OllamaClient) Generate(ctx context.Context, prompt string) (string, error) {
	return c.GenerateWithModel(ctx, prompt, c.model)
}

type GenerateOptions struct {
	Format  map[string]any
	Options map[string]any
}

func (c *OllamaClient) GenerateWithModel(ctx context.Context, prompt string, model string, opts ...GenerateOptions) (string, error) {
	req := ollamaGenerateRequest{
		Model:  model,
		Prompt: prompt,
		Stream: false,
	}
	if len(opts) > 0 {
		req.Format = opts[0].Format
		req.Options = opts[0].Options
	}
	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshal ollama request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.host+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("ollama generate: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama returned %d: %s", resp.StatusCode, respBody)
	}

	var result ollamaGenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode ollama response: %w", err)
	}
	return result.Response, nil
}

func (c *OllamaClient) GenerateStream(ctx context.Context, prompt string) (<-chan string, error) {
	req := ollamaGenerateRequest{
		Model:  c.model,
		Prompt: prompt,
		Stream: true,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal ollama request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.host+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ollama generate stream: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("ollama returned %d", resp.StatusCode)
	}

	ch := make(chan string, 64)
	go func() {
		defer resp.Body.Close()
		defer close(ch)

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			var chunk ollamaGenerateResponse
			if err := json.Unmarshal(scanner.Bytes(), &chunk); err != nil {
				continue
			}
			if chunk.Response != "" {
				select {
				case ch <- chunk.Response:
				case <-ctx.Done():
					return
				}
			}
			if chunk.Done {
				return
			}
		}
	}()

	return ch, nil
}

func (c *OllamaClient) Embed(ctx context.Context, text string) ([]float32, error) {
	req := ollamaEmbedRequest{
		Model: c.embeddingModel,
		Input: text,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal embed request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.host+"/api/embed", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create embed request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ollama embed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama embed returned %d: %s", resp.StatusCode, respBody)
	}

	var result ollamaEmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode embed response: %w", err)
	}
	if len(result.Embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}
	return result.Embeddings[0], nil
}

func (c *OllamaClient) ChatStream(ctx context.Context, messages []model.OllamaChatMessage, tools []model.OllamaTool) (<-chan model.OllamaChatChunk, error) {
	req := model.OllamaChatRequest{
		Model:    c.model,
		Messages: messages,
		Tools:    tools,
		Stream:   true,
	}
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("marshal chat request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.host+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create chat request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("ollama chat stream: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("ollama chat returned %d: %s", resp.StatusCode, respBody)
	}

	ch := make(chan model.OllamaChatChunk, 64)
	go func() {
		defer resp.Body.Close()
		defer close(ch)

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 256*1024), 256*1024)
		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}
			var chunk model.OllamaChatChunk
			if err := json.Unmarshal(line, &chunk); err != nil {
				continue
			}
			select {
			case ch <- chunk:
			case <-ctx.Done():
				return
			}
			if chunk.Done {
				return
			}
		}
	}()

	return ch, nil
}

func (c *OllamaClient) ExtractEntities(ctx context.Context, text string) ([]extractedEntity, []extractedRelation, error) {
	prompt := `Extract entities and relationships from this text. Return ONLY valid JSON with no extra text:
{"entities": [{"name": "...", "type": "...", "summary": "..."}],
 "relations": [{"source": "...", "target": "...", "type": "...", "fact": "..."}]}

Text: ` + text

	raw, err := c.Generate(ctx, prompt)
	if err != nil {
		return nil, nil, fmt.Errorf("extract entities: %w", err)
	}

	// Try to find JSON in the response
	jsonStr := raw
	if start := strings.Index(raw, "{"); start != -1 {
		if end := strings.LastIndex(raw, "}"); end != -1 {
			jsonStr = raw[start : end+1]
		}
	}

	var data extractedData
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, nil, fmt.Errorf("parse extraction response: %w (raw: %s)", err, raw)
	}
	return data.Entities, data.Relations, nil
}
