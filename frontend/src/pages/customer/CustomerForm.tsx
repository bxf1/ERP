import { useEffect } from 'react';
import { Form, Input, InputNumber, Select, Button, Space } from 'antd';
import type { Customer } from '../../mock/data';

interface Props {
  record: Customer | null;
  onFinish: (values: Partial<Customer>) => void;
  onCancel: () => void;
}

export default function CustomerForm({ record, onFinish, onCancel }: Props) {
  const [form] = Form.useForm();

  useEffect(() => {
    if (record) form.setFieldsValue(record);
    else form.resetFields();
  }, [record, form]);

  return (
    <Form form={form} layout="vertical" onFinish={onFinish}>
      <Form.Item name="code" label="客户编码" rules={[{ required: true, message: '请输入编码' }]}>
        <Input placeholder="如 KH0001" />
      </Form.Item>
      <Form.Item name="name" label="客户名称" rules={[{ required: true, message: '请输入名称' }]}>
        <Input placeholder="客户名称" />
      </Form.Item>
      <Form.Item name="contact" label="联系人">
        <Input placeholder="联系人" />
      </Form.Item>
      <Form.Item name="phone" label="联系电话">
        <Input placeholder="联系电话" />
      </Form.Item>
      <Form.Item name="email" label="邮箱">
        <Input placeholder="邮箱地址" />
      </Form.Item>
      <Form.Item name="address" label="地址">
        <Input placeholder="公司地址" />
      </Form.Item>
      <Form.Item name="creditLimit" label="信用额度" initialValue={50000}>
        <InputNumber style={{ width: '100%' }} min={0} step={10000} placeholder="信用额度" />
      </Form.Item>
      <Form.Item name="status" label="状态" initialValue="active">
        <Select options={[{ value: 'active', label: '启用' }, { value: 'inactive', label: '停用' }]} />
      </Form.Item>
      <Form.Item style={{ marginBottom: 0, textAlign: 'right' }}>
        <Space>
          <Button onClick={onCancel}>取消</Button>
          <Button type="primary" htmlType="submit">保存</Button>
        </Space>
      </Form.Item>
    </Form>
  );
}
