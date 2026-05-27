import { useState, useCallback, useRef } from 'react';
import type { FormStatus } from '../components/DynamicForm/types';

interface UseFormStateOptions {
  onSaveDraft?: (values: Record<string, unknown>) => Promise<void> | void;
  onSubmit?: (values: Record<string, unknown>) => Promise<void> | void;
}

interface UseFormStateReturn {
  status: FormStatus;
  error: string | null;
  submit: (values: Record<string, unknown>) => Promise<void>;
  saveDraft: (values: Record<string, unknown>) => Promise<void>;
  reset: () => void;
}

export function useFormState({ onSaveDraft, onSubmit }: UseFormStateOptions): UseFormStateReturn {
  const [status, setStatus] = useState<FormStatus>('draft');
  const [error, setError] = useState<string | null>(null);
  const statusRef = useRef<FormStatus>('draft');

  const saveDraft = useCallback(
    async (values: Record<string, unknown>) => {
      try {
        setStatus('draft');
        statusRef.current = 'draft';
        await onSaveDraft?.(values);
      } catch (e) {
        setError(e instanceof Error ? e.message : '保存草稿失败');
      }
    },
    [onSaveDraft]
  );

  const submit = useCallback(
    async (values: Record<string, unknown>) => {
      try {
        setStatus('submitting');
        statusRef.current = 'submitting';
        setError(null);
        await onSubmit?.(values);
        setStatus('submitted');
        statusRef.current = 'submitted';
      } catch (e) {
        setStatus('draft');
        statusRef.current = 'draft';
        setError(e instanceof Error ? e.message : '提交失败');
        throw e;
      }
    },
    [onSubmit]
  );

  const reset = useCallback(() => {
    setStatus('draft');
    statusRef.current = 'draft';
    setError(null);
  }, []);

  return { status, error, submit, saveDraft, reset };
}
