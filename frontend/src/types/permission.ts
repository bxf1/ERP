export interface Permission {
  id: string;
  name: string;
  code: string;
  resource: string;
  action: string;
  description: string;
  parent_id: string | null;
  sort_order: number;
  created_at: string;
  updated_at: string;
  children?: Permission[];
}

export interface Role {
  id: string;
  name: string;
  code: string;
  description: string;
  status: 'active' | 'inactive';
  is_system: boolean;
  sort_order: number;
  created_at: string;
  updated_at: string;
  permissions?: Permission[];
  data_scopes?: DataScope[];
}

export interface UserRole {
  user_id: string;
  role_id: string;
  effective_from: string;
  effective_to: string | null;
  created_at: string;
  role?: Role;
}

export interface DataScope {
  id: string;
  role_id: string;
  target_model: string;
  scope_type: string;
  scope_rule: string;
  created_at: string;
  updated_at: string;
}

export interface UserPermissions {
  user_id: string;
  permissions: string[];
  roles: string[];
}
