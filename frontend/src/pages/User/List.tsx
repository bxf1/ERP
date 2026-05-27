import { ProTable } from '@ant-design/pro-components';
import type { ProColumns } from '@ant-design/pro-components';
import { Button } from 'antd';
import { PlusOutlined } from '@ant-design/icons';

interface UserItem {
  id: number;
  username: string;
  email: string;
  phone: string;
  status: number;
}

const columns: ProColumns<UserItem>[] = [
  { title: 'ID', dataIndex: 'id', key: 'id' },
  { title: '用户名', dataIndex: 'username', key: 'username' },
  { title: '邮箱', dataIndex: 'email', key: 'email' },
  { title: '手机', dataIndex: 'phone', key: 'phone' },
  {
    title: '状态',
    dataIndex: 'status',
    key: 'status',
    valueEnum: {
      0: { text: '禁用', status: 'Error' },
      1: { text: '启用', status: 'Success' },
    },
  },
];

export function Component() {
  return (
    <ProTable<UserItem>
      columns={columns}
      request={async () => {
        return { data: [], success: true, total: 0 };
      }}
      rowKey="id"
      search={{ labelWidth: 'auto' }}
      toolBarRender={() => [
        <Button key="add" type="primary" icon={<PlusOutlined />}>
          新增用户
        </Button>,
      ]}
    />
  );
}
