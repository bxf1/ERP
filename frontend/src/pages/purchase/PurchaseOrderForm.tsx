import { useMemo } from 'react';
import { createForm } from '@formily/core';
import { FormProvider, createSchemaField } from '@formily/react';
import { Form, FormItem, Input, Select, DatePicker, NumberPicker, ArrayTable, Space, Submit, Reset } from '@formily/antd-v5';
import { Card, Button } from 'antd';
import type { PurchaseOrder, OrderItem } from '../../mock/data';

const SchemaField = createSchemaField({
  components: { Form, FormItem, Input, Select, DatePicker, NumberPicker, ArrayTable, Space },
});

interface Props {
  record: PurchaseOrder | null;
  onFinish: (values: Partial<PurchaseOrder>) => void;
  onCancel: () => void;
}

export default function PurchaseOrderForm({ record, onFinish, onCancel }: Props) {
  const form = useMemo(() => createForm({
    initialValues: record || {
      orderNo: `CG${new Date().getFullYear().toString().slice(2)}${Date.now().toString().slice(-4)}`,
      orderDate: new Date().toISOString().slice(0, 10),
      status: 'draft',
      items: [],
    },
  }), [record]);

  const schema = {
    type: 'object',
    properties: {
      orderNo: { type: 'string', title: '订单号', 'x-decorator': 'FormItem', 'x-component': 'Input', 'x-component-props': { disabled: true }, required: true },
      supplierId: { type: 'string', title: '供应商', 'x-decorator': 'FormItem', 'x-component': 'Select', 'x-component-props': { showSearch: true, placeholder: '选择供应商' }, required: true, enum: Array.from({ length: 25 }, (_, i) => ({ label: `供应商${i + 1}`, value: `sup-${i + 1}` })) },
      orderDate: { type: 'string', title: '订单日期', 'x-decorator': 'FormItem', 'x-component': 'DatePicker', required: true },
      deliveryDate: { type: 'string', title: '交货日期', 'x-decorator': 'FormItem', 'x-component': 'DatePicker', required: true },
      remark: { type: 'string', title: '备注', 'x-decorator': 'FormItem', 'x-component': 'Input.TextArea' },
      items: {
        type: 'array', title: '订单明细',
        'x-decorator': 'FormItem',
        'x-component': 'ArrayTable',
        'x-component-props': { bordered: true },
        items: {
          type: 'object',
          properties: {
            productCode: { type: 'string', title: '商品编码', 'x-decorator': 'FormItem', 'x-component': 'Input', required: true },
            productName: { type: 'string', title: '商品名称', 'x-decorator': 'FormItem', 'x-component': 'Input', required: true },
            spec: { type: 'string', title: '规格', 'x-decorator': 'FormItem', 'x-component': 'Input' },
            unit: { type: 'string', title: '单位', 'x-decorator': 'FormItem', 'x-component': 'Input' },
            quantity: { type: 'number', title: '数量', 'x-decorator': 'FormItem', 'x-component': 'NumberPicker', required: true },
            unitPrice: { type: 'number', title: '单价', 'x-decorator': 'FormItem', 'x-component': 'NumberPicker', required: true },
            remove: { type: 'void', title: '操作', 'x-decorator': 'FormItem', 'x-component': 'ArrayTable.Remove' },
          },
        },
        properties: {
          add: {
            type: 'void', title: '添加明细',
            'x-component': 'ArrayTable.Addition',
            'x-component-props': { defaultValue: { productCode: '', productName: '', spec: '', unit: '个', quantity: 0, unitPrice: 0 } },
          },
        },
      },
    },
  };

  return (
    <FormProvider form={form}>
      <Card>
        <Form layout="vertical" onAutoSave={onFinish}>
          <SchemaField schema={schema} />
          <div style={{ marginTop: 16, textAlign: 'right' }}>
            <Space>
              <Reset>重置</Reset>
              <Button onClick={onCancel}>取消</Button>
              <Submit onSubmit={(values: any) => {
                const supplierId = values.supplierId;
                const supplierName = `供应商${Number(supplierId.split('-')[1])}`;
                onFinish({ ...values, supplierId, supplierName });
              }}>保存</Submit>
            </Space>
          </div>
        </Form>
      </Card>
    </FormProvider>
  );
}
