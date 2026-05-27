package embedding

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type OpenaiProvider struct {
	apiKey     string
	baseURL    string
	model      string
	dimensions int
	client     *http.Client
}

type embeddingRequest struct {
	Input []string `json:"input"`
	Model string   `json:"model"`
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
}

func NewOpenaiProvider(apiKey, baseURL, model string) *OpenaiProvider {
	return &OpenaiProvider{
		apiKey:     apiKey,
		baseURL:    baseURL,
		model:      model,
		dimensions: 1536,
		client:     &http.Client{Timeout: 60 * time.Second},
	}
}

func (p *OpenaiProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	embeddings, err := p.EmbedBatch(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(embeddings) == 0 {
		return nil, fmt.Errorf("empty embedding response")
	}
	return embeddings[0], nil
}

func (p *OpenaiProvider) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	reqBody := embeddingRequest{
		Input: texts,
		Model: p.model,
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/embeddings", p.baseURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("embedding API error %d: %s", resp.StatusCode, string(body))
	}

	var embResp embeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	result := make([][]float32, len(embResp.Data))
	for _, d := range embResp.Data {
		if d.Index >= len(result) {
			return nil, fmt.Errorf("unexpected embedding index %d", d.Index)
		}
		result[d.Index] = d.Embedding
	}

	return result, nil
}

func (p *OpenaiProvider) Dimensions() int {
	return p.dimensions
}
