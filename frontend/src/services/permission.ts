import request from '../utils/request';
import type { Role, Permission, UserRole, UserPermissions, DataScope } from '../types/permission';

// —— Roles ———————————————————————————————————————

export async function getRoles(): Promise<Role[]> {
  const res = await request.get('/api/v1/roles');
  return res.data ?? [];
}

export async function getRole(id: string): Promise<Role> {
  const res = await request.get(`/api/v1/roles/${id}`);
  return res.data;
}

export async function createRole(data: Partial<Role>): Promise<Role> {
  const res = await request.post('/api/v1/roles', data);
  return res.data;
}

export async function updateRole(id: string, data: Partial<Role>): Promise<void> {
  await request.put(`/api/v1/roles/${id}`, data);
}

export async function deleteRole(id: string): Promise<void> {
  await request.delete(`/api/v1/roles/${id}`);
}

export async function assignPermissions(roleId: string, permissionIds: string[]): Promise<void> {
  await request.post(`/api/v1/roles/${roleId}/permissions`, { permission_ids: permissionIds });
}

export async function setDataScope(roleId: string, data: Partial<DataScope>): Promise<DataScope> {
  const res = await request.post(`/api/v1/roles/${roleId}/data-scope`, data);
  return res.data;
}

// —— Permissions ——————————————————————————————————

export async function getPermissions(): Promise<Permission[]> {
  const res = await request.get('/api/v1/permissions');
  return res.data ?? [];
}

export async function getPermissionsFlat(): Promise<Permission[]> {
  const res = await request.get('/api/v1/permissions/flat');
  return res.data ?? [];
}

export async function createPermission(data: Partial<Permission>): Promise<Permission> {
  const res = await request.post('/api/v1/permissions', data);
  return res.data;
}

export async function updatePermission(id: string, data: Partial<Permission>): Promise<void> {
  await request.put(`/api/v1/permissions/${id}`, data);
}

export async function deletePermission(id: string): Promise<void> {
  await request.delete(`/api/v1/permissions/${id}`);
}

// —— User roles ————————————————————————————————————

export async function getAvailableRoles(): Promise<Role[]> {
  const res = await request.get('/api/v1/roles/available');
  return res.data ?? [];
}

export async function getUserRoles(userId: string): Promise<UserRole[]> {
  const res = await request.get(`/api/v1/users/${userId}/roles`);
  return res.data ?? [];
}

export async function getUserPermissions(userId: string): Promise<UserPermissions> {
  const res = await request.get(`/api/v1/users/${userId}/permissions`);
  return res.data;
}

export async function assignRoleToUser(userId: string, roleId: string, effectiveFrom?: string, effectiveTo?: string): Promise<void> {
  await request.post(`/api/v1/users/${userId}/roles`, { role_id: roleId, effective_from: effectiveFrom, effective_to: effectiveTo });
}

export async function removeRoleFromUser(userId: string, roleId: string): Promise<void> {
  await request.delete(`/api/v1/users/${userId}/roles/${roleId}`);
}
