import { useState, useCallback, useRef } from 'react';
import type {
  ChatMessage,
  BuilderPlan,
  CreationStepInfo,
  CreationStep,
  CreationResult,
} from '@/types';
import { mockPlan, mockMessages } from '@/mock/data';

const CREATION_STEPS: CreationStepInfo[] = [
  { key: 'validating', label: '校验配置', status: 'waiting' },
  { key: 'creating_model', label: '创建数据模型', status: 'waiting' },
  { key: 'creating_form', label: '创建动态表单', status: 'waiting' },
  { key: 'creating_flow', label: '创建工作流', status: 'waiting' },
  { key: 'creating_menu', label: '创建菜单项', status: 'waiting' },
  { key: 'setting_permissions', label: '配置权限', status: 'waiting' },
];

function generateId(): string {
  return `${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;
}

export function useBuilderChat() {
  const [messages, setMessages] = useState<ChatMessage[]>(mockMessages);
  const [loading, setLoading] = useState(false);
  const [plan, setPlan] = useState<BuilderPlan | null>(mockPlan);
  const [confirmOpen, setConfirmOpen] = useState(false);
  const [confirming, setConfirming] = useState(false);
  const [progressOpen, setProgressOpen] = useState(false);
  const [progressSteps, setProgressSteps] = useState<CreationStepInfo[] | null>(null);
  const [result, setResult] = useState<CreationResult | null>(null);
  const loadingRef = useRef(false);

  const sendMessage = useCallback(
    async (text: string) => {
      if (loadingRef.current) return;
      loadingRef.current = true;
      setLoading(true);

      const userMsg: ChatMessage = {
        id: generateId(),
        role: 'user',
        content: text,
        timestamp: Date.now(),
      };
      setMessages((prev) => [...prev, userMsg]);

      // Simulate AI response delay
      await new Promise((r) => setTimeout(r, 1500));

      const updatedPlan: BuilderPlan = {
        ...mockPlan,
        id: generateId(),
        status: 'draft',
        remark: text.includes('备注') ? text : '',
        createdAt: Date.now(),
      };

      const assistantMsg: ChatMessage = {
        id: generateId(),
        role: 'assistant',
        content: `好的，根据你的需求，我设计了以下方案：

**模块名称**：${updatedPlan.moduleName}
**表单字段**：${updatedPlan.formFields.filter((f) => f.visible).map((f) => f.label).join('、')}
**审批流程**：${updatedPlan.approvalFlow.nodes.map((n) => n.label).join(' → ')}
**菜单位置**：${updatedPlan.menuPosition.parentMenu} → ${updatedPlan.menuPosition.menuName}

请查看右侧预览。如需调整，请在调整需求框中描述。确认无误后点击"确认创建"。`,
        timestamp: Date.now(),
        plan: updatedPlan,
      };

      setMessages((prev) => [...prev, assistantMsg]);
      setPlan(updatedPlan);
      setLoading(false);
      loadingRef.current = false;
    },
    [],
  );

  const adjustPlan = useCallback(async (remark: string) => {
    setLoading(true);
    loadingRef.current = true;

    await new Promise((r) => setTimeout(r, 1500));

    const updatedPlan: BuilderPlan = {
      ...mockPlan,
      id: generateId(),
      status: 'draft',
      remark,
      formFields: [
        ...mockPlan.formFields,
        {
          key: `custom_${Date.now()}`,
          label: remark.includes('备注') ? '备注' : '新增字段',
          type: 'textarea',
          required: false,
          placeholder: remark,
          order: mockPlan.formFields.length + 1,
          visible: true,
          span: 24,
        },
      ],
      createdAt: Date.now(),
    };

    const assistantMsg: ChatMessage = {
      id: generateId(),
      role: 'assistant',
      content: `已根据你的调整重新生成方案，新增了字段。请查看右侧预览确认。`,
      timestamp: Date.now(),
      plan: updatedPlan,
    };

    setMessages((prev) => [...prev, assistantMsg]);
    setPlan(updatedPlan);
    setLoading(false);
    loadingRef.current = false;
  }, []);

  const openConfirm = useCallback(() => {
    if (plan) {
      setConfirmOpen(true);
    }
  }, [plan]);

  const confirmCreate = useCallback(async () => {
    setConfirming(true);
    await new Promise((r) => setTimeout(r, 800));
    setConfirmOpen(false);
    setConfirming(false);

    setPlan((prev) => (prev ? { ...prev, status: 'creating' } : null));

    // Start progress simulation
    setProgressOpen(true);
    const steps: CreationStepInfo[] = CREATION_STEPS.map((s) => ({ ...s }));
    setProgressSteps([...steps]);

    for (let i = 0; i < steps.length; i++) {
      steps[i] = { ...steps[i], status: 'running' };
      setProgressSteps([...steps]);
      await new Promise((r) => setTimeout(r, 800 + Math.random() * 600));

      steps[i] = { ...steps[i], status: 'done', message: '完成' };
      setProgressSteps([...steps]);
    }

    const creationResult: CreationResult = {
      success: true,
      moduleId: `MOD-${Date.now()}`,
      moduleName: plan?.moduleName || '新模块',
      links: {
        formPage: `/form/${plan?.moduleName || 'module'}`,
        flowPage: `/flow/${plan?.moduleName || 'module'}`,
        menuEntry: `/menu/${plan?.menuPosition.routePath || 'module'}`,
      },
    };

    setResult(creationResult);
    setPlan((prev) => (prev ? { ...prev, status: 'done' } : null));
  }, [plan]);

  const closeProgress = useCallback(() => {
    setProgressOpen(false);
    setProgressSteps(null);
    setResult(null);
  }, []);

  const regenerate = useCallback(async () => {
    setLoading(true);
    loadingRef.current = true;
    await new Promise((r) => setTimeout(r, 1000));

    const newPlan: BuilderPlan = {
      ...mockPlan,
      id: generateId(),
      status: 'draft',
      createdAt: Date.now(),
    };

    const assistantMsg: ChatMessage = {
      id: generateId(),
      role: 'assistant',
      content: '已重新生成方案，请查看右侧预览。',
      timestamp: Date.now(),
      plan: newPlan,
    };

    setMessages((prev) => [...prev, assistantMsg]);
    setPlan(newPlan);
    setLoading(false);
    loadingRef.current = false;
  }, []);

  return {
    messages,
    loading,
    plan,
    confirmOpen,
    confirming,
    progressOpen,
    progressSteps,
    result,
    sendMessage,
    adjustPlan,
    openConfirm,
    confirmCreate,
    closeProgress,
    regenerate,
    setConfirmOpen,
  };
}
