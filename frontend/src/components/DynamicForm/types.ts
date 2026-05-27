import type { ReactNode } from 'react';

export type FormMode = 'create' | 'edit' | 'view';

export type FieldType =
  | 'text'
  | 'number'
  | 'date'
  | 'datetime'
  | 'select'
  | 'reference_selector'
  | 'sub_table'
  | 'textarea'
  | 'switch'
  | 'radio'
  | 'checkbox'
  | 'custom';

export type FormStatus = 'draft' | 'submitting' | 'submitted';

export interface FieldOption {
  label: string;
  value: string | number;
  disabled?: boolean;
}

export interface ValidationRule {
  required?: boolean;
  message?: string;
  pattern?: string;
  min?: number;
  max?: number;
  minLength?: number;
  maxLength?: number;
  validator?: string;
}

export interface ReferenceSelectorConfig {
  referenceType: string;
  displayField: string;
  searchFields?: string[];
  filters?: Record<string, unknown>;
  multiple?: boolean;
  placeholder?: string;
}

export interface SubTableColumn {
  key: string;
  title: string;
  type: FieldType;
  width?: number;
  required?: boolean;
  rules?: ValidationRule[];
  props?: Record<string, unknown>;
}

export interface FieldSchema {
  key: string;
  type: FieldType;
  title: string;
  description?: string;
  required?: boolean;
  defaultValue?: unknown;
  placeholder?: string;
  disabled?: boolean;
  hidden?: boolean;
  rules?: ValidationRule[];
  props?: ReferenceSelectorConfig | { columns: SubTableColumn[]; maxRows?: number; minRows?: number } | Record<string, unknown>;
  render?: (props: { value: unknown; onChange: (v: unknown) => void; disabled: boolean }) => ReactNode;
}

export interface FormGroup {
  title: string;
  description?: string;
  fields: FieldSchema[];
  collapsible?: boolean;
  defaultCollapsed?: boolean;
}

export interface FormSchema {
  title?: string;
  description?: string;
  groups?: FormGroup[];
  fields: FieldSchema[];
}

export interface DynamicFormProps {
  schema: FormSchema;
  mode?: FormMode;
  initialValues?: Record<string, unknown>;
  onSubmit?: (values: Record<string, unknown>) => Promise<void> | void;
  onSaveDraft?: (values: Record<string, unknown>) => Promise<void> | void;
  onValuesChange?: (values: Record<string, unknown>) => void;
  onStatusChange?: (status: FormStatus) => void;
  showSubmit?: boolean;
  submitText?: string;
  extraActions?: ReactNode;
  fetchReferences?: (referenceType: string, keyword?: string) => Promise<FieldOption[]>;
  className?: string;
  style?: React.CSSProperties;
}
