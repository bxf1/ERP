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
    ],
  },
];

export { routes };
