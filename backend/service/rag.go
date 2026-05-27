package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/bxf1/ERP/backend/config"
	"github.com/bxf1/ERP/backend/pkg/rag/model"
	"github.com/bxf1/ERP/backend/pkg/embedding"
	"github.com/bxf1/ERP/backend/repository"
)

type RAGService struct {
	cfg        *config.Config
	embedder   embedding.Provider
	docRepo    *repository.KnowledgeDocRepository
	qaRepo     *repository.KnowledgeQARepository
}

func NewRAGService(
	cfg *config.Config,
	embedder embedding.Provider,
	docRepo *repository.KnowledgeDocRepository,
	qaRepo *repository.KnowledgeQARepository,
) *RAGService {
	return &RAGService{
		cfg:      cfg,
		embedder: embedder,
		docRepo:  docRepo,
		qaRepo:   qaRepo,
	}
}

type IngestDocInput struct {
	Title   string
	Content string
	Source  string
}

func (s *RAGService) IngestDocument(ctx context.Context, input IngestDocInput) error {
	chunks := splitText(input.Content, s.cfg.RAG.ChunkSize, s.cfg.RAG.ChunkOverlap)

	for i, chunk := range chunks {
		emb, err := s.embedder.Embed(ctx, chunk)
		if err != nil {
			return fmt.Errorf("embed chunk %d: %w", i, err)
		}

		doc := &model.KnowledgeDoc{
			Title:    input.Title,
			Content:  chunk,
			Source:   input.Source,
			ChunkIdx: i,
		}

		if err := s.docRepo.Create(doc, emb); err != nil {
			return fmt.Errorf("save chunk %d: %w", i, err)
		}
	}

	return nil
}

type IngestQAInput struct {
	Question string
	Answer   string
	Category string
}

func (s *RAGService) IngestQA(ctx context.Context, input IngestQAInput) error {
	text := input.Question + "\n" + input.Answer
	emb, err := s.embedder.Embed(ctx, text)
	if err != nil {
		return fmt.Errorf("embed qa: %w", err)
	}

	qa := &model.KnowledgeQA{
		Question: input.Question,
		Answer:   input.Answer,
		Category: input.Category,
	}

	return s.qaRepo.Create(qa, emb)
}

type SearchResult struct {
	Docs      []model.KnowledgeDoc
	QAs       []model.KnowledgeQA
}

func (s *RAGService) Search(ctx context.Context, query string) (*SearchResult, error) {
	emb, err := s.embedder.Embed(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("embed query: %w", err)
	}

	topK := s.cfg.RAG.TopK
	docs, err := s.docRepo.SearchSimilar(emb, topK)
	if err != nil {
		return nil, fmt.Errorf("search docs: %w", err)
	}

	qas, err := s.qaRepo.SearchSimilar(emb, topK)
	if err != nil {
		return nil, fmt.Errorf("search qas: %w", err)
	}

	return &SearchResult{Docs: docs, QAs: qas}, nil
}

func (s *RAGService) BuildContext(result *SearchResult) string {
	var parts []string

	if len(result.Docs) > 0 {
		parts = append(parts, "【相关文档】")
		for i, doc := range result.Docs {
			parts = append(parts, fmt.Sprintf("--- 文档片段 %d: %s ---\n%s", i+1, doc.Title, doc.Content))
		}
	}

	if len(result.QAs) > 0 {
		parts = append(parts, "\n【历史问答】")
		for i, qa := range result.QAs {
			parts = append(parts, fmt.Sprintf("--- 历史问答 %d ---\n问: %s\n答: %s", i+1, qa.Question, qa.Answer))
		}
	}

	return strings.Join(parts, "\n")
}

func (s *RAGService) GetDocCount() (int64, error) {
	return s.docRepo.Count()
}

func (s *RAGService) GetQACount() (int64, error) {
	return s.qaRepo.Count()
}

func (s *RAGService) DeleteDocument(id uint) error {
	return s.docRepo.Delete(id)
}

func (s *RAGService) DeleteQA(id uint) error {
	return s.qaRepo.Delete(id)
}

func splitText(text string, chunkSize, overlap int) []string {
	if len(text) <= chunkSize {
		return []string{text}
	}

	runes := []rune(text)
	var chunks []string

	for i := 0; i < len(runes); {
		end := i + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[i:end]))

		i += chunkSize - overlap
		if i >= len(runes) {
			break
		}
	}

	return chunks
}
