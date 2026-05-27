import React from 'react';
import { Typography, Input, Button, Space, Tag, Empty } from 'antd';
import { EditOutlined, ReloadOutlined } from '@ant-design/icons';
import FormPreview from './FormPreview';
import FlowPreview from './FlowPreview';
import MenuPreview from './MenuPreview';
import type { BuilderPlan } from '@/types';

const { Title, Text, Paragraph } = Typography;
const { TextArea } = Input;

interface Props {
  plan: BuilderPlan | null;
  adjusting: boolean;
  onAdjust: (remark: string) => void;
  onConfirm: () => void;
  onRegenerate: () => void;
}

const PlanPreview: React.FC<Props> = ({
  plan,
  adjusting,
  onAdjust,
  onConfirm,
  onRegenerate,
}) => {
  const [remark, setRemark] = React.useState(plan?.remark || '');
  const [editingRemark, setEditingRemark] = React.useState(false);

  React.useEffect(() => {
    setRemark(plan?.remark || '');
    setEditingRemark(false);
  }, [plan?.id]);

  if (!plan) {
    return (
      <div
        style={{
          height: '100%',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'center',
          padding: 40,
        }}
      >
        <Empty description="在左侧对话中描述你的需求，AI 生成的方案将在这里预览" />
      </div>
    );
  }

  const handleApplyRemark = () => {
    onAdjust(remark);
    setEditingRemark(false);
  };

  return (
    <div style={{ padding: 16, overflowY: 'auto', height: '100%' }}>
      <div
        style={{
          marginBottom: 16,
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'flex-start',
        }}
      >
        <div>
          <Title level={4} style={{ margin: 0 }}>
            {plan.moduleName}
          </Title>
          <Text type="secondary">{plan.moduleDescription}</Text>
        </div>
        <Tag
          color={
            plan.status === 'confirmed'
              ? 'green'
              : plan.status === 'creating'
                ? 'processing'
                : plan.status === 'done'
                  ? 'success'
                  : plan.status === 'failed'
                    ? 'error'
                    : 'blue'
          }
        >
          {plan.status === 'draft'
            ? '待确认'
            : plan.status === 'confirmed'
              ? '已确认'
              : plan.status === 'creating'
                ? '创建中'
                : plan.status === 'done'
                  ? '已创建'
                  : '失败'}
        </Tag>
      </div>

      <FormPreview fields={plan.formFields} moduleName={plan.moduleName} />

      <FlowPreview flowConfig={plan.approvalFlow} />

      <MenuPreview menu={plan.menuPosition} />

      {/* Remark adjustment section */}
      <div
        style={{
          background: '#fffbe6',
          border: '1px solid #ffe58f',
          borderRadius: 8,
          padding: '12px 16px',
          marginBottom: 16,
        }}
      >
        <Text strong style={{ display: 'block', marginBottom: 8 }}>
          <EditOutlined /> 调整需求
        </Text>
        {editingRemark ? (
          <>
            <TextArea
              value={remark}
              onChange={(e) => setRemark(e.target.value)}
              placeholder="描述你想要的调整，例如：再多加一个备注字段..."
              autoSize={{ minRows: 2, maxRows: 4 }}
              style={{ marginBottom: 8 }}
            />
            <Space>
              <Button
                type="primary"
                size="small"
                onClick={handleApplyRemark}
                loading={adjusting}
                icon={<ReloadOutlined />}
              >
                AI 重新生成
              </Button>
              <Button
                size="small"
                onClick={() => {
                  setEditingRemark(false);
                  setRemark(plan.remark);
                }}
              >
                取消
              </Button>
            </Space>
          </>
        ) : (
          <>
            <Paragraph
              style={{ margin: 0, color: '#666' }}
              onClick={() => setEditingRemark(true)}
            >
              {remark || '点击这里添加备注或调整需求……'}
            </Paragraph>
            <Button
              type="link"
              size="small"
              icon={<EditOutlined />}
              onClick={() => setEditingRemark(true)}
              style={{ padding: 0, marginTop: 4 }}
            >
              编辑调整
            </Button>
          </>
        )}
      </div>

      {/* Action buttons */}
      <Space style={{ width: '100%', justifyContent: 'flex-end' }} size={12}>
        <Button icon={<ReloadOutlined />} onClick={onRegenerate}>
          重新生成
        </Button>
        <Button
          type="primary"
          size="large"
          onClick={onConfirm}
          disabled={plan.status !== 'draft'}
        >
          确认创建
        </Button>
      </Space>
    </div>
  );
};

export default PlanPreview;
