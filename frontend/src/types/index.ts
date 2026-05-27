// ============================================================
// Builder Agent Type Definitions
// ============================================================

// ---- Message Types ----

export type MessageRole = 'user' | 'assistant' | 'system';

export interface ChatMessage {
  id: string;
  role: MessageRole;
  content: string;
  timestamp: number;
  plan?: BuilderPlan;
}

// ---- Builder Plan ----

export interface BuilderPlan {
  id: string;
  status: 'draft' | 'confirmed' | 'creating' | 'done' | 'failed';
  moduleName: string;
  moduleDescription: string;
  formFields: FormFieldConfig[];
  approvalFlow: ApprovalFlowConfig;
  menuPosition: MenuPositionConfig;
  remark: string;
  createdAt: number;
}

// ---- Form Field Configuration ----

export type FieldType =
  | 'string'
  | 'number'
  | 'boolean'
  | 'date'
  | 'datetime'
  | 'select'
  | 'multiSelect'
  | 'textarea'
  | 'file'
  | 'image'
  | 'richtext'
  | 'user'
  | 'department';

export interface FormFieldConfig {
  key: string;
  label: string;
  type: FieldType;
  required: boolean;
  placeholder?: string;
  defaultValue?: unknown;
  options?: { label: string; value: string }[];
  validationRules?: ValidationRule[];
  order: number;
  visible: boolean;
  span: 24 | 12 | 8 | 6;
}

export interface ValidationRule {
  type: 'required' | 'minLength' | 'maxLength' | 'pattern' | 'min' | 'max';
  value?: unknown;
  message: string;
}

// ---- Approval Flow Configuration ----

export interface ApprovalFlowConfig {
  name: string;
  nodes: ApprovalNode[];
  edges: ApprovalEdge[];
}

export type ApprovalNodeType = 'start' | 'approval' | 'condition' | 'end';

export interface ApprovalNode {
  id: string;
  type: ApprovalNodeType;
  label: string;
  assignee?: string;
  conditions?: ApprovalCondition[];
}

export interface ApprovalEdge {
  id: string;
  source: string;
  target: string;
  label?: string;
}

export interface ApprovalCondition {
  field: string;
  operator: 'eq' | 'neq' | 'gt' | 'lt' | 'gte' | 'lte' | 'in';
  value: unknown;
  targetNodeId: string;
}

// ---- Menu Position Configuration ----

export type MenuType = 'directory' | 'page' | 'button';

export interface MenuPositionConfig {
  parentMenu: string;
  menuName: string;
  menuType: MenuType;
  icon?: string;
  order: number;
  routePath?: string;
}

// ---- Creation Progress & Feedback ----

export type CreationStep =
  | 'validating'
  | 'creating_model'
  | 'creating_form'
  | 'creating_flow'
  | 'creating_menu'
  | 'setting_permissions'
  | 'done';

export interface CreationProgress {
  currentStep: CreationStep;
  steps: CreationStepInfo[];
  startedAt: number;
}

export interface CreationStepInfo {
  key: CreationStep;
  label: string;
  status: 'waiting' | 'running' | 'done' | 'failed';
  message?: string;
}

export interface CreationResult {
  success: boolean;
  moduleId?: string;
  moduleName?: string;
  errors?: string[];
  links?: {
    formPage?: string;
    flowPage?: string;
    menuEntry?: string;
  };
}

// ---- API Types ----

export interface CreateModuleRequest {
  plan: BuilderPlan;
}

export interface CreateModuleResponse {
  success: boolean;
  moduleId: string;
  message: string;
}

export interface PlanPreviewRequest {
  moduleDescription: string;
  remark: string;
}

export interface PlanPreviewResponse {
  plan: BuilderPlan;
}

export interface AdjustPlanRequest {
  planId: string;
  adjustment: string;
}

export interface AdjustPlanResponse {
  plan: BuilderPlan;
  changes: PlanChange[];
}

export interface PlanChange {
  field: string;
  before: unknown;
  after: unknown;
}
