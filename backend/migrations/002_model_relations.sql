-- Migration 002: Enhanced metadata model with model relationships
-- Adds relationship support between models (one-to-many, many-to-many via junction tables).

BEGIN;

CREATE OR REPLACE FUNCTION install_tenant_v2_tables(_schema text)
RETURNS void AS $$
BEGIN

    -- =========================================================================
    -- MODEL_RELATIONS (模型关系表 — for linking models)
    -- =========================================================================
    EXECUTE format($f$
        CREATE TABLE IF NOT EXISTS %I.model_relations (
            id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            source_model_id UUID         NOT NULL REFERENCES %I.models(id) ON DELETE CASCADE,
            target_model_id UUID         NOT NULL REFERENCES %I.models(id) ON DELETE CASCADE,
            relation_type   VARCHAR(20)  NOT NULL CHECK (relation_type IN ('one_to_one', 'one_to_many', 'many_to_many')),
            source_field    VARCHAR(200) NOT NULL,
            target_field    VARCHAR(200),
            junction_table  VARCHAR(200),
            label           VARCHAR(200),
            created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
            updated_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
            UNIQUE (source_model_id, target_model_id, source_field)
        );

        CREATE INDEX IF NOT EXISTS idx_model_relations_src ON %I.model_relations (source_model_id);
        CREATE INDEX IF NOT EXISTS idx_model_relations_tgt ON %I.model_relations (target_model_id);
    $f$, _schema, _schema, _schema, _schema);

    -- Record migration
    INSERT INTO public.schema_migrations (schema_name, version, name)
    VALUES (_schema, 2, 'install_tenant_v2_tables')
    ON CONFLICT (schema_name, version) DO NOTHING;

END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

COMMIT;
