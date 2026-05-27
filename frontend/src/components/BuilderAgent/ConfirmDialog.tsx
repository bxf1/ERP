import React from 'react';
import { Modal, Typography, Descriptions, Tag, Space } from 'antd';
import type { BuilderPlan } from '@/types';

const { Text } = Typography;

interface Props {
  open: boolean;
  plan: BuilderPlan | null;
  confirming: boolean;
  onCancel: () => void;
  onConfirm: () => void;
}

const ConfirmDialog: React.FC<Props> = ({
  open,
  plan,
  confirming,
  onCancel,
  onConfirm,
}) => {
  if (!plan) return null;

  return (
    <Modal
      title="确认创建模块"
      open={open}
      onOk={onConfirm}
      onCancel={onCancel}
      confirmLoading={confirming}
      okText="确认创建"
      cancelText="取消"
      width={640}
      destroyOnClose
    >
      <Text type="secondary">
        请确认以下模块配置，确认后将自动创建表单、审批流程和菜单。
      </Text>

      <Descriptions
        column={1}
        size="small"
        style={{ marginTop: 16 }}
        bordered
      >
        <Descriptions.Item label="模块名称">
          <Text strong>{plan.moduleName}</Text>
        </Descriptions.Item>
        <Descriptions.Item label="描述">
          {plan.moduleDescription}
        </Descriptions.Item>
        <Descriptions.Item label="表单字段">
          <Space wrap size={[4, 4]}>
            {plan.formFields
              .filter((f) => f.visible)
              .map((f) => (
                <Tag key={f.key} color="blue">
                  {f.label}
                  {f.required && ' *'}
                </Tag>
              ))}
          </Space>
        </Descriptions.Item>
        <Descriptions.Item label="审批流程">
          <Space wrap size={[4, 4]}>
            {plan.approvalFlow.nodes.map((n) => (
              <Tag key={n.id}>{n.label}</Tag>
            ))}
          </Space>
        </Descriptions.Item>
        <Descriptions.Item label="菜单位置">
          {plan.menuPosition.parentMenu} &gt; {plan.menuPosition.menuName}
        </Descriptions.Item>
        {plan.remark && (
          <Descriptions.Item label="备注">{plan.remark}</Descriptions.Item>
        )}
      </Descriptions>
    </Modal>
  );
};

export default ConfirmDialog;
