-- RBAC Permission Engine Schema
-- Migration 001: Core RBAC tables

CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    code VARCHAR(50) NOT NULL UNIQUE,
    description TEXT DEFAULT '',
    status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    is_system BOOLEAN DEFAULT false,
    sort_order INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    code VARCHAR(100) NOT NULL UNIQUE,
    resource VARCHAR(50) NOT NULL,
    action VARCHAR(50) NOT NULL,
    description TEXT DEFAULT '',
    parent_id UUID REFERENCES permissions(id) ON DELETE SET NULL,
    sort_order INT DEFAULT 0,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE IF NOT EXISTS user_roles (
    user_id UUID NOT NULL,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    effective_from TIMESTAMPTZ DEFAULT NOW(),
    effective_to TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    PRIMARY KEY (user_id, role_id)
);

CREATE TABLE IF NOT EXISTS data_scopes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    target_model VARCHAR(100) NOT NULL,
    scope_type VARCHAR(20) NOT NULL CHECK (scope_type IN ('all', 'department', 'self', 'custom')),
    scope_rule JSONB,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    UNIQUE (role_id, target_model)
);

CREATE TABLE IF NOT EXISTS permission_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    action VARCHAR(100) NOT NULL,
    resource VARCHAR(100) NOT NULL,
    target_id VARCHAR(100),
    result VARCHAR(20) NOT NULL CHECK (result IN ('allow', 'deny')),
    reason TEXT DEFAULT '',
    request_path VARCHAR(500),
    ip_address VARCHAR(50),
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_roles_code ON roles(code);
CREATE INDEX idx_roles_status ON roles(status);
CREATE INDEX idx_permissions_code ON permissions(code);
CREATE INDEX idx_permissions_resource ON permissions(resource);
CREATE INDEX idx_permissions_parent ON permissions(parent_id);
CREATE INDEX idx_role_permissions_role ON role_permissions(role_id);
CREATE INDEX idx_role_permissions_perm ON role_permissions(permission_id);
CREATE INDEX idx_user_roles_user ON user_roles(user_id);
CREATE INDEX idx_user_roles_role ON user_roles(role_id);
CREATE INDEX idx_data_scopes_role ON data_scopes(role_id);
CREATE INDEX idx_data_scopes_model ON data_scopes(target_model);
CREATE INDEX idx_audit_log_user ON permission_audit_log(user_id);
CREATE INDEX idx_audit_log_created ON permission_audit_log(created_at);

-- Seed default data
INSERT INTO roles (id, name, code, description, is_system, sort_order) VALUES
    ('10000000-0000-0000-0000-000000000001', '超级管理员', 'super_admin', '系统超级管理员，拥有所有权限', true, 1),
    ('10000000-0000-0000-0000-000000000002', '管理员', 'admin', '系统管理员', true, 2),
    ('10000000-0000-0000-0000-000000000003', '普通用户', 'user', '普通用户', true, 3)
ON CONFLICT DO NOTHING;

-- Seed default permissions (resource:action pattern)
INSERT INTO permissions (id, name, code, resource, action, description, sort_order) VALUES
    ('20000000-0000-0000-0000-000000000001', '用户查看', 'user:read', 'user', 'read', '查看用户列表和详情', 1),
    ('20000000-0000-0000-0000-000000000002', '用户创建', 'user:create', 'user', 'create', '创建新用户', 2),
    ('20000000-0000-0000-0000-000000000003', '用户编辑', 'user:update', 'user', 'update', '编辑用户信息', 3),
    ('20000000-0000-0000-0000-000000000004', '用户删除', 'user:delete', 'user', 'delete', '删除用户', 4),
    ('20000000-0000-0000-0000-000000000005', '角色管理', 'role:manage', 'role', 'manage', '管理角色和权限', 5),
    ('20000000-0000-0000-0000-000000000006', '系统配置', 'system:config', 'system', 'config', '系统配置管理', 6),
    ('20000000-0000-0000-0000-000000000007', '审计日志', 'audit:read', 'audit', 'read', '查看审计日志', 7)
ON CONFLICT DO NOTHING;

-- Super admin gets all permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT '10000000-0000-0000-0000-000000000001', id FROM permissions
ON CONFLICT DO NOTHING;

-- Default data scope: super_admin sees all data
INSERT INTO data_scopes (role_id, target_model, scope_type, scope_rule) VALUES
    ('10000000-0000-0000-0000-000000000001', '*', 'all', '{}'::jsonb),
    ('10000000-0000-0000-0000-000000000003', '*', 'self', '{"field": "created_by"}'::jsonb)
ON CONFLICT DO NOTHING;
