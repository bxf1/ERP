-- Migration 001: Create public schema & tenants table
-- This is the bootstrap migration for the multi-tenant ERP system.
-- The "public" schema holds system-wide tables shared across all tenants.
-- Each tenant gets its own PostgreSQL schema (schema-per-tenant isolation).

BEGIN;

-- ============================================================================
-- TENANTS REGISTRY (shared — lives in public schema)
-- ============================================================================
CREATE TABLE IF NOT EXISTS tenants (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(200) NOT NULL,
    slug        VARCHAR(100) NOT NULL UNIQUE,
    schema_name VARCHAR(63)  NOT NULL UNIQUE,
    status      VARCHAR(20)  NOT NULL DEFAULT 'active'
                CHECK (status IN ('active', 'suspended', 'deleted')),
    config      JSONB        NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
);

CREATE INDEX idx_tenants_slug   ON tenants (slug);
CREATE INDEX idx_tenants_status ON tenants (status);

-- ============================================================================
-- Migration log (tracks applied migrations per schema)
-- ============================================================================
CREATE TABLE IF NOT EXISTS schema_migrations (
    id          SERIAL PRIMARY KEY,
    schema_name VARCHAR(63)  NOT NULL,
    version     INTEGER      NOT NULL,
    name        VARCHAR(300) NOT NULL,
    applied_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
    UNIQUE (schema_name, version)
);

-- ============================================================================
-- Helper: create a new tenant schema and apply all tenant-level migrations.
-- Called when a new tenant is registered.
-- ============================================================================
CREATE OR REPLACE FUNCTION create_tenant_schema(_schema_name text, _slug text, _name text)
RETURNS UUID AS $$
DECLARE
    _tenant_id UUID;
BEGIN
    -- 1. Create the tenant schema
    EXECUTE format('CREATE SCHEMA IF NOT EXISTS %I', _schema_name);

    -- 2. Insert into public.tenants
    INSERT INTO tenants (name, slug, schema_name)
    VALUES (_name, _slug, _schema_name)
    RETURNING id INTO _tenant_id;

    RETURN _tenant_id;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- ============================================================================
-- Helper: install tenant-level tables inside a given schema.
-- This is also run by the migration runner for manual schema creation.
-- ============================================================================
CREATE OR REPLACE FUNCTION install_tenant_tables(_schema text)
RETURNS void AS $$
BEGIN
    -- =========================================================================
    -- MODELS (元数据 — 模型定义表)
    -- =========================================================================
    EXECUTE format($f$
        CREATE TABLE IF NOT EXISTS %I.models (
            id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            name          VARCHAR(200) NOT NULL,
            table_name    VARCHAR(200) NOT NULL UNIQUE,
            label         VARCHAR(200) NOT NULL,
            description   TEXT,
            is_system     BOOLEAN      NOT NULL DEFAULT false,
            config        JSONB        NOT NULL DEFAULT '{}',
            created_at    TIMESTAMPTZ  NOT NULL DEFAULT now(),
            updated_at    TIMESTAMPTZ  NOT NULL DEFAULT now()
        );

        CREATE INDEX IF NOT EXISTS idx_models_name ON %I.models (name);
    $f$, _schema, _schema);

    -- =========================================================================
    -- FIELDS (元数据 — 字段定义表)
    -- =========================================================================
    EXECUTE format($f$
        CREATE TABLE IF NOT EXISTS %I.fields (
            id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            model_id         UUID         NOT NULL REFERENCES %I.models(id) ON DELETE CASCADE,
            name             VARCHAR(200) NOT NULL,
            column_name      VARCHAR(200) NOT NULL,
            label            VARCHAR(200) NOT NULL,
            field_type       VARCHAR(50)  NOT NULL DEFAULT 'string'
                             CHECK (field_type IN (
                                 'string', 'text', 'integer', 'decimal',
                                 'boolean', 'date', 'datetime', 'json',
                                 'relation', 'file'
                             )),
            is_required      BOOLEAN      NOT NULL DEFAULT false,
            is_unique        BOOLEAN      NOT NULL DEFAULT false,
            default_value    TEXT,
            validation_rules JSONB        NOT NULL DEFAULT '{}',
            ui_config        JSONB        NOT NULL DEFAULT '{}',
            order_index      INTEGER      NOT NULL DEFAULT 0,
            created_at       TIMESTAMPTZ  NOT NULL DEFAULT now(),
            updated_at       TIMESTAMPTZ  NOT NULL DEFAULT now(),
            UNIQUE (model_id, name)
        );

        CREATE INDEX IF NOT EXISTS idx_fields_model_id ON %I.fields (model_id);
    $f$, _schema, _schema, _schema);

    -- =========================================================================
    -- MENUS (菜单表)
    -- =========================================================================
    EXECUTE format($f$
        CREATE TABLE IF NOT EXISTS %I.menus (
            id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            parent_id      UUID         REFERENCES %I.menus(id) ON DELETE SET NULL,
            name           VARCHAR(200) NOT NULL,
            label          VARCHAR(200) NOT NULL,
            icon           VARCHAR(100),
            path           VARCHAR(500) NOT NULL,
            component      VARCHAR(500),
            permission_key VARCHAR(200),
            order_index    INTEGER      NOT NULL DEFAULT 0,
            is_visible     BOOLEAN      NOT NULL DEFAULT true,
            created_at     TIMESTAMPTZ  NOT NULL DEFAULT now(),
            updated_at     TIMESTAMPTZ  NOT NULL DEFAULT now()
        );

        CREATE INDEX IF NOT EXISTS idx_menus_parent ON %I.menus (parent_id);
    $f$, _schema, _schema, _schema);

    -- =========================================================================
    -- ROLES (角色表)
    -- =========================================================================
    EXECUTE format($f$
        CREATE TABLE IF NOT EXISTS %I.roles (
            id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            name        VARCHAR(200) NOT NULL,
            code         VARCHAR(200) NOT NULL UNIQUE,
            description TEXT,
            is_system   BOOLEAN      NOT NULL DEFAULT false,
            created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
            updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
        );
    $f$, _schema);

    -- =========================================================================
    -- PERMISSIONS (权限表)
    -- =========================================================================
    EXECUTE format($f$
        CREATE TABLE IF NOT EXISTS %I.permissions (
            id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
            name        VARCHAR(200) NOT NULL,
            code         VARCHAR(200) NOT NULL UNIQUE,
            description TEXT,
            resource    VARCHAR(200) NOT NULL,
            action      VARCHAR(100) NOT NULL,
            created_at  TIMESTAMPTZ  NOT NULL DEFAULT now(),
            updated_at  TIMESTAMPTZ  NOT NULL DEFAULT now()
        );
    $f$, _schema);

    -- =========================================================================
    -- ROLE_PERMISSIONS (角色-权限关联表)
    -- =========================================================================
    EXECUTE format($f$
        CREATE TABLE IF NOT EXISTS %I.role_permissions (
            role_id       UUID NOT NULL REFERENCES %I.roles(id) ON DELETE CASCADE,
            permission_id UUID NOT NULL REFERENCES %I.permissions(id) ON DELETE CASCADE,
            PRIMARY KEY (role_id, permission_id)
        );
    $f$, _schema, _schema, _schema);

    -- =========================================================================
    -- USER_ROLES (用户-角色关联表)
    -- Users are identified by an external user ID (from auth system).
    -- =========================================================================
    EXECUTE format($f$
        CREATE TABLE IF NOT EXISTS %I.user_roles (
            user_id    UUID        NOT NULL,
            role_id    UUID        NOT NULL REFERENCES %I.roles(id) ON DELETE CASCADE,
            granted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
            PRIMARY KEY (user_id, role_id)
        );

        CREATE INDEX IF NOT EXISTS idx_user_roles_user ON %I.user_roles (user_id);
    $f$, _schema, _schema, _schema);

    -- =========================================================================
    -- SEED: default admin role & permission for new tenant
    -- =========================================================================
    EXECUTE format($f$
        INSERT INTO %I.roles (name, code, description, is_system) VALUES
            ('管理员', 'admin', 'Tenant administrator with full access', true),
            ('普通用户', 'user', 'Regular user', true)
        ON CONFLICT (code) DO NOTHING;

        INSERT INTO %I.permissions (name, code, resource, action) VALUES
            ('全部管理权限', 'admin.full_access', '*', '*')
        ON CONFLICT (code) DO NOTHING;

        INSERT INTO %I.role_permissions (role_id, permission_id)
        SELECT r.id, p.id
        FROM %I.roles r, %I.permissions p
        WHERE r.code = 'admin' AND p.code = 'admin.full_access'
        ON CONFLICT DO NOTHING;
    $f$, _schema, _schema, _schema, _schema, _schema);

    -- Record migration
    INSERT INTO public.schema_migrations (schema_name, version, name)
    VALUES (_schema, 1, 'install_tenant_tables')
    ON CONFLICT (schema_name, version) DO NOTHING;

END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

COMMIT;
