import { useRef } from 'react';
import { ProTable, type ProColumns, type ActionType } from '@ant-design/pro-components';
import { Tag, Badge } from 'antd';
import { getInventory } from '../../services/api';
import type { InventoryRecord } from '../../mock/data';

export default function InventoryList() {
  const actionRef = useRef<ActionType>();

  const columns: ProColumns<InventoryRecord>[] = [
    { title: '商品编码', dataIndex: 'productCode', width: 120 },
    { title: '商品名称', dataIndex: 'productName', width: 150 },
    { title: '规格', dataIndex: 'spec', width: 80 },
    { title: '单位', dataIndex: 'unit', width: 60 },
    { title: '仓库', dataIndex: 'warehouseName', width: 100 },
    {
      title: '库存数量', dataIndex: 'quantity', width: 100, sorter: true,
      render: (_, record) => {
        const isLow = record.quantity < record.safeStock;
        return (
          <span>
            {record.quantity}
            {isLow && <Tag color="red" style={{ marginLeft: 8 }}>低于安全库存</Tag>}
          </span>
        );
      },
    },
    { title: '安全库存', dataIndex: 'safeStock', width: 100 },
    {
      title: '库存状态', width: 100,
      render: (_, record) => {
        if (record.quantity === 0) return <Badge status="error" text="缺货" />;
        if (record.quantity < record.safeStock) return <Badge status="warning" text="低库存" />;
        return <Badge status="success" text="正常" />;
      },
    },
    {
      title: '最近采购价', dataIndex: 'lastPurchasePrice', width: 110,
      render: (_, record) => `¥${record.lastPurchasePrice}`,
    },
    {
      title: '库存价值', dataIndex: 'totalValue', width: 130,
      render: (_, record) => `¥${record.totalValue.toLocaleString()}`,
      sorter: true,
    },
  ];

  return (
    <ProTable<InventoryRecord>
      headerTitle="库存台账"
      actionRef={actionRef}
      columns={columns}
      request={async (params) => {
        const res = await getInventory({ page: params.current, pageSize: params.pageSize, keyword: params.keyword });
        return { data: res.data, total: res.total, success: res.success };
      }}
      rowKey="id"
      search={{ labelWidth: 'auto' }}
      pagination={{ pageSize: 10 }}
      scroll={{ x: 1000 }}
    />
  );
}
