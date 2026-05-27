import React from 'react';
import { Card, Tree, Typography, Tag } from 'antd';
import {
  FolderOutlined,
  FileOutlined,
  AppstoreOutlined,
} from '@ant-design/icons';
import type { MenuPositionConfig } from '@/types';

const { Text } = Typography;

interface Props {
  menu: MenuPositionConfig;
}

const iconMap: Record<string, React.ReactNode> = {
  directory: <FolderOutlined style={{ color: '#faad14' }} />,
  page: <FileOutlined style={{ color: '#1677ff' }} />,
  button: <AppstoreOutlined style={{ color: '#52c41a' }} />,
};

const typeLabel: Record<string, string> = {
  directory: '目录',
  page: '页面',
  button: '按钮',
};

const MenuPreview: React.FC<Props> = ({ menu }) => {
  const treeData = [
    {
      title: 'ERP 菜单',
      key: 'root',
      icon: <FolderOutlined />,
      children: [
        {
          title: menu.parentMenu,
          key: 'parent',
          icon: <FolderOutlined style={{ color: '#faad14' }} />,
          children: [
            {
              title: (
                <span>
                  {menu.menuName}
                  <Tag color="processing" style={{ marginLeft: 8, fontSize: 10 }}>
                    {typeLabel[menu.menuType]}
                  </Tag>
                </span>
              ),
              key: 'new-menu',
              icon: iconMap[menu.menuType],
              style: { background: '#e6f4ff' },
            },
          ],
        },
      ],
    },
  ];

  return (
    <Card title="菜单位置预览" size="small" style={{ marginBottom: 16 }}>
      <Tree
        showIcon
        defaultExpandAll
        treeData={treeData}
        style={{ background: 'transparent' }}
      />
      <div style={{ marginTop: 12, padding: '8px 12px', background: '#fafafa', borderRadius: 8 }}>
        <Text style={{ fontSize: 13 }}>
          路径: <Text code>{menu.routePath}</Text>
        </Text>
        <br />
        <Text style={{ fontSize: 13 }}>
          图标: <Text code>{menu.icon}</Text>
        </Text>
        <br />
        <Text style={{ fontSize: 13 }}>
          排序: <Text code>{menu.order}</Text>
        </Text>
      </div>
    </Card>
  );
};

export default MenuPreview;
