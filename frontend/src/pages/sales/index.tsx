import { useRef, useState } from 'react';
import { ProTable, type ProColumns, type ActionType } from '@ant-design/pro-components';
import { Button, Tag, Modal, message, Descriptions, Badge } from 'antd';
import { PlusOutlined, EyeOutlined } from '@ant-design/icons';
import { getSalesOrders, createSalesOrder, updateSalesOrder } from '../../services/api';
import type { SalesOrder, OrderItem } from '../../mock/data';
import SalesOrderForm from './SalesOrderForm';

const statusMap: Record<string, { color: string; text: string }> = {
  draft: { color: 'default', text: '草稿' },
  submitted: { color: 'processing', text: '已提交' },
  approved: { color: 'success', text: '已审核' },
  shipped: { color: 'blue', text: '已发货' },
  cancelled: { color: 'error', text: '已取消' },
};

export default function SalesOrderList() {
  const actionRef = useRef<ActionType>();
  const [modalOpen, setModalOpen] = useState(false);
  const [detailOpen, setDetailOpen] = useState(false);
  const [editingRecord, setEditingRecord] = useState<SalesOrder | null>(null);
  const [detailRecord, setDetailRecord] = useState<SalesOrder | null>(null);

  const columns: ProColumns<SalesOrder>[] = [
    { title: '订单号', dataIndex: 'orderNo', width: 140 },
    { title: '客户', dataIndex: 'customerName', width: 150 },
    { title: '订单日期', dataIndex: 'orderDate', width: 100 },
    { title: '交货日期', dataIndex: 'deliveryDate', width: 100 },
    {
      title: '金额', dataIndex: 'totalAmount', width: 120,
      render: (_, record) => `¥${record.totalAmount.toLocaleString()}`,
    },
    {
      title: '状态', dataIndex: 'status', width: 80,
      render: (_, record) => <Badge status={statusMap[record.status]?.color as any} text={statusMap[record.status]?.text} />,
    },
    { title: '备注', dataIndex: 'remark', ellipsis: true },
    {
      title: '操作', valueType: 'option', width: 180,
      render: (_, record) => [
        <a key="view" onClick={() => { setDetailRecord(record); setDetailOpen(true); }}><EyeOutlined /> 查看</a>,
        record.status === 'draft' && <a key="edit" onClick={() => { setEditingRecord(record); setModalOpen(true); }}>编辑</a>,
        record.status === 'submitted' && <a key="approve" onClick={async () => {
          await updateSalesOrder(record.id, { status: 'approved' });
          message.success('已审核'); actionRef.current?.reload();
        }}>审核</a>,
        record.status === 'approved' && <a key="ship" onClick={async () => {
          await updateSalesOrder(record.id, { status: 'shipped' });
          message.success('已发货'); actionRef.current?.reload();
        }}>发货</a>,
        record.status === 'submitted' && <a key="reject" style={{ color: 'red' }} onClick={async () => {
          await updateSalesOrder(record.id, { status: 'cancelled' });
          message.success('已取消'); actionRef.current?.reload();
        }}>取消</a>,
      ],
    },
  ];

  return (
    <>
      <ProTable<SalesOrder>
        headerTitle="销售订单列表"
        actionRef={actionRef}
        columns={columns}
        request={async (params) => {
          const res = await getSalesOrders({ page: params.current, pageSize: params.pageSize, status: params.status as string });
          return { data: res.data, total: res.total, success: res.success };
        }}
        rowKey="id"
        search={{ labelWidth: 'auto' }}
        toolBarRender={() => [
          <Button key="new" type="primary" icon={<PlusOutlined />} onClick={() => { setEditingRecord(null); setModalOpen(true); }}>新增销售订单</Button>,
        ]}
        pagination={{ pageSize: 10 }}
      />

      <Modal
        title={editingRecord ? '编辑销售订单' : '新增销售订单'}
        open={modalOpen}
        onCancel={() => setModalOpen(false)}
        footer={null}
        destroyOnClose
        width={900}
      >
        <SalesOrderForm
          record={editingRecord}
          onFinish={async (values) => {
            const totalAmount = (values.items || []).reduce((sum: number, item: OrderItem) => sum + (item.quantity || 0) * (item.unitPrice || 0), 0);
            const payload = { ...values, totalAmount };
            if (editingRecord) { await updateSalesOrder(editingRecord.id, payload); }
            else { await createSalesOrder(payload); }
            message.success(editingRecord ? '更新成功' : '创建成功');
            setModalOpen(false);
            actionRef.current?.reload();
          }}
          onCancel={() => setModalOpen(false)}
        />
      </Modal>

      <Modal title="销售订单详情" open={detailOpen} onCancel={() => setDetailOpen(false)} footer={null} width={800}>
        {detailRecord && (
          <Descriptions column={2} bordered size="small">
            <Descriptions.Item label="订单号">{detailRecord.orderNo}</Descriptions.Item>
            <Descriptions.Item label="客户">{detailRecord.customerName}</Descriptions.Item>
            <Descriptions.Item label="订单日期">{detailRecord.orderDate}</Descriptions.Item>
            <Descriptions.Item label="交货日期">{detailRecord.deliveryDate}</Descriptions.Item>
            <Descriptions.Item label="状态"><Badge status={statusMap[detailRecord.status]?.color as any} text={statusMap[detailRecord.status]?.text} /></Descriptions.Item>
            <Descriptions.Item label="总金额">¥{detailRecord.totalAmount.toLocaleString()}</Descriptions.Item>
            <Descriptions.Item label="备注" span={2}>{detailRecord.remark || '无'}</Descriptions.Item>
          </Descriptions>
        )}
        {detailRecord?.items && detailRecord.items.length > 0 && (
          <ProTable<OrderItem>
            style={{ marginTop: 16 }}
            headerTitle="订单明细"
            dataSource={detailRecord.items}
            columns={[
              { title: '商品编码', dataIndex: 'productCode', width: 100 },
              { title: '商品名称', dataIndex: 'productName', width: 100 },
              { title: '规格', dataIndex: 'spec', width: 80 },
              { title: '单位', dataIndex: 'unit', width: 60 },
              { title: '数量', dataIndex: 'quantity', width: 80 },
              { title: '单价', dataIndex: 'unitPrice', width: 80, render: (_, r) => `¥${r.unitPrice}` },
              { title: '金额', dataIndex: 'amount', width: 100, render: (_, r) => `¥${r.amount.toLocaleString()}` },
            ]}
            rowKey="id"
            search={false}
            options={false}
            pagination={false}
          />
        )}
      </Modal>
    </>
  );
}
