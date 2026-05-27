package semantic

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/bxf1/ERP/backend/pkg/datadict"
)

// ContextBuilder assembles the LLM prompt context from both the data dictionary
// and the semantic layer configuration, producing a compact representation that
// an NL2SQL model can consume as a system prompt.
type ContextBuilder struct {
	dictSvc *datadict.Service
	cfg     *Config
}

// NewContextBuilder creates a ContextBuilder that merges data dictionary
// tables/columns with semantic-layer metrics/dimensions.
func NewContextBuilder(dictSvc *datadict.Service, cfg *Config) *ContextBuilder {
	return &ContextBuilder{dictSvc: dictSvc, cfg: cfg}
}

// Build produces the full LLMContext from cached schema and config.
func (cb *ContextBuilder) Build() (*LLMContext, error) {
	ctx := &LLMContext{GeneratedAt: time.Now().UTC()}

	// Merge schema tables.
	dict, err := cb.dictSvc.GetSchema(context.Background())
	if err != nil {
		return nil, fmt.Errorf("get schema: %w", err)
	}
	for _, t := range dict.Tables {
		tRef := TableRef{Name: t.Name}
		for _, c := range t.Columns {
			tRef.Columns = append(tRef.Columns, ColumnRef{
				Name:    c.Name,
				Type:    c.DataType,
				Comment: c.Comment,
			})
		}
		ctx.Tables = append(ctx.Tables, tRef)
	}

	// Collect metrics and dimensions across all models.
	builder := NewSQLBuilder("postgres")
	for _, m := range cb.cfg.Models {
		for _, metric := range m.Metrics {
			ctx.Metrics = append(ctx.Metrics, MetricRef{
				Name:        metric.Name,
				Label:       metric.Label,
				Description: metric.Description,
				SQL:         builder.BuildMetricSQL(metric),
			})
		}
		for _, dim := range m.Dimensions {
			ctx.Dimensions = append(ctx.Dimensions, dim)
		}
	}

	// Build join hints from foreign keys.
	for _, rel := range dict.Relations {
		ctx.Joins = append(ctx.Joins, JoinHint{
			FromTable: rel.SourceTable,
			ToTable:   rel.TargetTable,
			Condition: fmt.Sprintf("%s.%s = %s.%s",
				rel.SourceTable, rel.SourceColumn,
				rel.TargetTable, rel.TargetColumn),
		})
	}

	return ctx, nil
}

// PromptFragment returns a string suitable for injection into an LLM system
// prompt. It includes the table schemas, available metrics, and join paths.
func (cb *ContextBuilder) PromptFragment() (string, error) {
	llmCtx, err := cb.Build()
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString("## Database Schema\n\n")
	for _, t := range llmCtx.Tables {
		b.WriteString(fmt.Sprintf("### %s\n", t.Name))
		for _, c := range t.Columns {
			comment := ""
			if c.Comment != "" {
				comment = fmt.Sprintf(" -- %s", c.Comment)
			}
			b.WriteString(fmt.Sprintf("- %s (%s)%s\n", c.Name, c.Type, comment))
		}
		b.WriteString("\n")
	}

	b.WriteString("## Available Metrics\n\n")
	for _, m := range llmCtx.Metrics {
		b.WriteString(fmt.Sprintf("- **%s** (%s): %s → `%s`\n", m.Label, m.Name, m.Description, m.SQL))
	}

	b.WriteString("\n## Table Relationships\n\n")
	for _, j := range llmCtx.Joins {
		b.WriteString(fmt.Sprintf("- %s → %s ON %s\n", j.FromTable, j.ToTable, j.Condition))
	}

	return b.String(), nil
}
