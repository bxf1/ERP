package datadict

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// Extractor reads schema metadata from PostgreSQL information_schema.
type Extractor struct {
	db *sql.DB
}

// NewExtractor creates an Extractor with the given database connection.
func NewExtractor(db *sql.DB) *Extractor {
	return &Extractor{db: db}
}

const (
	tablesQuery = `
		SELECT table_schema, table_name, obj_description(pc.oid) AS comment, table_type
		FROM information_schema.tables t
		JOIN pg_class pc ON pc.relname = t.table_name
		JOIN pg_namespace pn ON pn.oid = pc.relnamespace AND pn.nspname = t.table_schema
		WHERE table_schema NOT IN ('information_schema', 'pg_catalog')
		ORDER BY table_schema, table_name`

	columnsQuery = `
		SELECT
			c.table_schema, c.table_name,
			c.column_name, c.data_type, c.is_nullable,
			COALESCE(c.column_default, ''),
			COALESCE(pd.description, ''),
			c.character_maximum_length,
			c.numeric_scale
		FROM information_schema.columns c
		LEFT JOIN pg_catalog.pg_statio_all_tables st
			ON c.table_schema = st.schemaname AND c.table_name = st.relname
		LEFT JOIN pg_catalog.pg_description pd
			ON pd.objoid = st.relid AND pd.objsubid = c.ordinal_position
		WHERE c.table_schema NOT IN ('information_schema', 'pg_catalog')
		ORDER BY c.table_schema, c.table_name, c.ordinal_position`

	primaryKeysQuery = `
		SELECT kcu.table_schema, kcu.table_name, kcu.column_name
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
			ON tc.constraint_schema = kcu.constraint_schema
			AND tc.constraint_name = kcu.constraint_name
		WHERE tc.constraint_type = 'PRIMARY KEY'
		  AND tc.table_schema NOT IN ('information_schema', 'pg_catalog')
		ORDER BY kcu.table_schema, kcu.table_name, kcu.ordinal_position`

	indexesQuery = `
		SELECT
			i.relname AS index_name,
			a.attname AS column_name,
			t.relname AS table_name,
			ix.indisunique AS is_unique,
			n.nspname AS table_schema
		FROM pg_index ix
		JOIN pg_class i ON i.oid = ix.indexrelid
		JOIN pg_class t ON t.oid = ix.indrelid
		JOIN pg_namespace n ON n.oid = t.relnamespace
		JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = ANY(ix.indkey)
		WHERE n.nspname NOT IN ('information_schema', 'pg_catalog')
		ORDER BY n.nspname, t.relname, i.relname, a.attnum`

	relationsQuery = `
		SELECT
			tc.constraint_name,
			kcu.table_schema  AS source_schema,
			kcu.table_name     AS source_table,
			kcu.column_name    AS source_column,
			ccu.table_schema   AS target_schema,
			ccu.table_name     AS target_table,
			ccu.column_name    AS target_column
		FROM information_schema.table_constraints tc
		JOIN information_schema.key_column_usage kcu
			ON tc.constraint_schema = kcu.constraint_schema
			AND tc.constraint_name = kcu.constraint_name
		JOIN information_schema.constraint_column_usage ccu
			ON tc.constraint_schema = ccu.constraint_schema
			AND tc.constraint_name = ccu.constraint_name
		WHERE tc.constraint_type = 'FOREIGN KEY'
		ORDER BY tc.constraint_name`
)

// columnRow is an intermediate struct for scanning column query results.
type columnRow struct {
	schema, table, name, dataType, nullable, defaultVal, comment string
	maxLength, numericScale                                      sql.NullInt64
}

// Extract performs a full schema extraction and returns the DataDict.
func (e *Extractor) Extract(ctx context.Context) (*DataDict, error) {
	dict := &DataDict{ExtractedAt: time.Now().UTC()}

	tables, err := e.extractTables(ctx)
	if err != nil {
		return nil, fmt.Errorf("extract tables: %w", err)
	}

	cols, err := e.extractColumns(ctx)
	if err != nil {
		return nil, fmt.Errorf("extract columns: %w", err)
	}

	pks, err := e.extractPrimaryKeys(ctx)
	if err != nil {
		return nil, fmt.Errorf("extract primary keys: %w", err)
	}

	idxs, err := e.extractIndexes(ctx)
	if err != nil {
		return nil, fmt.Errorf("extract indexes: %w", err)
	}

	relations, err := e.extractRelations(ctx)
	if err != nil {
		return nil, fmt.Errorf("extract relations: %w", err)
	}

	// Assemble columns into tables and attach primary keys.
	pkSet := buildPKSet(pks)
	idxMap := buildIndexMap(idxs)

	for i := range tables {
		key := tableKey(tables[i].Schema, tables[i].Name)
		tables[i].Columns = cols[key]
		tables[i].PrimaryKeys = pkSet[key]
		tables[i].Indexes = idxMap[key]
	}

	dict.Tables = tables
	dict.Relations = relations
	return dict, nil
}

func (e *Extractor) extractTables(ctx context.Context) ([]TableInfo, error) {
	rows, err := e.db.QueryContext(ctx, tablesQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []TableInfo
	for rows.Next() {
		var t TableInfo
		var comment sql.NullString
		if err := rows.Scan(&t.Schema, &t.Name, &comment, &t.Type); err != nil {
			return nil, err
		}
		t.Comment = comment.String
		tables = append(tables, t)
	}
	return tables, rows.Err()
}

func (e *Extractor) extractColumns(ctx context.Context) (map[string][]ColumnInfo, error) {
	rows, err := e.db.QueryContext(ctx, columnsQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string][]ColumnInfo)
	for rows.Next() {
		var r columnRow
		if err := rows.Scan(&r.schema, &r.table, &r.name, &r.dataType,
			&r.nullable, &r.defaultVal, &r.comment,
			&r.maxLength, &r.numericScale); err != nil {
			return nil, err
		}
		col := ColumnInfo{
			Name:         r.name,
			DataType:     r.dataType,
			Nullable:     r.nullable == "YES",
			DefaultValue: r.defaultVal,
			Comment:      r.comment,
		}
		if r.maxLength.Valid {
			v := int(r.maxLength.Int64)
			col.MaxLength = &v
		}
		if r.numericScale.Valid {
			v := int(r.numericScale.Int64)
			col.NumericScale = &v
		}
		key := tableKey(r.schema, r.table)
		result[key] = append(result[key], col)
	}
	return result, rows.Err()
}

func (e *Extractor) extractPrimaryKeys(ctx context.Context) ([]struct{ schema, table, column string }, error) {
	rows, err := e.db.QueryContext(ctx, primaryKeysQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var pks []struct{ schema, table, column string }
	for rows.Next() {
		var pk struct{ schema, table, column string }
		if err := rows.Scan(&pk.schema, &pk.table, &pk.column); err != nil {
			return nil, err
		}
		pks = append(pks, pk)
	}
	return pks, rows.Err()
}

func (e *Extractor) extractIndexes(ctx context.Context) ([]struct{ schema, table, index, column string; unique bool }, error) {
	rows, err := e.db.QueryContext(ctx, indexesQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var idxs []struct{ schema, table, index, column string; unique bool }
	for rows.Next() {
		var idx struct{ schema, table, index, column string; unique bool }
		if err := rows.Scan(&idx.index, &idx.column, &idx.table, &idx.unique, &idx.schema); err != nil {
			return nil, err
		}
		idxs = append(idxs, idx)
	}
	return idxs, rows.Err()
}

func (e *Extractor) extractRelations(ctx context.Context) ([]RelationInfo, error) {
	rows, err := e.db.QueryContext(ctx, relationsQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rels []RelationInfo
	for rows.Next() {
		var r RelationInfo
		if err := rows.Scan(&r.Name, &r.SourceSchema, &r.SourceTable, &r.SourceColumn,
			&r.TargetSchema, &r.TargetTable, &r.TargetColumn); err != nil {
			return nil, err
		}
		rels = append(rels, r)
	}
	return rels, rows.Err()
}

func tableKey(schema, table string) string {
	return schema + "." + table
}

func buildPKSet(pks []struct{ schema, table, column string }) map[string][]string {
	out := make(map[string][]string)
	for _, pk := range pks {
		key := tableKey(pk.schema, pk.table)
		out[key] = append(out[key], pk.column)
	}
	return out
}

func buildIndexMap(idxs []struct{ schema, table, index, column string; unique bool }) map[string][]IndexInfo {
	// Group columns by index first, then group indexes by table.
	type idxGroup struct {
		columns []string
		unique  bool
	}
	idxGroups := make(map[string]*idxGroup)
	for _, ix := range idxs {
		g, ok := idxGroups[ix.index]
		if !ok {
			g = &idxGroup{unique: ix.unique}
			idxGroups[ix.index] = g
		}
		g.columns = append(g.columns, ix.column)
	}

	out := make(map[string][]IndexInfo)
	for _, ix := range idxs {
		g := idxGroups[ix.index]
		key := tableKey(ix.schema, ix.table)
		// Deduplicate by index name.
		found := false
		for _, existing := range out[key] {
			if existing.Name == ix.index {
				found = true
				break
			}
		}
		if !found {
			out[key] = append(out[key], IndexInfo{
				Name:    ix.index,
				Columns: g.columns,
				Unique:  g.unique,
			})
		}
	}
	return out
}
