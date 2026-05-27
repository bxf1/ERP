package embedding

import "context"

type MockProvider struct {
	dimensions int
}

func NewMockProvider(dimensions int) *MockProvider {
	if dimensions <= 0 {
		dimensions = 256
	}
	return &MockProvider{dimensions: dimensions}
}

func (p *MockProvider) Embed(ctx context.Context, text string) ([]float32, error) {
	return p.simpleHashEmbed(text), nil
}

func (p *MockProvider) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	result := make([][]float32, len(texts))
	for i, t := range texts {
		result[i] = p.simpleHashEmbed(t)
	}
	return result, nil
}

func (p *MockProvider) Dimensions() int {
	return p.dimensions
}

func (p *MockProvider) simpleHashEmbed(text string) []float32 {
	vec := make([]float32, p.dimensions)
	runes := []rune(text)
	for i := 0; i < p.dimensions; i++ {
		if i < len(runes) {
			vec[i] = float32(runes[i%len(runes)]) / 65536.0
		} else {
			vec[i] = float32(runes[i%len(runes)]>>8) / 256.0
		}
	}
	return vec
}
