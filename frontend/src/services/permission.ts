import axios from 'axios';
import type { Role, Permission, UserRole, UserPermissions, DataScope } from '../types/permission';

const api = axios.create({
  baseURL: '/api/v1',
  timeout: 30000,
});

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

api.interceptors.response.use(
  (response) => response.data,
  (error) => Promise.reject(error),
);

// —— Roles ———————————————————————————————————————

export async function getRoles(): Promise<Role[]> {
  const res: any = await api.get('/roles');
  return res.data ?? [];
}

export async function getRole(id: string): Promise<Role> {
  const res: any = await api.get(`/roles/${id}`);
  return res.data;
}

export async function createRole(data: Partial<Role>): Promise<Role> {
  const res: any = await api.post('/roles', data);
  return res.data;
}

export async function updateRole(id: string, data: Partial<Role>): Promise<void> {
  await api.put(`/roles/${id}`, data);
}

export async function deleteRole(id: string): Promise<void> {
  await api.delete(`/roles/${id}`);
}

export async function assignPermissions(roleId: string, permissionIds: string[]): Promise<void> {
  await api.post(`/roles/${roleId}/permissions`, { permission_ids: permissionIds });
}

export async function setDataScope(roleId: string, data: Partial<DataScope>): Promise<DataScope> {
  const res: any = await api.post(`/roles/${roleId}/data-scope`, data);
  return res.data;
}

// —— Permissions ——————————————————————————————————

export async function getPermissions(): Promise<Permission[]> {
  const res: any = await api.get('/permissions');
  return res.data ?? [];
}

export async function getPermissionsFlat(): Promise<Permission[]> {
  const res: any = await api.get('/permissions/flat');
  return res.data ?? [];
}

export async function createPermission(data: Partial<Permission>): Promise<Permission> {
  const res: any = await api.post('/permissions', data);
  return res.data;
}

export async function updatePermission(id: string, data: Partial<Permission>): Promise<void> {
  await api.put(`/permissions/${id}`, data);
}

export async function deletePermission(id: string): Promise<void> {
  await api.delete(`/permissions/${id}`);
}

// —— User roles ————————————————————————————————————

export async function getAvailableRoles(): Promise<Role[]> {
  const res: any = await api.get('/roles/available');
  return res.data ?? [];
}

export async function getUserRoles(userId: string): Promise<UserRole[]> {
  const res: any = await api.get(`/users/${userId}/roles`);
  return res.data ?? [];
}

export async function getUserPermissions(userId: string): Promise<UserPermissions> {
  const res: any = await api.get(`/users/${userId}/permissions`);
  return res.data;
}

export async function assignRoleToUser(userId: string, roleId: string, effectiveFrom?: string, effectiveTo?: string): Promise<void> {
  await api.post(`/users/${userId}/roles`, { role_id: roleId, effective_from: effectiveFrom, effective_to: effectiveTo });
}

export async function removeRoleFromUser(userId: string, roleId: string): Promise<void> {
  await api.delete(`/users/${userId}/roles/${roleId}`);
}
