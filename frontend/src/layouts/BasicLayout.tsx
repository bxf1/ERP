import { useState } from 'react';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import {
  ProLayout,
  PageContainer,
} from '@ant-design/pro-components';
import {
  DashboardOutlined,
  UserOutlined,
  SettingOutlined,
  SafetyOutlined,
  KeyOutlined,
} from '@ant-design/icons';

const defaultMenus = [
  {
    path: '/',
    name: 'Dashboard',
    icon: <DashboardOutlined />,
  },
  {
    path: '/users',
    name: '用户管理',
    icon: <UserOutlined />,
  },
  {
    path: '/roles',
    name: '角色管理',
    icon: <SafetyOutlined />,
  },
  {
    path: '/permissions',
    name: '权限管理',
    icon: <KeyOutlined />,
  },
  {
    path: '/settings',
    name: '系统设置',
    icon: <SettingOutlined />,
  },
];

export default function BasicLayout() {
  const location = useLocation();
  const navigate = useNavigate();
  const [pathname, setPathname] = useState(location.pathname);

  return (
    <ProLayout
      title="ERP System"
      logo={null}
      location={{ pathname }}
      menuDataRender={() => defaultMenus}
      menuItemRender={(item, dom) => (
        <a
          onClick={() => {
            setPathname(item.path || '/');
            navigate(item.path || '/');
          }}
        >
          {dom}
        </a>
      )}
    >
      <PageContainer>
        <Outlet />
      </PageContainer>
    </ProLayout>
  );
}
