import type { BuilderPlan, ChatMessage, ApprovalFlowConfig } from '@/types';

export const mockApprovalFlow: ApprovalFlowConfig = {
  name: '采购审批流程',
  nodes: [
    { id: 'start', type: 'start', label: '提交申请' },
    { id: 'manager', type: 'approval', label: '部门经理审批', assignee: '部门经理' },
    {
      id: 'amount_check',
      type: 'condition',
      label: '金额判断',
      conditions: [
        {
          field: 'amount',
          operator: 'gt',
          value: 10000,
          targetNodeId: 'director',
        },
      ],
    },
    { id: 'director', type: 'approval', label: '总监审批', assignee: '总监' },
    { id: 'end', type: 'end', label: '完成' },
  ],
  edges: [
    { id: 'e1', source: 'start', target: 'manager' },
    { id: 'e2', source: 'manager', target: 'amount_check' },
    { id: 'e3', source: 'amount_check', target: 'director', label: '金额 > 10000' },
    { id: 'e4', source: 'amount_check', target: 'end', label: '金额 <= 10000' },
    { id: 'e5', source: 'director', target: 'end' },
  ],
};

export const mockPlan: BuilderPlan = {
  id: 'plan-001',
  status: 'draft',
  moduleName: '采购订单管理',
  moduleDescription: '创建采购订单模块，包含供应商选择、商品明细、金额计算、审批流程',
  formFields: [
    {
      key: 'order_no',
      label: '订单编号',
      type: 'string',
      required: true,
      placeholder: '自动生成',
      order: 1,
      visible: true,
      span: 12,
    },
    {
      key: 'supplier',
      label: '供应商',
      type: 'select',
      required: true,
      placeholder: '请选择供应商',
      options: [
        { label: '供应商A', value: 'A' },
        { label: '供应商B', value: 'B' },
        { label: '供应商C', value: 'C' },
      ],
      order: 2,
      visible: true,
      span: 12,
    },
    {
      key: 'order_date',
      label: '订单日期',
      type: 'date',
      required: true,
      order: 3,
      visible: true,
      span: 12,
    },
    {
      key: 'amount',
      label: '订单金额',
      type: 'number',
      required: true,
      placeholder: '请输入金额',
      validationRules: [{ type: 'min', value: 0, message: '金额必须大于0' }],
      order: 4,
      visible: true,
      span: 12,
    },
    {
      key: 'remark',
      label: '备注',
      type: 'textarea',
      required: false,
      placeholder: '请输入备注信息',
      order: 5,
      visible: true,
      span: 24,
    },
  ],
  approvalFlow: mockApprovalFlow,
  menuPosition: {
    parentMenu: '采购管理',
    menuName: '采购订单',
    menuType: 'page',
    icon: 'ShoppingOutlined',
    order: 1,
    routePath: '/purchase/order',
  },
  remark: '',
  createdAt: Date.now(),
};

export const mockMessages: ChatMessage[] = [
  {
    id: 'msg-1',
    role: 'user',
    content: '我需要一个采购订单管理模块，包含供应商选择、商品明细、金额计算，需要审批流程',
    timestamp: Date.now() - 300000,
  },
  {
    id: 'msg-2',
    role: 'assistant',
    content:
      '好的，我为您设计了采购订单管理模块的方案，请查看右侧预览：\n\n**表单字段**：订单编号、供应商、订单日期、订单金额、备注\n**审批流程**：提交 → 部门经理审批 → 金额判断 → 总监审批（>10000元）→ 完成\n**菜单位置**：采购管理 → 采购订单\n\n如果需要调整，请直接告诉我。确认无误后点击"确认创建"即可。',
    timestamp: Date.now() - 280000,
    plan: mockPlan,
  },
];
