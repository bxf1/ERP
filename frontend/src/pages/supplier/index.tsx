import { useRef, useState, type MutableRefObject } from 'react';
import { ProTable, type ProColumns, type ActionType } from '@ant-design/pro-components';
import { Button, Tag, Modal, message, Popconfirm } from 'antd';
import { PlusOutlined, ReloadOutlined } from '@ant-design/icons';
import { getSuppliers, createSupplier, updateSupplier } from '../../services/api';
import type { Supplier } from '../../mock/data';
import SupplierForm from './SupplierForm';

export default function SupplierList() {
  const actionRef = useRef<ActionType>();
  const [modalOpen, setModalOpen] = useState(false);
  const [editingRecord, setEditingRecord] = useState<Supplier | null>(null);

  const columns: ProColumns<Supplier>[] = [
    { title: '供应商编码', dataIndex: 'code', width: 120 },
    { title: '供应商名称', dataIndex: 'name', width: 150 },
    { title: '联系人', dataIndex: 'contact', width: 100 },
    { title: '电话', dataIndex: 'phone', width: 140 },
    { title: '邮箱', dataIndex: 'email', width: 180 },
    { title: '地址', dataIndex: 'address', ellipsis: true },
    { title: '开户银行账号', dataIndex: 'bankAccount', width: 180 },
    { title: '税号', dataIndex: 'taxId', width: 160 },
    {
      title: '状态', dataIndex: 'status', width: 80,
      render: (_, record) => <Tag color={record.status === 'active' ? 'green' : 'red'}>{record.status === 'active' ? '启用' : '停用'}</Tag>,
    },
    { title: '创建日期', dataIndex: 'createdAt', width: 110 },
    {
      title: '操作', valueType: 'option', width: 120,
      render: (_, record) => [
        <a key="edit" onClick={() => { setEditingRecord(record); setModalOpen(true); }}>编辑</a>,
        <Popconfirm key="toggle" title={`确认${record.status === 'active' ? '停用' : '启用'}？`} onConfirm={async () => {
          await updateSupplier(record.id, { status: record.status === 'active' ? 'inactive' : 'active' });
          message.success('操作成功'); (actionRef.current as ActionType)?.reload();
        }}><a>{record.status === 'active' ? '停用' : '启用'}</a></Popconfirm>,
      ],
    },
  ];

  return (
    <>
      <ProTable<Supplier>
        headerTitle="供应商列表"
        actionRef={actionRef}
        columns={columns}
        request={async (params) => {
          const res = await getSuppliers({ page: params.current, pageSize: params.pageSize, keyword: params.keyword });
          return { data: res.data, total: res.total, success: res.success };
        }}
        rowKey="id"
        search={{ labelWidth: 'auto' }}
        toolBarRender={() => [
          <Button key="new" type="primary" icon={<PlusOutlined />} onClick={() => { setEditingRecord(null); setModalOpen(true); }}>新增供应商</Button>,
          <Button key="reload" icon={<ReloadOutlined />} onClick={() => (actionRef.current as ActionType)?.reload()}>刷新</Button>,
        ]}
        pagination={{ pageSize: 10 }}
      />
      <Modal
        title={editingRecord ? '编辑供应商' : '新增供应商'}
        open={modalOpen}
        onCancel={() => setModalOpen(false)}
        footer={null}
        destroyOnClose
        width={640}
      >
        <SupplierForm
          record={editingRecord}
          onFinish={async (values) => {
            if (editingRecord) { await updateSupplier(editingRecord.id, values); }
            else { await createSupplier(values); }
            message.success(editingRecord ? '更新成功' : '创建成功');
            setModalOpen(false);
            (actionRef.current as ActionType)?.reload();
          }}
          onCancel={() => setModalOpen(false)}
        />
      </Modal>
    </>
  );
}
