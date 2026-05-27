import type { FormSchema, FieldSchema } from './types';

export function validateSchema(schema: unknown): schema is FormSchema {
  if (!schema || typeof schema !== 'object') return false;
  const s = schema as Record<string, unknown>;
  if (!Array.isArray(s.fields)) return false;
  if (s.groups !== undefined && !Array.isArray(s.groups)) return false;
  return s.fields.every(isValidField) && (!s.groups || (s.groups as unknown[]).every(isValidGroup));
}

function isValidGroup(g: unknown): boolean {
  if (!g || typeof g !== 'object') return false;
  const group = g as Record<string, unknown>;
  return typeof group.title === 'string' && Array.isArray(group.fields);
}

function isValidField(f: unknown): f is FieldSchema {
  if (!f || typeof f !== 'object') return false;
  const field = f as Record<string, unknown>;
  return typeof field.key === 'string' && typeof field.type === 'string';
}

export function toFormilySchema(schema: FormSchema): Record<string, unknown> {
  const properties: Record<string, unknown> = {};

  for (const field of schema.fields) {
    properties[field.key] = fieldToSchema(field);
  }

  return {
    type: 'object',
    properties,
  };
}

export function fieldToSchema(field: FieldSchema): Record<string, unknown> {
  const base: Record<string, unknown> = {
    title: field.title,
    'x-decorator': 'FormItem',
  };

  if (field.description) {
    base.description = field.description;
  }

  if (field.required || field.rules?.some((r) => r.required)) {
    base.required = true;
  }

  const xValidator: Record<string, unknown> = {};
  for (const rule of field.rules || []) {
    if (rule.pattern) xValidator.pattern = rule.pattern;
    if (rule.min !== undefined) xValidator.minimum = rule.min;
    if (rule.max !== undefined) xValidator.maximum = rule.max;
    if (rule.minLength !== undefined) xValidator.minLength = rule.minLength;
    if (rule.maxLength !== undefined) xValidator.maxLength = rule.maxLength;
    if (rule.message) xValidator.message = rule.message;
  }
  if (Object.keys(xValidator).length > 0) {
    base['x-validator'] = xValidator;
  }

  if (field.props) {
    base['x-component-props'] = field.props;
  }

  if (field.defaultValue !== undefined) {
    base.default = field.defaultValue;
  }

  return base;
}
