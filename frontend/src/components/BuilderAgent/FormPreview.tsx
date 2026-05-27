import React from 'react';
import {
  Card,
  Form,
  Input,
  InputNumber,
  DatePicker,
  Select,
  Row,
  Col,
  Tag,
  Typography,
} from 'antd';
import type { FormFieldConfig } from '@/types';

const { TextArea } = Input;
const { Text } = Typography;

interface Props {
  fields: FormFieldConfig[];
  moduleName: string;
}

const fieldTypeLabel: Record<string, string> = {
  string: '文本',
  number: '数字',
  boolean: '布尔',
  date: '日期',
  datetime: '日期时间',
  select: '下拉选择',
  multiSelect: '多选',
  textarea: '长文本',
  file: '文件',
  image: '图片',
  richtext: '富文本',
  user: '用户选择',
  department: '部门选择',
};

const FormPreview: React.FC<Props> = ({ fields, moduleName }) => {
  return (
    <Card
      title={`表单预览 — ${moduleName}`}
      size="small"
      style={{ marginBottom: 16 }}
    >
      <Form layout="vertical" size="small">
        <Row gutter={16}>
          {fields
            .filter((f) => f.visible)
            .sort((a, b) => a.order - b.order)
            .map((field) => (
              <Col span={field.span} key={field.key}>
                <Form.Item
                  label={
                    <span>
                      {field.label}
                      {field.required && (
                        <Text type="danger" style={{ marginLeft: 2 }}>
                          *
                        </Text>
                      )}
                      <Tag
                        style={{ marginLeft: 8, fontSize: 11 }}
                        color="blue"
                      >
                        {fieldTypeLabel[field.type] || field.type}
                      </Tag>
                    </span>
                  }
                  rules={
                    field.required
                      ? [{ required: true, message: `请输入${field.label}` }]
                      : undefined
                  }
                >
                  {renderFieldInput(field)}
                </Form.Item>
              </Col>
            ))}
        </Row>
      </Form>
    </Card>
  );
};

function renderFieldInput(field: FormFieldConfig) {
  switch (field.type) {
    case 'number':
      return <InputNumber style={{ width: '100%' }} placeholder={field.placeholder} />;
    case 'date':
    case 'datetime':
      return <DatePicker style={{ width: '100%' }} placeholder={field.placeholder} />;
    case 'select':
      return (
        <Select
          placeholder={field.placeholder}
          options={field.options}
          style={{ width: '100%' }}
        />
      );
    case 'multiSelect':
      return (
        <Select
          mode="multiple"
          placeholder={field.placeholder}
          options={field.options}
          style={{ width: '100%' }}
        />
      );
    case 'textarea':
      return <TextArea rows={3} placeholder={field.placeholder} />;
    case 'richtext':
      return <TextArea rows={4} placeholder={field.placeholder || '富文本内容…'} />;
    default:
      return <Input placeholder={field.placeholder} />;
  }
}

export default FormPreview;
