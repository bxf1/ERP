import { useMemo, useEffect } from 'react';
import { createForm, onFieldValueChange } from '@formily/core';
import { FormProvider, createSchemaField } from '@formily/react';
import {
  FormLayout,
  FormItem,
  Input,
  NumberPicker,
  Select,
  DatePicker,
  Switch,
  Radio,
  Checkbox,
  FormButtonGroup,
  Submit,
  Reset,
} from '@formily/antd-v5';
import { Button, Card, Space, Alert, Spin, Typography, message } from 'antd';
import { SaveOutlined } from '@ant-design/icons';
import { ReferenceSelector } from './fields/ReferenceSelector';
import { SubTable } from './fields/SubTable';
import { useFormState } from '../../hooks/useFormState';
import { validateSchema } from './schema';
import type { DynamicFormProps, FieldSchema, FormMode } from './types';

const { Text, Title } = Typography;

const SchemaField = createSchemaField({
  components: {
    FormLayout,
    FormItem,
    Input,
    'Input.TextArea': Input.TextArea,
    NumberPicker,
    Select,
    DatePicker,
    Switch,
    'Radio.Group': Radio.Group,
    'Checkbox.Group': Checkbox.Group,
    ReferenceSelector,
    SubTable,
  },
});

function buildFields(
  schemaFields: FieldSchema[],
  mode: FormMode,
  fetchReferences?: (referenceType: string, keyword?: string) => Promise<{ label: string; value: string | number }[]>
) {
  const isView = mode === 'view';
  const properties: Record<string, unknown> = {};

  for (const field of schemaFields) {
    if (field.hidden) continue;

    const base: Record<string, unknown> = {
      title: field.title,
      'x-decorator': 'FormItem',
    };

    if (field.description) {
      base.description = field.description;
    }

    // Determine required status
    const required = field.required || field.rules?.some((r) => r.required);

    // Map field type to Formily component
    switch (field.type) {
      case 'text':
        base['x-component'] = 'Input';
        base['x-component-props'] = { placeholder: field.placeholder || `请输入${field.title}` };
        break;
      case 'textarea':
        base['x-component'] = 'Input.TextArea';
        base['x-component-props'] = {
          placeholder: field.placeholder || `请输入${field.title}`,
          rows: (field.props as Record<string, unknown>)?.rows || 4,
        };
        break;
      case 'number':
        base['x-component'] = 'NumberPicker';
        base['x-component-props'] = { placeholder: field.placeholder || `请输入${field.title}`, style: { width: '100%' } };
        break;
      case 'date':
      case 'datetime':
        base['x-component'] = 'DatePicker';
        base['x-component-props'] = { placeholder: field.placeholder || '请选择日期', style: { width: '100%' } };
        if (field.type === 'datetime') {
          (base['x-component-props'] as Record<string, unknown>).showTime = true;
        }
        break;
      case 'select':
        base['x-component'] = 'Select';
        base['x-component-props'] = {
          placeholder: field.placeholder || `请选择${field.title}`,
          ...(field.props as Record<string, unknown>),
        };
        break;
      case 'switch':
        base['x-component'] = 'Switch';
        break;
      case 'radio':
        base['x-component'] = 'Radio.Group';
        base['x-component-props'] = { ...(field.props as Record<string, unknown>) };
        break;
      case 'checkbox':
        base['x-component'] = 'Checkbox.Group';
        base['x-component-props'] = { ...(field.props as Record<string, unknown>) };
        break;
      case 'reference_selector':
        base['x-component'] = 'ReferenceSelector';
        base['x-component-props'] = { ...(field.props as Record<string, unknown>), onFetch: fetchReferences };
        break;
      case 'sub_table':
        base['x-component'] = 'SubTable';
        base['x-component-props'] = { ...(field.props as Record<string, unknown>) };
        break;
      default:
        base['x-component'] = 'Input';
    }

    // Apply view/edit mode
    if (field.disabled || isView) {
      base['x-component-props'] = { ...(base['x-component-props'] as Record<string, unknown>), disabled: true };
    }

    // Attach default value
    if (field.defaultValue !== undefined) {
      base.default = field.defaultValue;
    }

    // Attach required
    if (required) {
      base.required = true;
    }

    // Attach validation rules
    const rules: Record<string, unknown>[] = [];
    for (const rule of field.rules || []) {
      const mapped: Record<string, unknown> = {};
      if (rule.required) mapped.required = true;
      if (rule.message) mapped.message = rule.message;
      if (rule.pattern) {
        mapped.pattern = new RegExp(rule.pattern);
        mapped.message = rule.message || '格式不正确';
      }
      if (rule.min !== undefined) mapped.minimum = rule.min;
      if (rule.max !== undefined) mapped.maximum = rule.max;
      if (rule.minLength !== undefined) mapped.minLength = rule.minLength;
      if (rule.maxLength !== undefined) mapped.maxLength = rule.maxLength;
      if (Object.keys(mapped).length > 0) rules.push(mapped);
    }
    if (rules.length > 0) {
      base['x-validator'] = rules;
    }

    properties[field.key] = base;
  }

  return { type: 'object', properties };
}

export function DynamicForm({
  schema,
  mode = 'create',
  initialValues,
  onSubmit,
  onSaveDraft,
  onValuesChange,
  onStatusChange,
  showSubmit = true,
  submitText = '提交',
  extraActions,
  fetchReferences,
  className,
  style,
}: DynamicFormProps) {
  const { status, error, submit, saveDraft } = useFormState({ onSaveDraft, onSubmit });
  const isView = mode === 'view';

  useEffect(() => {
    onStatusChange?.(status);
  }, [status, onStatusChange]);

  if (!validateSchema(schema)) {
    return <Alert type="error" message="无效的表单 Schema" showIcon />;
  }

  const hasGroups = !!(schema.groups && schema.groups.length > 0);

  const groupedSchemas = useMemo(() => {
    if (!hasGroups) return null;
    return schema.groups!.map((group) => ({
      title: group.title,
      description: group.description,
      schema: buildFields(group.fields, mode, fetchReferences),
    }));
  }, [schema.groups, mode, fetchReferences, hasGroups]);

  const ungroupedSchema = useMemo(() => {
    if (schema.fields.length === 0) return null;
    return buildFields(schema.fields, mode, fetchReferences);
  }, [schema.fields, mode, fetchReferences]);

  const form = useMemo(() => {
    const f = createForm({
      initialValues,
      effects() {
        onFieldValueChange('*', () => {
          if (onValuesChange) {
            onValuesChange({ ...f.values });
          }
        });
      },
    });
    return f;
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const handleSaveDraft = async () => {
    try {
      await form.validate();
      await saveDraft(form.values);
      message.success('草稿已保存');
    } catch {
      message.warning('请先修正表单中的错误');
    }
  };

  const handleSubmit = async (formValues: Record<string, unknown>) => {
    await submit(formValues);
    message.success('提交成功');
  };

  return (
    <div className={className} style={style}>
      {schema.title && (
        <Title level={4} style={{ marginBottom: 4 }}>
          {schema.title}
        </Title>
      )}
      {schema.description && (
        <Text type="secondary" style={{ display: 'block', marginBottom: 16 }}>
          {schema.description}
        </Text>
      )}

      {error && (
        <Alert type="error" message={error} showIcon closable style={{ marginBottom: 16 }} />
      )}

      {status === 'submitting' && (
        <div style={{ textAlign: 'center', padding: 40 }}>
          <Spin size="large" tip="提交中..." />
        </div>
      )}

      {status === 'submitted' && (
        <Card style={{ marginBottom: 16 }}>
          <Alert type="success" message="表单已成功提交" showIcon />
        </Card>
      )}

      <FormProvider form={form}>
        {hasGroups && groupedSchemas ? (
          <>
            {groupedSchemas.map((group) => (
              <Card key={group.title} title={group.title} style={{ marginBottom: 16 }}>
                {group.description && (
                  <Text type="secondary" style={{ display: 'block', marginBottom: 12 }}>
                    {group.description}
                  </Text>
                )}
                <FormLayout labelCol={5} wrapperCol={19} colon>
                  {/* eslint-disable-next-line @typescript-eslint/no-explicit-any */}
                  <SchemaField schema={group.schema as any} />
                </FormLayout>
              </Card>
            ))}
            {ungroupedSchema && Object.keys(ungroupedSchema.properties).length > 0 && (
              <Card title="其他" style={{ marginBottom: 16 }}>
                <FormLayout labelCol={5} wrapperCol={19} colon>
                  {/* eslint-disable-next-line @typescript-eslint/no-explicit-any */}
                  <SchemaField schema={ungroupedSchema as any} />
                </FormLayout>
              </Card>
            )}
          </>
        ) : (
          <FormLayout labelCol={5} wrapperCol={19} colon>
            {/* eslint-disable-next-line @typescript-eslint/no-explicit-any */}
            {ungroupedSchema && <SchemaField schema={ungroupedSchema as any} />}
          </FormLayout>
        )}

        {!isView && status !== 'submitted' && (
          <FormButtonGroup align="right" style={{ marginTop: 24 }}>
            {onSaveDraft && (
              <Button icon={<SaveOutlined />} onClick={handleSaveDraft} loading={status === 'submitting'}>
                保存草稿
              </Button>
            )}
            {extraActions}
            <Reset>重置</Reset>
            {showSubmit && (
              <Submit onSubmit={handleSubmit} loading={status === 'submitting'}>
                {submitText}
              </Submit>
            )}
          </FormButtonGroup>
        )}
      </FormProvider>

      {isView && (
        <Space style={{ marginTop: 16 }}>
          <Text type="secondary">只读模式</Text>
        </Space>
      )}
    </div>
  );
}

export type { DynamicFormProps, FormSchema, FieldSchema, FormMode, FieldType, FormStatus } from './types';
export { validateSchema } from './schema';
