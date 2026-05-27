import { Select, Spin } from 'antd';
import { connect, mapProps } from '@formily/react';
import { useState, useEffect, useCallback } from 'react';
import type { FieldOption } from '../types';

interface ReferenceSelectorProps {
  referenceType?: string;
  displayField?: string;
  searchFields?: string[];
  filters?: Record<string, unknown>;
  multiple?: boolean;
  placeholder?: string;
  value?: unknown;
  onChange?: (value: unknown) => void;
  disabled?: boolean;
  options?: FieldOption[];
  /** External fetch function for loading reference data */
  onFetch?: (referenceType: string, keyword?: string) => Promise<FieldOption[]>;
}

function ReferenceSelectorInner(props: ReferenceSelectorProps) {
  const {
    multiple = false,
    placeholder = '请选择',
    value,
    onChange,
    disabled,
    referenceType,
    options: staticOptions,
    onFetch,
  } = props;

  const [dynamicOptions, setDynamicOptions] = useState<FieldOption[]>([]);
  const [loading, setLoading] = useState(false);

  const fetchOptions = useCallback(
    async (keyword?: string) => {
      if (!onFetch || !referenceType) return;
      setLoading(true);
      try {
        const result = await onFetch(referenceType, keyword);
        setDynamicOptions(result);
      } finally {
        setLoading(false);
      }
    },
    [onFetch, referenceType]
  );

  useEffect(() => {
    if (onFetch && referenceType) {
      fetchOptions();
    }
  }, [onFetch, referenceType, fetchOptions]);

  const options = staticOptions && staticOptions.length > 0 ? staticOptions : dynamicOptions;

  return (
    <Select
      mode={multiple ? 'multiple' : undefined}
      placeholder={placeholder}
      value={value as string | string[]}
      onChange={onChange}
      disabled={disabled}
      showSearch
      allowClear
      loading={loading}
      onSearch={onFetch ? (keyword) => fetchOptions(keyword) : undefined}
      filterOption={
        onFetch
          ? false
          : (input, option) => (option?.label as string)?.toLowerCase().includes(input.toLowerCase())
      }
      notFoundContent={loading ? <Spin size="small" /> : null}
      options={options.map((opt) => ({ label: opt.label, value: opt.value, disabled: opt.disabled }))}
    />
  );
}

export const ReferenceSelector = connect(
  ReferenceSelectorInner,
  mapProps({ dataSource: 'options' })
);

export default ReferenceSelector;
