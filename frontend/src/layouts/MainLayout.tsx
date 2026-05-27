import { useState } from 'react';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import { ProLayout, PageContainer } from '@ant-design/pro-components';
import {
  DashboardOutlined, ShopOutlined, TeamOutlined, ShoppingCartOutlined,
  DollarOutlined, DatabaseOutlined, CheckSquareOutlined, UnorderedListOutlined,
} from '@ant-design/icons';

const defaultMenus = {
  path: '/',
  children: [
    { path: '/dashboard', name: '进销存看板', icon: <DashboardOutlined /> },
    {
      name: '基础资料', icon: <DatabaseOutlined />,
      children: [
        { path: '/supplier', name: '供应商管理', icon: <TeamOutlined /> },
        { path: '/customer', name: '客户管理', icon: <TeamOutlined /> },
      ],
    },
    {
      name: '采购管理', icon: <ShoppingCartOutlined />,
      children: [
        { path: '/purchase', name: '采购订单', icon: <UnorderedListOutlined /> },
      ],
    },
    {
      name: '销售管理', icon: <DollarOutlined />,
      children: [
        { path: '/sales', name: '销售订单', icon: <UnorderedListOutlined /> },
      ],
    },
    {
      name: '库存管理', icon: <ShopOutlined />,
      children: [
        { path: '/inventory', name: '库存台账', icon: <DatabaseOutlined /> },
        { path: '/stocktaking', name: '库存盘点', icon: <CheckSquareOutlined /> },
      ],
    },
  ],
};

export default function MainLayout() {
  const navigate = useNavigate();
  const location = useLocation();
  const [collapsed, setCollapsed] = useState(false);

  return (
    <ProLayout
      title="ERP 进销存"
      logo="📦"
      route={defaultMenus}
      location={{ pathname: location.pathname }}
      collapsed={collapsed}
      onCollapse={setCollapsed}
      menuItemRender={(item, dom) => (
        <a onClick={() => item.path && navigate(item.path)}>{dom}</a>
      )}
      headerTitleRender={(_, __, ___) => <span style={{ fontWeight: 600, fontSize: 16 }}>ERP 进销存管理系统</span>}
    >
      <PageContainer>
        <Outlet />
      </PageContainer>
    </ProLayout>
  );
}
