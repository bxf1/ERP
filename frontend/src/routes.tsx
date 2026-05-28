import type { RouteObject } from 'react-router-dom';
import BasicLayout from './layouts/BasicLayout';

const routes: RouteObject[] = [
  {
    path: '/',
    element: <BasicLayout />,
    children: [
      {
        index: true,
        lazy: () => import('./pages/NL2SQL'),
      },
      {
        path: 'users',
        lazy: () => import('./pages/User/List'),
      },
      {
        path: 'roles',
        lazy: () => import('./pages/permission/RoleList'),
      },
      {
        path: 'permissions',
        lazy: () => import('./pages/permission/PermissionList'),
      },
    ],
  },
];

export { routes };
