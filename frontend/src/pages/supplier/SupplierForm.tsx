import { useEffect } from 'react';
import { Form, Input, Select, Button, Space } from 'antd';
import type { Supplier } from '../../mock/data';

interface Props {
  record: Supplier | null;
  onFinish: (values: Partial<Supplier>) => void;
  onCancel: () => void;
}

export default function SupplierForm({ record, onFinish, onCancel }: Props) {
  const [form] = Form.useForm();

  useEffect(() => {
    if (record) form.setFieldsValue(record);
    else form.resetFields();
  }, [record, form]);

  return (
    <Form form={form} layout="vertical" onFinish={onFinish}>
      <Form.Item name="code" label="供应商编码" rules={[{ required: true, message: '请输入编码' }]}>
        <Input placeholder="如 GYS0001" />
      </Form.Item>
      <Form.Item name="name" label="供应商名称" rules={[{ required: true, message: '请输入名称' }]}>
        <Input placeholder="供应商名称" />
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
      <Form.Item name="bankAccount" label="开户银行账号">
        <Input placeholder="银行账号" />
      </Form.Item>
      <Form.Item name="taxId" label="税号">
        <Input placeholder="税务登记号" />
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
