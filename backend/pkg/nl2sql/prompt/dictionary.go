package prompt

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/bxf1/ERP/backend/pkg/nl2sql/models"
)

// Dictionary loads schema metadata from the database.
type Dictionary struct {
	db *sql.DB
}

func NewDictionary(db *sql.DB) *Dictionary {
	return &Dictionary{db: db}
}

// LoadTables extracts table and column metadata from information_schema.
func (d *Dictionary) LoadTables(schema string) ([]models.TableMeta, error) {
	query := `
		SELECT t.table_name,
		       COALESCE(obj_description(pc.oid), '') AS table_comment
		FROM information_schema.tables t
		JOIN pg_catalog.pg_class pc ON pc.relname = t.table_name
		WHERE t.table_schema = $1
		  AND t.table_type = 'BASE TABLE'
		ORDER BY t.table_name`

	rows, err := d.db.Query(query, schema)
	if err != nil {
		return nil, fmt.Errorf("query tables: %w", err)
	}
	defer rows.Close()

	var tables []models.TableMeta
	for rows.Next() {
		var name, desc string
		if err := rows.Scan(&name, &desc); err != nil {
			return nil, fmt.Errorf("scan table: %w", err)
		}

		cols, err := d.loadColumns(schema, name)
		if err != nil {
			return nil, err
		}

		pk, _ := d.loadPrimaryKey(schema, name)

		tables = append(tables, models.TableMeta{
			Name:        name,
			Description: desc,
			Columns:     cols,
			PrimaryKey:  pk,
		})
	}
	return tables, nil
}

func (d *Dictionary) loadColumns(schema, table string) ([]models.ColumnMeta, error) {
	query := `
		SELECT c.column_name,
		       c.data_type,
		       c.is_nullable,
		       COALESCE(pgd.description, '') AS col_comment,
		       CASE WHEN c.column_name LIKE '%_id' AND c.column_name != 'tenant_id' THEN true ELSE false END
		FROM information_schema.columns c
		LEFT JOIN pg_catalog.pg_description pgd
		       ON pgd.objoid = (SELECT pc.oid FROM pg_catalog.pg_class pc
		                        JOIN pg_catalog.pg_namespace pn ON pn.oid = pc.relnamespace
		                        WHERE pc.relname = $2 AND pn.nspname = $1)
		      AND pgd.objsubid = c.ordinal_position
		WHERE c.table_schema = $1 AND c.table_name = $2
		ORDER BY c.ordinal_position`

	rows, err := d.db.Query(query, schema, table)
	if err != nil {
		return nil, fmt.Errorf("query columns for %s: %w", table, err)
	}
	defer rows.Close()

	var cols []models.ColumnMeta
	for rows.Next() {
		var name, dtype, nullable, comment string
		var isFK bool
		if err := rows.Scan(&name, &dtype, &nullable, &comment, &isFK); err != nil {
			return nil, fmt.Errorf("scan column: %w", err)
		}

		fkRef := ""
		if isFK {
			ref := d.resolveFK(schema, table, name)
			if ref != "" {
				fkRef = ref
			}
		}

		cols = append(cols, models.ColumnMeta{
			Name:        name,
			Type:        dtype,
			Nullable:    nullable == "YES",
			Description: comment,
			IsFK:        isFK,
			FKRef:       fkRef,
		})
	}
	return cols, nil
}

func (d *Dictionary) loadPrimaryKey(schema, table string) (string, error) {
	query := `
		SELECT kcu.column_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
		  ON kcu.constraint_name = tc.constraint_name
		WHERE tc.table_schema = $1
		  AND tc.table_name = $2
		  AND tc.constraint_type = 'PRIMARY KEY'
		ORDER BY kcu.ordinal_position
		LIMIT 1`

	var pk string
	err := d.db.QueryRow(query, schema, table).Scan(&pk)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return pk, err
}

func (d *Dictionary) resolveFK(schema, table, column string) string {
	query := `
		SELECT ccu.table_name || '.' || ccu.column_name
		FROM information_schema.key_column_usage kcu
		JOIN information_schema.constraint_column_usage ccu
		  ON ccu.constraint_name = kcu.constraint_name
		JOIN information_schema.table_constraints tc
		  ON tc.constraint_name = kcu.constraint_name
		WHERE tc.constraint_type = 'FOREIGN KEY'
		  AND tc.table_schema = $1
		  AND kcu.table_name = $2
		  AND kcu.column_name = $3
		LIMIT 1`

	var ref string
	err := d.db.QueryRow(query, schema, table, column).Scan(&ref)
	if err != nil {
		return ""
	}
	return ref
}

// LoadRelationships extracts foreign key relationships.
func (d *Dictionary) LoadRelationships(schema string) ([]models.Relationship, error) {
	query := `
		SELECT kcu.table_name, kcu.column_name,
		       ccu.table_name, ccu.column_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
		  ON kcu.constraint_name = tc.constraint_name
		JOIN information_schema.constraint_column_usage ccu
		  ON ccu.constraint_name = tc.constraint_name
		WHERE tc.constraint_type = 'FOREIGN KEY'
		  AND tc.table_schema = $1
		ORDER BY kcu.table_name`

	rows, err := d.db.Query(query, schema)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rels []models.Relationship
	for rows.Next() {
		var r models.Relationship
		if err := rows.Scan(&r.FromTable, &r.FromColumn, &r.ToTable, &r.ToColumn); err != nil {
			return nil, err
		}
		rels = append(rels, r)
	}
	return rels, nil
}

// FormatForPrompt renders the table metadata as a string for the LLM prompt.
func FormatForPrompt(tables []models.TableMeta, rels []models.Relationship, semantics []models.SemanticMapping) string {
	var b strings.Builder

	b.WriteString("=== 数据表结构 ===\n\n")
	for _, t := range tables {
		b.WriteString(fmt.Sprintf("表: %s", t.Name))
		if t.Description != "" {
			b.WriteString(fmt.Sprintf(" (%s)", t.Description))
		}
		b.WriteString("\n")
		if t.PrimaryKey != "" {
			b.WriteString(fmt.Sprintf("  主键: %s\n", t.PrimaryKey))
		}
		for _, c := range t.Columns {
			b.WriteString(fmt.Sprintf("  - %s (%s)", c.Name, c.Type))
			if c.IsFK && c.FKRef != "" {
				b.WriteString(fmt.Sprintf(" -> %s", c.FKRef))
			}
			if c.Description != "" {
				b.WriteString(fmt.Sprintf(" -- %s", c.Description))
			}
			if !c.Nullable {
				b.WriteString(" [NOT NULL]")
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	b.WriteString("=== 表关联关系 ===\n\n")
	for _, r := range rels {
		b.WriteString(fmt.Sprintf("%s.%s -> %s.%s\n", r.FromTable, r.FromColumn, r.ToTable, r.ToColumn))
	}

	if len(semantics) > 0 {
		b.WriteString("\n=== 业务语义映射 ===\n\n")
		for _, s := range semantics {
			b.WriteString(fmt.Sprintf("- \"%s\" → %s", s.BusinessTerm, s.SQLFragment))
			if s.Description != "" {
				b.WriteString(fmt.Sprintf(" (%s)", s.Description))
			}
			b.WriteString("\n")
		}
	}

	return b.String()
}
