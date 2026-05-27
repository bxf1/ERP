import type { Meta, StoryObj } from '@storybook/react';
import { DynamicForm } from './index';

const fn = () => {
  // eslint-disable-next-line no-console
  return (...args: unknown[]) => console.log('Action:', ...args);
};
import type { FormSchema } from './types';

const meta: Meta<typeof DynamicForm> = {
  title: 'DynamicForm/DynamicForm',
  component: DynamicForm,
  tags: ['autodocs'],
  argTypes: {
    mode: {
      control: 'radio',
      options: ['create', 'edit', 'view'],
    },
    showSubmit: { control: 'boolean' },
    submitText: { control: 'text' },
  },
};

export default meta;
type Story = StoryObj<typeof DynamicForm>;

/** 基础表单：文本、数字、日期、下拉选择 */
const basicSchema: FormSchema = {
  title: '员工基本信息',
  description: '请填写员工的基本信息',
  fields: [
    { key: 'name', type: 'text', title: '姓名', required: true, placeholder: '请输入姓名' },
    { key: 'age', type: 'number', title: '年龄', required: true },
    { key: 'birthday', type: 'date', title: '出生日期' },
    {
      key: 'department',
      type: 'select',
      title: '部门',
      placeholder: '请选择部门',
      props: {
        options: [
          { label: '技术部', value: 'tech' },
          { label: '产品部', value: 'product' },
          { label: '设计部', value: 'design' },
          { label: '市场部', value: 'marketing' },
        ],
      },
    },
    { key: 'email', type: 'text', title: '邮箱', placeholder: '请输入邮箱' },
    { key: 'active', type: 'switch', title: '在职状态', defaultValue: true },
  ],
};

export const Basic: Story = {
  args: {
    schema: basicSchema,
    mode: 'create',
    onSubmit: fn(),
  },
};

/** 所有字段类型展示 */
const allTypesSchema: FormSchema = {
  title: '字段类型展示',
  description: '包含所有支持的字段类型',
  fields: [
    { key: 'text_field', type: 'text', title: '文本输入', placeholder: '请输入文本' },
    { key: 'textarea_field', type: 'textarea', title: '多行文本', placeholder: '请输入多行文本' },
    { key: 'number_field', type: 'number', title: '数字输入', placeholder: '请输入数字' },
    { key: 'date_field', type: 'date', title: '日期选择' },
    { key: 'datetime_field', type: 'datetime', title: '日期时间' },
    {
      key: 'select_field',
      type: 'select',
      title: '下拉选择',
      props: {
        options: [
          { label: '选项一', value: '1' },
          { label: '选项二', value: '2' },
          { label: '选项三', value: '3' },
        ],
      },
    },
    {
      key: 'radio_field',
      type: 'radio',
      title: '单选组',
      props: {
        options: [
          { label: '男', value: 'male' },
          { label: '女', value: 'female' },
        ],
      },
    },
    {
      key: 'checkbox_field',
      type: 'checkbox',
      title: '多选组',
      props: {
        options: [
          { label: '阅读', value: 'reading' },
          { label: '运动', value: 'sports' },
          { label: '音乐', value: 'music' },
        ],
      },
    },
    { key: 'switch_field', type: 'switch', title: '开关', defaultValue: false },
  ],
};

export const AllFieldTypes: Story = {
  args: {
    schema: allTypesSchema,
    mode: 'create',
    onSubmit: fn(),
  },
};

/** 带校验规则的表单 */
const validationSchema: FormSchema = {
  title: '用户注册',
  description: '带前端校验规则的表单',
  fields: [
    {
      key: 'username',
      type: 'text',
      title: '用户名',
      required: true,
      placeholder: '请输入用户名',
      rules: [
        { required: true, message: '用户名不能为空' },
        { minLength: 3, message: '用户名至少3个字符' },
        { maxLength: 20, message: '用户名最多20个字符' },
      ],
    },
    {
      key: 'email',
      type: 'text',
      title: '邮箱',
      required: true,
      placeholder: '请输入邮箱',
      rules: [
        { required: true, message: '邮箱不能为空' },
        { pattern: '^[\\w.-]+@[\\w.-]+\\.\\w+$', message: '请输入有效的邮箱地址' },
      ],
    },
    {
      key: 'age',
      type: 'number',
      title: '年龄',
      required: true,
      rules: [
        { required: true, message: '年龄不能为空' },
        { min: 1, message: '年龄必须大于0' },
        { max: 150, message: '请输入有效年龄' },
      ],
    },
  ],
};

export const WithValidation: Story = {
  args: {
    schema: validationSchema,
    mode: 'create',
    onSubmit: fn(),
  },
};

/** 分组表单 */
const groupedSchema: FormSchema = {
  title: '供应商注册',
  description: '请填写供应商的完整信息',
  groups: [
    {
      title: '基本信息',
      description: '供应商的基本信息',
      fields: [
        { key: 'company_name', type: 'text', title: '企业名称', required: true },
        { key: 'credit_code', type: 'text', title: '统一社会信用代码', required: true },
        {
          key: 'type',
          type: 'select',
          title: '供应商类型',
          props: {
            options: [
              { label: '原材料供应商', value: 'raw_material' },
              { label: '零部件供应商', value: 'parts' },
              { label: '服务供应商', value: 'service' },
            ],
          },
        },
      ],
    },
    {
      title: '联系方式',
      description: '联系人及地址信息',
      fields: [
        { key: 'contact_name', type: 'text', title: '联系人', required: true },
        { key: 'contact_phone', type: 'text', title: '联系电话', required: true },
        { key: 'address', type: 'textarea', title: '地址', placeholder: '请输入详细地址' },
      ],
    },
  ],
  fields: [
    { key: 'remark', type: 'textarea', title: '备注', placeholder: '其他需要说明的信息' },
  ],
};

export const GroupedForm: Story = {
  args: {
    schema: groupedSchema,
    mode: 'create',
    onSubmit: fn(),
  },
};

/** 编辑模式：带初始值填充 */
const editInitialValues = {
  name: '张三',
  age: 28,
  birthday: '1998-05-15',
  department: 'tech',
  email: 'zhangsan@example.com',
  active: true,
};

export const EditMode: Story = {
  args: {
    schema: basicSchema,
    mode: 'edit',
    initialValues: editInitialValues,
    onSubmit: fn(),
  },
};

/** 查看模式：所有字段只读 */
export const ViewMode: Story = {
  args: {
    schema: basicSchema,
    mode: 'view',
    initialValues: editInitialValues,
  },
};

/** 带草稿保存功能的表单 */
export const WithDraftSave: Story = {
  args: {
    schema: basicSchema,
    mode: 'create',
    onSubmit: fn(),
    onSaveDraft: fn(),
  },
};

/** 引用选择器 */
const referenceSchema: FormSchema = {
  title: '订单创建',
  description: '通过 fetchReferences 加载引用数据',
  fields: [
    { key: 'order_no', type: 'text', title: '订单编号', required: true },
    {
      key: 'customer_id',
      type: 'reference_selector',
      title: '客户',
      required: true,
      placeholder: '请选择客户',
      props: { referenceType: 'customer', displayField: 'name' },
    },
    {
      key: 'product_ids',
      type: 'reference_selector',
      title: '产品（多选）',
      placeholder: '请选择产品',
      props: { referenceType: 'product', displayField: 'name', multiple: true },
    },
  ],
};

export const ReferenceSelectorDemo: Story = {
  args: {
    schema: referenceSchema,
    mode: 'create',
    onSubmit: fn(),
    fetchReferences: async (type: string, keyword?: string) => {
      await new Promise((r) => setTimeout(r, 500));
      if (type === 'customer') {
        const customers = [
          { label: '腾讯科技', value: 'c1' },
          { label: '阿里巴巴', value: 'c2' },
          { label: '字节跳动', value: 'c3' },
          { label: '华为技术', value: 'c4' },
        ];
        return keyword ? customers.filter((c) => c.label.includes(keyword)) : customers;
      }
      const products = [
        { label: 'ERP 企业版', value: 'p1' },
        { label: 'ERP 标准版', value: 'p2' },
        { label: 'CRM 模块', value: 'p3' },
        { label: 'HR 模块', value: 'p4' },
      ];
      return keyword ? products.filter((p) => p.label.includes(keyword)) : products;
    },
  },
};

/** 子表（可编辑表格） */
const subTableSchema: FormSchema = {
  title: '采购订单',
  description: '包含明细子表',
  fields: [
    { key: 'order_no', type: 'text', title: '采购单号', required: true },
    {
      key: 'supplier', type: 'select', title: '供应商', required: true,
      props: {
        options: [
          { label: '供应商A', value: 's1' },
          { label: '供应商B', value: 's2' },
        ],
      },
    },
    {
      key: 'items',
      type: 'sub_table',
      title: '采购明细',
      required: true,
      props: {
        columns: [
          { key: 'product_name', title: '产品名称', type: 'text', width: 150, required: true },
          { key: 'quantity', title: '数量', type: 'number', width: 100, required: true },
          { key: 'unit_price', title: '单价', type: 'number', width: 120, required: true },
          { key: 'delivery_date', title: '预计到货', type: 'date', width: 150 },
        ],
        minRows: 1,
        maxRows: 20,
      },
    },
  ],
};

export const SubTableDemo: Story = {
  args: {
    schema: subTableSchema,
    mode: 'create',
    onSubmit: fn(),
  },
};

/** 隐藏提交按钮 */
export const NoSubmitButton: Story = {
  args: {
    schema: basicSchema,
    mode: 'create',
    showSubmit: false,
  },
};

/** 自定义提交按钮文字 */
export const CustomSubmitText: Story = {
  args: {
    schema: basicSchema,
    mode: 'create',
    submitText: '保存并提交',
    onSubmit: fn(),
  },
};

/** 无效 Schema 的错误提示 */
export const InvalidSchema: Story = {
  args: {
    // @ts-expect-error testing invalid schema
    schema: { title: '无效表单' },
  },
};
