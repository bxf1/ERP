package prompt

import (
	"strings"
	"testing"

	"github.com/bxf1/ERP/backend/pkg/nl2sql/models"
)

func TestFormatForPrompt(t *testing.T) {
	tables := []models.TableMeta{
		{
			Name:        "orders",
			Description: "销售订单表",
			PrimaryKey:  "id",
			Columns: []models.ColumnMeta{
				{Name: "id", Type: "uuid", Nullable: false, Description: "主键"},
				{Name: "tenant_id", Type: "uuid", Nullable: false, Description: "租户ID"},
				{Name: "total_amount", Type: "numeric", Nullable: false, Description: "订单金额"},
			},
		},
	}
	rels := []models.Relationship{}
	semantics := []models.SemanticMapping{}

	result := FormatForPrompt(tables, rels, semantics)

	if !strings.Contains(result, "orders") {
		t.Error("expected table name in prompt")
	}
	if !strings.Contains(result, "total_amount") {
		t.Error("expected column name in prompt")
	}
	if !strings.Contains(result, "销售订单表") {
		t.Error("expected table description in prompt")
	}
}

func TestBuildSystemPrompt(t *testing.T) {
	tables := []models.TableMeta{
		{Name: "test_table", Columns: []models.ColumnMeta{{Name: "id", Type: "int"}}},
	}
	tmpl := NewTemplate(tables, nil, nil, "t-001")
	prompt := tmpl.BuildSystemPrompt()

	if !strings.Contains(prompt, "t-001") {
		t.Error("expected tenant_id in system prompt")
	}
	if !strings.Contains(prompt, "SELECT") {
		t.Error("expected SELECT instruction in prompt")
	}
	if !strings.Contains(prompt, "test_table") {
		t.Error("expected table info in prompt")
	}
}

func TestMergeSemantics(t *testing.T) {
	defaults := []models.SemanticMapping{
		{BusinessTerm: "default_term", SQLFragment: "COUNT(*)"},
	}
	customs := []models.SemanticMapping{
		{BusinessTerm: "custom_term", SQLFragment: "SUM(x)"},
	}

	result := MergeSemantics(defaults, customs)
	if len(result) != 2 {
		t.Errorf("expected 2 semantics, got %d", len(result))
	}
}

func TestMergeSemantics_CustomOverridesDefault(t *testing.T) {
	defaults := []models.SemanticMapping{
		{BusinessTerm: "sales", SQLFragment: "COUNT(*)"},
	}
	customs := []models.SemanticMapping{
		{BusinessTerm: "sales", SQLFragment: "SUM(amount)"},
	}

	result := MergeSemantics(defaults, customs)
	if len(result) != 1 {
		t.Errorf("expected 1 semantic after override, got %d", len(result))
	}
	if result[0].SQLFragment != "SUM(amount)" {
		t.Errorf("expected custom override to take precedence")
	}
}
