-- Migration 005: Workflow engine tables
-- Each tenant schema gets workflow definition, node, edge, instance, approval, and history tables.

BEGIN;

CREATE OR REPLACE FUNCTION install_tenant_v5_tables(_schema text)
RETURNS void AS $$
BEGIN

    -- =========================================================================
    -- WORKFLOW_DEFINITIONS (工作流定义)
    -- =========================================================================
    EXECUTE format($f$
        CREATE TABLE IF NOT EXISTS %I.workflow_definitions (
            id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            name         VARCHAR(200) NOT NULL,
            description  TEXT,
            version      INTEGER NOT NULL DEFAULT 1,
            status       VARCHAR(20) NOT NULL DEFAULT 'draft',
            form_config  JSONB,
            created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
            updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
        );

        CREATE INDEX IF NOT EXISTS idx_wf_def_status ON %I.workflow_definitions (status);
    $f$, _schema, _schema);

    -- =========================================================================
    -- WORKFLOW_NODES (工作流节点)
    -- =========================================================================
    EXECUTE format($f$
        CREATE TABLE IF NOT EXISTS %I.workflow_nodes (
            id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            definition_id   UUID NOT NULL REFERENCES %I.workflow_definitions(id) ON DELETE CASCADE,
            name            VARCHAR(200) NOT NULL,
            node_type       VARCHAR(50) NOT NULL,
            approver_rule   JSONB,
            form_view_id    UUID,
            position_x      FLOAT DEFAULT 0,
            position_y      FLOAT DEFAULT 0,
            config          JSONB,
            created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
            updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
        );

        CREATE INDEX IF NOT EXISTS idx_wf_node_def ON %I.workflow_nodes (definition_id);
    $f$, _schema, _schema, _schema);

    -- =========================================================================
    -- WORKFLOW_EDGES (工作流边/转换)
    -- =========================================================================
    EXECUTE format($f$
        CREATE TABLE IF NOT EXISTS %I.workflow_edges (
            id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            definition_id   UUID NOT NULL REFERENCES %I.workflow_definitions(id) ON DELETE CASCADE,
            source_node_id  UUID NOT NULL REFERENCES %I.workflow_nodes(id) ON DELETE CASCADE,
            target_node_id  UUID NOT NULL REFERENCES %I.workflow_nodes(id) ON DELETE CASCADE,
            condition_expr  TEXT,
            label           VARCHAR(200),
            created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
        );

        CREATE INDEX IF NOT EXISTS idx_wf_edge_def ON %I.workflow_edges (definition_id);
    $f$, _schema, _schema, _schema, _schema);

    -- =========================================================================
    -- WORKFLOW_INSTANCES (流程实例)
    -- =========================================================================
    EXECUTE format($f$
        CREATE TABLE IF NOT EXISTS %I.workflow_instances (
            id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            definition_id   UUID NOT NULL REFERENCES %I.workflow_definitions(id),
            current_node_id UUID REFERENCES %I.workflow_nodes(id),
            status          VARCHAR(20) NOT NULL DEFAULT 'running',
            form_data       JSONB,
            submitted_by    UUID,
            submitted_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
            completed_at    TIMESTAMPTZ,
            created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
            updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
        );

        CREATE INDEX IF NOT EXISTS idx_wf_inst_def     ON %I.workflow_instances (definition_id);
        CREATE INDEX IF NOT EXISTS idx_wf_inst_status  ON %I.workflow_instances (status);
        CREATE INDEX IF NOT EXISTS idx_wf_inst_sub_by  ON %I.workflow_instances (submitted_by);
        CREATE INDEX IF NOT EXISTS idx_wf_inst_cur_node ON %I.workflow_instances (current_node_id);
    $f$, _schema, _schema, _schema, _schema);

    -- =========================================================================
    -- WORKFLOW_APPROVALS (审批记录)
    -- =========================================================================
    EXECUTE format($f$
        CREATE TABLE IF NOT EXISTS %I.workflow_approvals (
            id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            instance_id   UUID NOT NULL REFERENCES %I.workflow_instances(id) ON DELETE CASCADE,
            node_id       UUID NOT NULL REFERENCES %I.workflow_nodes(id),
            approver_id   UUID,
            action        VARCHAR(20) NOT NULL,
            comment       TEXT,
            created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
        );

        CREATE INDEX IF NOT EXISTS idx_wf_appr_inst ON %I.workflow_approvals (instance_id);
        CREATE INDEX IF NOT EXISTS idx_wf_appr_node ON %I.workflow_approvals (node_id);
    $f$, _schema, _schema, _schema, _schema);

    -- =========================================================================
    -- WORKFLOW_INSTANCE_HISTORY (流程历史)
    -- =========================================================================
    EXECUTE format($f$
        CREATE TABLE IF NOT EXISTS %I.workflow_instance_history (
            id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            instance_id   UUID NOT NULL REFERENCES %I.workflow_instances(id) ON DELETE CASCADE,
            node_id       UUID REFERENCES %I.workflow_nodes(id),
            action        VARCHAR(50) NOT NULL,
            operator_id   UUID,
            comment       TEXT,
            from_node_id  UUID,
            to_node_id    UUID,
            details       JSONB,
            created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
        );

        CREATE INDEX IF NOT EXISTS idx_wf_hist_inst ON %I.workflow_instance_history (instance_id);
        CREATE INDEX IF NOT EXISTS idx_wf_hist_time ON %I.workflow_instance_history (created_at DESC);
    $f$, _schema, _schema, _schema);

    -- Record migration
    INSERT INTO public.schema_migrations (schema_name, version, name)
    VALUES (_schema, 5, 'install_tenant_v5_tables')
    ON CONFLICT (schema_name, version) DO NOTHING;

END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

COMMIT;
