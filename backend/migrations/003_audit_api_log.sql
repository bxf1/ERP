-- Migration 003: Audit log and API log tables
-- Each tenant schema gets an audit_log and api_log for compliance and debugging.

BEGIN;

CREATE OR REPLACE FUNCTION install_tenant_v3_tables(_schema text)
RETURNS void AS $$
BEGIN

    -- =========================================================================
    -- AUDIT_LOG (操作审计日志)
    -- =========================================================================
    EXECUTE format($f$
        CREATE TABLE IF NOT EXISTS %I.audit_log (
            id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            user_id      UUID,
            action       VARCHAR(100) NOT NULL,
            resource     VARCHAR(200) NOT NULL,
            resource_id  UUID,
            old_values   JSONB,
            new_values   JSONB,
            ip_address   INET,
            user_agent   TEXT,
            created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
        );

        CREATE INDEX IF NOT EXISTS idx_audit_log_resource ON %I.audit_log (resource, resource_id);
        CREATE INDEX IF NOT EXISTS idx_audit_log_user    ON %I.audit_log (user_id);
        CREATE INDEX IF NOT EXISTS idx_audit_log_time    ON %I.audit_log (created_at DESC);
    $f$, _schema, _schema);

    -- =========================================================================
    -- API_LOG (API调用日志 — per-tenant)
    -- =========================================================================
    EXECUTE format($f$
        CREATE TABLE IF NOT EXISTS %I.api_log (
            id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            user_id      UUID,
            method       VARCHAR(10)  NOT NULL,
            path         VARCHAR(500) NOT NULL,
            status_code  INTEGER      NOT NULL,
            latency_ms   INTEGER      NOT NULL,
            request_body JSONB,
            response_body JSONB,
            ip_address   INET,
            user_agent   TEXT,
            trace_id     VARCHAR(100),
            created_at   TIMESTAMPTZ  NOT NULL DEFAULT now()
        );

        CREATE INDEX IF NOT EXISTS idx_api_log_path   ON %I.api_log (path);
        CREATE INDEX IF NOT EXISTS idx_api_log_status ON %I.api_log (status_code);
        CREATE INDEX IF NOT EXISTS idx_api_log_time   ON %I.api_log (created_at DESC);
    $f$, _schema, _schema);

    -- Record migration
    INSERT INTO public.schema_migrations (schema_name, version, name)
    VALUES (_schema, 3, 'install_tenant_v3_tables')
    ON CONFLICT (schema_name, version) DO NOTHING;

END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

COMMIT;
