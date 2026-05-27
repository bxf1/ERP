package repository

import (
	"fmt"
	"strings"

	"github.com/bxf1/ERP/backend/pkg/rag/model"
	"gorm.io/gorm"
)

type KnowledgeDocRepository struct {
	db *gorm.DB
}

func NewKnowledgeDocRepository(db *gorm.DB) *KnowledgeDocRepository {
	return &KnowledgeDocRepository{db: db}
}

func (r *KnowledgeDocRepository) AutoMigrate(dimensions int) error {
	if err := r.db.AutoMigrate(&model.KnowledgeDoc{}); err != nil {
		return err
	}
	return r.db.Exec(
		fmt.Sprintf(`
			DO $$ BEGIN
				IF NOT EXISTS (
					SELECT 1 FROM information_schema.columns
					WHERE table_name = 'knowledge_docs' AND column_name = 'embedding'
				) THEN
					ALTER TABLE knowledge_docs ADD COLUMN embedding vector(%d);
				END IF;
			END $$;
		`, dimensions),
	).Error
}

func (r *KnowledgeDocRepository) Create(doc *model.KnowledgeDoc, embedding []float32) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(doc).Error; err != nil {
			return err
		}
		return tx.Exec(
			"UPDATE knowledge_docs SET embedding = $1::vector WHERE id = $2",
			vectorToString(embedding), doc.ID,
		).Error
	})
}

func (r *KnowledgeDocRepository) FindByID(id uint) (*model.KnowledgeDoc, error) {
	var doc model.KnowledgeDoc
	err := r.db.First(&doc, id).Error
	return &doc, err
}

func (r *KnowledgeDocRepository) FindAll() ([]model.KnowledgeDoc, error) {
	var docs []model.KnowledgeDoc
	err := r.db.Order("id DESC").Find(&docs).Error
	return docs, err
}

func (r *KnowledgeDocRepository) Delete(id uint) error {
	return r.db.Delete(&model.KnowledgeDoc{}, id).Error
}

func (r *KnowledgeDocRepository) SearchSimilar(embedding []float32, topK int) ([]model.KnowledgeDoc, error) {
	var docs []model.KnowledgeDoc
	err := r.db.Raw(`
		SELECT id, created_at, updated_at, deleted_at, title, content, source, chunk_idx
		FROM knowledge_docs
		WHERE deleted_at IS NULL AND embedding IS NOT NULL
		ORDER BY embedding <=> $1::vector
		LIMIT $2
	`, vectorToString(embedding), topK).Scan(&docs).Error
	return docs, err
}

func (r *KnowledgeDocRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.KnowledgeDoc{}).Count(&count).Error
	return count, err
}

type KnowledgeQARepository struct {
	db *gorm.DB
}

func NewKnowledgeQARepository(db *gorm.DB) *KnowledgeQARepository {
	return &KnowledgeQARepository{db: db}
}

func (r *KnowledgeQARepository) AutoMigrate(dimensions int) error {
	if err := r.db.AutoMigrate(&model.KnowledgeQA{}); err != nil {
		return err
	}
	return r.db.Exec(
		fmt.Sprintf(`
			DO $$ BEGIN
				IF NOT EXISTS (
					SELECT 1 FROM information_schema.columns
					WHERE table_name = 'knowledge_qas' AND column_name = 'embedding'
				) THEN
					ALTER TABLE knowledge_qas ADD COLUMN embedding vector(%d);
				END IF;
			END $$;
		`, dimensions),
	).Error
}

func (r *KnowledgeQARepository) Create(qa *model.KnowledgeQA, embedding []float32) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(qa).Error; err != nil {
			return err
		}
		return tx.Exec(
			"UPDATE knowledge_qas SET embedding = $1::vector WHERE id = $2",
			vectorToString(embedding), qa.ID,
		).Error
	})
}

func (r *KnowledgeQARepository) FindAll() ([]model.KnowledgeQA, error) {
	var qas []model.KnowledgeQA
	err := r.db.Order("id DESC").Find(&qas).Error
	return qas, err
}

func (r *KnowledgeQARepository) Delete(id uint) error {
	return r.db.Delete(&model.KnowledgeQA{}, id).Error
}

func (r *KnowledgeQARepository) SearchSimilar(embedding []float32, topK int) ([]model.KnowledgeQA, error) {
	var qas []model.KnowledgeQA
	err := r.db.Raw(`
		SELECT id, created_at, updated_at, deleted_at, question, answer, category
		FROM knowledge_qas
		WHERE deleted_at IS NULL AND embedding IS NOT NULL
		ORDER BY embedding <=> $1::vector
		LIMIT $2
	`, vectorToString(embedding), topK).Scan(&qas).Error
	return qas, err
}

func (r *KnowledgeQARepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.KnowledgeQA{}).Count(&count).Error
	return count, err
}

func vectorToString(vec []float32) string {
	parts := make([]string, len(vec))
	for i, v := range vec {
		parts[i] = fmt.Sprintf("%f", v)
	}
	return "[" + strings.Join(parts, ",") + "]"
}
