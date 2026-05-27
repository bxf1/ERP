import { useEffect, useState } from 'react';
import { Card, Row, Col, Statistic, Table, Spin } from 'antd';
import { ShoppingCartOutlined, DollarOutlined, DatabaseOutlined, FileTextOutlined, ArrowUpOutlined } from '@ant-design/icons';
import { Column, Line, Pie } from '@ant-design/charts';
import { getDashboardData } from '../../services/api';
import type { dashboardData as DashboardDataType } from '../../mock/data';

export default function Dashboard() {
  const [data, setData] = useState<typeof DashboardDataType | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    getDashboardData().then(res => { setData(res.data); setLoading(false); });
  }, []);

  if (loading || !data) return <Spin size="large" style={{ display: 'block', marginTop: 200 }} />;

  const purchaseTrendConfig = {
    data: data.purchaseTrend,
    xField: 'month',
    yField: 'amount',
    smooth: true,
    meta: { amount: { alias: '采购额' } },
    color: '#1677ff',
    point: { size: 5, shape: 'circle' },
  };

  const salesTrendConfig = {
    data: data.salesTrend,
    xField: 'month',
    yField: 'amount',
    smooth: true,
    meta: { amount: { alias: '销售额' } },
    color: '#52c41a',
    point: { size: 5, shape: 'circle' },
  };

  const stockPieConfig = {
    data: data.stockByCategory,
    angleField: 'value',
    colorField: 'category',
    radius: 0.8,
    label: { type: 'outer', content: '{name}\n¥{value}' },
    legend: { position: 'bottom' as const },
  };

  const topMovingConfig = {
    data: data.topMovingProducts,
    xField: 'name',
    yField: 'quantity',
    color: '#fa8c16',
    label: { position: 'top' as const },
    meta: { quantity: { alias: '出库数量' } },
  };

  const topColumns = [
    { title: '商品', dataIndex: 'name', key: 'name' },
    { title: '出库数量', dataIndex: 'quantity', key: 'quantity', sorter: (a: any, b: any) => a.quantity - b.quantity },
    { title: '出库金额', dataIndex: 'amount', key: 'amount', render: (v: number) => `¥${v.toLocaleString()}`, sorter: (a: any, b: any) => a.amount - b.amount },
  ];

  return (
    <div>
      <Row gutter={[16, 16]}>
        <Col xs={24} sm={12} md={6}>
          <Card hoverable><Statistic title="本月采购额" value={data.monthlyPurchase} prefix={<ShoppingCartOutlined />} suffix="元" valueStyle={{ color: '#1677ff' }} /></Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card hoverable><Statistic title="本月销售额" value={data.monthlySales} prefix={<DollarOutlined />} suffix="元" valueStyle={{ color: '#52c41a' }} /></Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card hoverable><Statistic title="库存总值" value={data.inventoryValue} prefix={<DatabaseOutlined />} suffix="元" valueStyle={{ color: '#fa8c16' }} /></Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card hoverable><Statistic title="待处理订单" value={data.pendingOrders} prefix={<FileTextOutlined />} suffix="单" valueStyle={{ color: '#eb2f96' }} /></Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} md={12}>
          <Card title="采购趋势（1-5月）"><Line {...purchaseTrendConfig} height={280} /></Card>
        </Col>
        <Col xs={24} md={12}>
          <Card title="销售趋势（1-5月）"><Line {...salesTrendConfig} height={280} /></Card>
        </Col>
      </Row>

      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col xs={24} md={10}>
          <Card title="库存分类占比"><Pie {...stockPieConfig} height={280} /></Card>
        </Col>
        <Col xs={24} md={14}>
          <Card title="热销商品 Top 5">
            <Column {...topMovingConfig} height={280} />
          </Card>
        </Col>
      </Row>

      <Card title="热销商品明细" style={{ marginTop: 16 }}>
        <Table columns={topColumns} dataSource={data.topMovingProducts} rowKey="name" pagination={false} size="small" />
      </Card>
    </div>
  );
}
