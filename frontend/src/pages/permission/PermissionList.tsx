import { useRef, useState } from 'react';
import { ProTable } from '@ant-design/pro-components';
import type { ProColumns, ActionType } from '@ant-design/pro-components';
import { Button, Tag, Modal, Form, Input, Select, InputNumber, TreeSelect, Popconfirm, message } from 'antd';
import { PlusOutlined, ReloadOutlined } from '@ant-design/icons';
import type { Permission } from '../../types/permission';
import {
  getPermissions,
  getPermissionsFlat,
  createPermission,
  updatePermission,
  deletePermission,
} from '../../services/permission';

function buildTreeSelectData(perms: Permission[]): any[] {
  return perms.map((p) => ({
    title: `${p.name} (${p.code})`,
    value: p.id,
    children: p.children ? buildTreeSelectData(p.children) : [],
  }));
}

export default function PermissionList() {
  const actionRef = useRef<ActionType>();
  const [modalOpen, setModalOpen] = useState(false);
  const [editingRecord, setEditingRecord] = useState<Permission | null>(null);
  const [treeData, setTreeData] = useState<any[]>([]);
  const [tabKey, setTabKey] = useState<string>('flat');
  const [form] = Form.useForm();

  const loadTreeData = async () => {
    const tree = await getPermissions();
    setTreeData(buildTreeSelectData(tree));
  };

  const columns: ProColumns<Permission>[] = [
    { title: '权限名称', dataIndex: 'name', width: 150 },
    { title: '权限编码', dataIndex: 'code', width: 160, copyable: true },
    { title: '资源', dataIndex: 'resource', width: 120 },
    { title: '操作', dataIndex: 'action', width: 100,
      render: (_, record) => {
        const colors: Record<string, string> = { read: 'blue', create: 'green', update: 'orange', delete: 'red', manage: 'purple' };
        return <Tag color={colors[record.action] || 'default'}>{record.action}</Tag>;
      },
    },
    { title: '描述', dataIndex: 'description', ellipsis: true },
    { title: '排序', dataIndex: 'sort_order', width: 80 },
    {
      title: '操作', valueType: 'option', width: 140,
      render: (_, record) => [
        <a key="edit" onClick={async () => {
          await loadTreeData();
          setEditingRecord(record);
          form.setFieldsValue({ ...record, parent_id: record.parent_id || undefined });
          setModalOpen(true);
        }}>编辑</a>,
        <Popconfirm key="del" title="确认删除此权限？" onConfirm={async () => {
          await deletePermission(record.id);
          message.success('已删除');
          actionRef.current?.reload();
        }}>
          <a>删除</a>
        </Popconfirm>,
      ],
    },
  ];

  const treeColumns: ProColumns<Permission>[] = [
    { title: '权限名称', dataIndex: 'name', width: 200 },
    { title: '权限编码', dataIndex: 'code', width: 180, copyable: true },
    { title: '资源', dataIndex: 'resource', width: 100 },
    { title: '操作', dataIndex: 'action', width: 90,
      render: (_, record) => <Tag>{record.action}</Tag>,
    },
    { title: '描述', dataIndex: 'description', ellipsis: true },
  ];

  return (
    <>
      <ProTable<Permission>
        headerTitle="权限管理"
        actionRef={actionRef}
        columns={tabKey === 'flat' ? columns : treeColumns}
        request={async () => {
          const data = tabKey === 'flat'
            ? await getPermissionsFlat()
            : (await getPermissions()).flatMap(flattenTree);
          return { data, total: data.length, success: true };
        }}
        rowKey="id"
        search={false}
        toolbar={{
          tabs: {
            activeKey: tabKey,
            items: [
              { key: 'flat', label: '列表视图' },
              { key: 'tree', label: '树形视图' },
            ],
            onChange: (key) => { setTabKey(key); actionRef.current?.reload(); },
          },
        }}
        toolBarRender={() => [
          <Button key="new" type="primary" icon={<PlusOutlined />} onClick={async () => {
            setEditingRecord(null);
            form.resetFields();
            await loadTreeData();
            setModalOpen(true);
          }}>
            新增权限
          </Button>,
          <Button key="reload" icon={<ReloadOutlined />} onClick={() => actionRef.current?.reload()}>
            刷新
          </Button>,
        ]}
      />

      <Modal
        title={editingRecord ? '编辑权限' : '新增权限'}
        open={modalOpen}
        onCancel={() => setModalOpen(false)}
        onOk={() => form.submit()}
        destroyOnClose
        width={560}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={async (values) => {
            if (editingRecord) {
              await updatePermission(editingRecord.id, values);
              message.success('更新成功');
            } else {
              await createPermission(values);
              message.success('创建成功');
            }
            setModalOpen(false);
            actionRef.current?.reload();
          }}
        >
          <Form.Item name="name" label="权限名称" rules={[{ required: true }]}>
            <Input placeholder="例如：用户查看" />
          </Form.Item>
          <Form.Item name="code" label="权限编码" rules={[{ required: true }]}>
            <Input placeholder="例如：user:read" disabled={!!editingRecord} />
          </Form.Item>
          <Form.Item name="resource" label="资源" rules={[{ required: true }]}>
            <Input placeholder="例如：user" />
          </Form.Item>
          <Form.Item name="action" label="操作" rules={[{ required: true }]}>
            <Select
              placeholder="选择操作类型"
              options={[
                { label: 'read', value: 'read' },
                { label: 'create', value: 'create' },
                { label: 'update', value: 'update' },
                { label: 'delete', value: 'delete' },
                { label: 'manage', value: 'manage' },
                { label: 'export', value: 'export' },
                { label: 'approve', value: 'approve' },
              ]}
            />
          </Form.Item>
          <Form.Item name="parent_id" label="父级权限">
            <TreeSelect
              allowClear
              placeholder="无（顶级权限）"
              treeData={treeData}
              treeDefaultExpandAll
            />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={2} />
          </Form.Item>
          <Form.Item name="sort_order" label="排序" initialValue={0}>
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>
    </>
  );
}

function flattenTree(nodes: Permission[]): Permission[] {
  const result: Permission[] = [];
  const walk = (list: Permission[], level: number) => {
    for (const node of list) {
      result.push({ ...node, name: `${'— '.repeat(level)}${node.name}` });
      if (node.children) walk(node.children, level + 1);
    }
  };
  walk(nodes, 0);
  return result;
}
