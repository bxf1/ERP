import { useRef, useState } from 'react';
import { ProTable, type ProColumns, type ActionType } from '@ant-design/pro-components';
import { Button, Tag, Modal, message, Descriptions, Progress } from 'antd';
import { PlusOutlined, EyeOutlined } from '@ant-design/icons';
import { getStocktakingList, createStocktaking, getStocktaking } from '../../services/api';
import type { Stocktaking, StocktakingItem } from '../../mock/data';

const statusMap: Record<string, { color: string; text: string }> = {
  pending: { color: 'default', text: '待盘点' },
  in_progress: { color: 'processing', text: '盘点中' },
  completed: { color: 'success', text: '已完成' },
};

export default function StocktakingList() {
  const actionRef = useRef<ActionType>();
  const [detailOpen, setDetailOpen] = useState(false);
  const [detailRecord, setDetailRecord] = useState<Stocktaking | null>(null);

  const columns: ProColumns<Stocktaking>[] = [
    { title: '任务编号', dataIndex: 'taskNo', width: 140 },
    { title: '盘点仓库', dataIndex: 'warehouseName', width: 120 },
    { title: '开始日期', dataIndex: 'startDate', width: 100 },
    { title: '结束日期', dataIndex: 'endDate', width: 100 },
    {
      title: '状态', dataIndex: 'status', width: 80,
      render: (_, record) => <Tag color={statusMap[record.status]?.color}>{statusMap[record.status]?.text}</Tag>,
    },
    {
      title: '差异项', width: 80,
      render: (_, record) => {
        const diffCount = record.items.filter(i => i.difference !== 0).length;
        return diffCount > 0 ? <Tag color="orange">{diffCount}项差异</Tag> : <Tag color="green">无差异</Tag>;
      },
    },
    { title: '创建日期', dataIndex: 'createdAt', width: 110 },
    {
      title: '操作', valueType: 'option', width: 120,
      render: (_, record) => [
        <a key="view" onClick={async () => {
          const res = await getStocktaking(record.id);
          if (res.data) { setDetailRecord(res.data); setDetailOpen(true); }
        }}><EyeOutlined /> 查看</a>,
        record.status === 'in_progress' && <a key="complete" onClick={async () => {
          message.success('盘点完成');
          actionRef.current?.reload();
        }}>完成</a>,
      ],
    },
  ];

  const diffItems = detailRecord?.items.filter(i => i.difference !== 0) || [];
  const progressValue = detailRecord ? ((detailRecord.items.filter(i => i.difference === 0).length / detailRecord.items.length) * 100) : 0;

  return (
    <>
      <ProTable<Stocktaking>
        headerTitle="库存盘点"
        actionRef={actionRef}
        columns={columns}
        request={async (params) => {
          const res = await getStocktakingList({ page: params.current, pageSize: params.pageSize });
          return { data: res.data, total: res.total, success: res.success };
        }}
        rowKey="id"
        search={false}
        toolBarRender={() => [
          <Button key="new" type="primary" icon={<PlusOutlined />} onClick={async () => {
            const warehouseName = ['仓库A', '仓库B', '仓库C'][Math.floor(Math.random() * 3)];
            await createStocktaking({
              taskNo: `PD${new Date().getFullYear().toString().slice(2)}${Date.now().toString().slice(-4)}`,
              warehouseId: 'wh-1',
              warehouseName,
              startDate: new Date().toISOString().slice(0, 10),
              status: 'pending',
              items: Array.from({ length: 10 }, (_, j) => {
                const book = Math.floor(Math.random() * 500);
                const actual = book + (Math.random() > 0.8 ? Math.floor(Math.random() * 10 - 5) : 0);
                return { productCode: `CP${String(j + 1).padStart(4, '0')}`, productName: `商品${j + 1}`, bookQuantity: book, actualQuantity, difference: actual - book, remark: actual !== book ? '盘点差异' : '' };
              }),
            });
            message.success('已创建盘点任务');
            actionRef.current?.reload();
          }}>新增盘点任务</Button>,
        ]}
        pagination={{ pageSize: 10 }}
      />

      <Modal title="盘点详情" open={detailOpen} onCancel={() => setDetailOpen(false)} footer={null} width={900}>
        {detailRecord && (
          <>
            <Descriptions column={3} bordered size="small" style={{ marginBottom: 16 }}>
              <Descriptions.Item label="任务编号">{detailRecord.taskNo}</Descriptions.Item>
              <Descriptions.Item label="盘点仓库">{detailRecord.warehouseName}</Descriptions.Item>
              <Descriptions.Item label="状态"><Tag color={statusMap[detailRecord.status]?.color}>{statusMap[detailRecord.status]?.text}</Tag></Descriptions.Item>
              <Descriptions.Item label="开始日期">{detailRecord.startDate}</Descriptions.Item>
              <Descriptions.Item label="结束日期">{detailRecord.endDate}</Descriptions.Item>
              <Descriptions.Item label="盘点进度"><Progress percent={Math.round(progressValue)} size="small" /></Descriptions.Item>
            </Descriptions>
            {diffItems.length > 0 && (
              <ProTable<StocktakingItem>
                headerTitle={`盘点差异（共${diffItems.length}项）`}
                dataSource={diffItems}
                columns={[
                  { title: '商品编码', dataIndex: 'productCode', width: 100 },
                  { title: '商品名称', dataIndex: 'productName', width: 100 },
                  { title: '账面数量', dataIndex: 'bookQuantity', width: 90 },
                  { title: '实盘数量', dataIndex: 'actualQuantity', width: 90 },
                  {
                    title: '差异', dataIndex: 'difference', width: 80,
                    render: (_, r) => <span style={{ color: r.difference > 0 ? 'green' : 'red', fontWeight: 600 }}>{r.difference > 0 ? '+' : ''}{r.difference}</span>,
                  },
                  { title: '备注', dataIndex: 'remark', ellipsis: true },
                ]}
                rowKey="productCode"
                search={false}
                options={false}
                pagination={false}
              />
            )}
          </>
        )}
      </Modal>
    </>
  );
}
