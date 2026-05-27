-- 004_seed_default_permissions.sql
-- Seed data: default CRUD permissions for common ERP resources.
-- Run within a tenant schema after provisioning.

INSERT INTO permissions (name, key, resource, action) VALUES
    -- Model management
    ('创建模型', 'model:create',   'model', 'create'),
    ('查看模型', 'model:read',     'model', 'read'),
    ('更新模型', 'model:update',   'model', 'update'),
    ('删除模型', 'model:delete',   'model', 'delete'),

    -- Field management
    ('创建字段', 'field:create',   'field', 'create'),
    ('查看字段', 'field:read',     'field', 'read'),
    ('更新字段', 'field:update',   'field', 'update'),
    ('删除字段', 'field:delete',   'field', 'delete'),

    -- Menu management
    ('创建菜单', 'menu:create',    'menu', 'create'),
    ('查看菜单', 'menu:read',      'menu', 'read'),
    ('更新菜单', 'menu:update',    'menu', 'update'),
    ('删除菜单', 'menu:delete',    'menu', 'delete'),

    -- Permission management
    ('创建权限', 'permission:create', 'permission', 'create'),
    ('查看权限', 'permission:read',   'permission', 'read'),
    ('更新权限', 'permission:update', 'permission', 'update'),
    ('删除权限', 'permission:delete', 'permission', 'delete'),

    -- Role management
    ('创建角色', 'role:create',    'role', 'create'),
    ('查看角色', 'role:read',      'role', 'read'),
    ('更新角色', 'role:update',    'role', 'update'),
    ('删除角色', 'role:delete',    'role', 'delete')
ON CONFLICT (key) DO NOTHING;

-- Grant all permissions to the admin role
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.key = 'admin'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Seed default menus
INSERT INTO menus (name, label, icon, path, order_index) VALUES
    ('dashboard',     '首页',       'HomeOutlined',       '/dashboard',     0),
    ('model_manager', '模型管理',   'DatabaseOutlined',   '/models',       10),
    ('menu_manager',  '菜单管理',   'MenuOutlined',       '/menus',        20),
    ('perm_manager',  '权限管理',   'SafetyOutlined',     '/permissions',  30),
    ('role_manager',  '角色管理',   'TeamOutlined',       '/roles',        40),
    ('settings',      '系统设置',   'SettingOutlined',    '/settings',     90)
ON CONFLICT DO NOTHING;
