import { useRef, useState } from 'react';
import { ProTable } from '@ant-design/pro-components';
import type { ProColumns, ActionType } from '@ant-design/pro-components';
import { Button, Tag, Modal, Form, Input, InputNumber, Popconfirm, Tree, message, Spin } from 'antd';
import { PlusOutlined, ReloadOutlined, SafetyOutlined } from '@ant-design/icons';
import type { Role, Permission } from '../../types/permission';
import {
  getRoles,
  createRole,
  updateRole,
  deleteRole,
  assignPermissions,
  getPermissions,
  getRole,
} from '../../services/permission';

function buildPermissionTree(permissions: Permission[]): any[] {
  return permissions.map((p) => ({
    title: `${p.name} (${p.code})`,
    key: p.id,
    children: p.children ? buildPermissionTree(p.children) : [],
  }));
}

function getAllPermissionIds(permissions: Permission[]): string[] {
  const ids: string[] = [];
  const walk = (list: Permission[]) => {
    for (const p of list) {
      ids.push(p.id);
      if (p.children) walk(p.children);
    }
  };
  walk(permissions);
  return ids;
}

export default function RoleList() {
  const actionRef = useRef<ActionType>();
  const [modalOpen, setModalOpen] = useState(false);
  const [permModalOpen, setPermModalOpen] = useState(false);
  const [editingRecord, setEditingRecord] = useState<Role | null>(null);
  const [permRole, setPermRole] = useState<Role | null>(null);
  const [permTree, setPermTree] = useState<any[]>([]);
  const [allPerms, setAllPerms] = useState<Permission[]>([]);
  const [checkedKeys, setCheckedKeys] = useState<string[]>([]);
  const [permLoading, setPermLoading] = useState(false);
  const [confirmLoading, setConfirmLoading] = useState(false);
  const [form] = Form.useForm();

  const columns: ProColumns<Role>[] = [
    { title: '角色名称', dataIndex: 'name', width: 150 },
    { title: '角色编码', dataIndex: 'code', width: 140, copyable: true },
    { title: '描述', dataIndex: 'description', ellipsis: true },
    { title: '排序', dataIndex: 'sort_order', width: 80 },
    {
      title: '状态', dataIndex: 'status', width: 80,
      render: (_, record) => (
        <Tag color={record.status === 'active' ? 'green' : 'red'}>
          {record.status === 'active' ? '启用' : '停用'}
        </Tag>
      ),
    },
    {
      title: '系统角色', dataIndex: 'is_system', width: 90,
      render: (_, record) => (record.is_system ? <Tag color="blue">系统</Tag> : <Tag>自定义</Tag>),
    },
    {
      title: '创建时间', dataIndex: 'created_at', width: 170,
      render: (_, record) => new Date(record.created_at).toLocaleString(),
    },
    {
      title: '操作', valueType: 'option', width: 220,
      render: (_, record) => [
        <a key="perm" onClick={() => openPermModal(record)}>
          <SafetyOutlined /> 权限
        </a>,
        <a key="edit" onClick={() => { setEditingRecord(record); form.setFieldsValue(record); setModalOpen(true); }}>
          编辑
        </a>,
        record.is_system ? (
          <span key="del" style={{ color: '#ccc' }}>删除</span>
        ) : (
          <Popconfirm key="del" title="确认删除此角色？" onConfirm={async () => {
            await deleteRole(record.id);
            message.success('已删除');
            actionRef.current?.reload();
          }}>
            <a>删除</a>
          </Popconfirm>
        ),
      ],
    },
  ];

  const openPermModal = async (role: Role) => {
    setPermRole(role);
    setPermModalOpen(true);
    setPermLoading(true);
    try {
      const [perms, roleDetail] = await Promise.all([
        getPermissions(),
        getRole(role.id),
      ]);
      setAllPerms(perms);
      setPermTree(buildPermissionTree(perms));
      const existingIds = (roleDetail.permissions || []).map((p: Permission) => p.id);
      setCheckedKeys(existingIds);
    } finally {
      setPermLoading(false);
    }
  };

  const handlePermSubmit = async () => {
    if (!permRole) return;
    // Only leaf nodes matter for the backend, but we send what's checked
    // Deduplicate: if parent is checked, we only send leaf descendants
    const leafIds = checkedKeys.filter((id) => {
      const findNode = (list: Permission[]): Permission | undefined => {
        for (const p of list) {
          if (p.id === id) return p;
          if (p.children) {
            const found = findNode(p.children);
            if (found) return found;
          }
        }
        return undefined;
      };
      const node = findNode(allPerms);
      return node && (!node.children || node.children.length === 0);
    });

    // If no leaves selected but intermediates are, include them
    const idsToSend = leafIds.length > 0 ? leafIds : checkedKeys;

    await assignPermissions(permRole.id, idsToSend);
    message.success('权限分配成功');
    setPermModalOpen(false);
    actionRef.current?.reload();
  };

  return (
    <>
      <ProTable<Role>
        headerTitle="角色管理"
        actionRef={actionRef}
        columns={columns}
        request={async () => {
          const data = await getRoles();
          return { data, total: data.length, success: true };
        }}
        rowKey="id"
        search={false}
        toolBarRender={() => [
          <Button key="new" type="primary" icon={<PlusOutlined />} onClick={() => {
            setEditingRecord(null);
            form.resetFields();
            setModalOpen(true);
          }}>
            新增角色
          </Button>,
          <Button key="reload" icon={<ReloadOutlined />} onClick={() => actionRef.current?.reload()}>
            刷新
          </Button>,
        ]}
      />

      {/* Create / Edit Role Modal */}
      <Modal
        title={editingRecord ? '编辑角色' : '新增角色'}
        open={modalOpen}
        onCancel={() => setModalOpen(false)}
        confirmLoading={confirmLoading}
        onOk={() => form.submit()}
        destroyOnClose
        width={520}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={async (values) => {
            setConfirmLoading(true);
            try {
              if (editingRecord) {
                await updateRole(editingRecord.id, values);
                message.success('更新成功');
              } else {
                await createRole(values);
                message.success('创建成功');
              }
              setModalOpen(false);
              actionRef.current?.reload();
            } finally {
              setConfirmLoading(false);
            }
          }}
        >
          <Form.Item name="name" label="角色名称" rules={[{ required: true, message: '请输入角色名称' }]}>
            <Input placeholder="例如：财务主管" />
          </Form.Item>
          <Form.Item name="code" label="角色编码" rules={[{ required: true, message: '请输入角色编码' }]}>
            <Input placeholder="例如：finance_manager" disabled={!!editingRecord} />
          </Form.Item>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={3} placeholder="角色职责描述" />
          </Form.Item>
          <Form.Item name="sort_order" label="排序" initialValue={0}>
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>

      {/* Permission Assignment Modal */}
      <Modal
        title={`权限分配 — ${permRole?.name || ''}`}
        open={permModalOpen}
        onCancel={() => setPermModalOpen(false)}
        onOk={handlePermSubmit}
        width={560}
        destroyOnClose
      >
        {permLoading ? (
          <div style={{ textAlign: 'center', padding: 40 }}><Spin /></div>
        ) : (
          <Tree
            checkable
            defaultExpandAll
            checkedKeys={checkedKeys}
            onCheck={(keys) => setCheckedKeys(keys as string[])}
            treeData={permTree}
          />
        )}
      </Modal>
    </>
  );
}
