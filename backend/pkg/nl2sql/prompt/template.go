package prompt

import (
	"fmt"
	"strings"

	"github.com/bxf1/ERP/backend/pkg/nl2sql/models"
)

// Template builds the system prompt for the LLM to convert NL to SQL.
type Template struct {
	Tables    []models.TableMeta
	Relations []models.Relationship
	Semantics []models.SemanticMapping
	TenantID  string
}

// NewTemplate creates a prompt template with all context injected.
func NewTemplate(tables []models.TableMeta, rels []models.Relationship, semantics []models.SemanticMapping, tenantID string) *Template {
	return &Template{
		Tables:    tables,
		Relations: rels,
		Semantics: semantics,
		TenantID:  tenantID,
	}
}

// BuildSystemPrompt constructs the full system prompt for the LLM.
func (t *Template) BuildSystemPrompt() string {
	var b strings.Builder

	b.WriteString("你是一个 PostgreSQL SQL 查询生成助手。根据用户的自然语言问题，生成正确的只读 SQL 查询。\n\n")

	b.WriteString("## 规则\n\n")
	b.WriteString("1. 只生成 SELECT 查询，禁止 INSERT / UPDATE / DELETE / DROP / TRUNCATE / ALTER / CREATE\n")
	b.WriteString("2. 所有查询必须包含租户过滤条件: WHERE tenant_id = '<tenant_id>'\n")
	b.WriteString("3. 如果查询涉及多张表，使用 JOIN 并确保 JOIN 条件正确\n")
	b.WriteString("4. 使用参数化列名，不要使用 SELECT *，明确列出需要的列\n")
	b.WriteString("5. 对于聚合查询，使用适当的 GROUP BY\n")
	b.WriteString("6. 限制返回行数最多 1000 行，使用 LIMIT 子句\n")
	b.WriteString("7. 日期字段使用 ISO 格式进行比较\n")
	b.WriteString("8. 不要包含注释或解释，只输出纯 SQL\n\n")

	b.WriteString(fmt.Sprintf("## 当前租户 ID: %s\n\n", t.TenantID))

	b.WriteString(FormatForPrompt(t.Tables, t.Relations, t.Semantics))

	b.WriteString("\n## 输出格式\n\n")
	b.WriteString("只输出一条 SQL 语句，以分号结尾。不要包含任何其他文字、注释或 markdown 代码块标记。")

	return b.String()
}

// BuildUserPrompt wraps the user's question with formatting instructions.
func BuildUserPrompt(question string) string {
	return fmt.Sprintf("用户问题: %s\n\n请生成对应的 PostgreSQL 查询语句。", question)
}
