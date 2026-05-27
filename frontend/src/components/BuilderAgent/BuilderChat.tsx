import React from 'react';
import { Card, Typography, Space, Tag } from 'antd';
import { ThunderboltOutlined } from '@ant-design/icons';
import ChatMessageList from './ChatMessageList';
import ChatInput from './ChatInput';
import type { ChatMessage } from '@/types';

const { Title, Text } = Typography;

interface Props {
  messages: ChatMessage[];
  loading: boolean;
  onSend: (text: string) => void;
}

const suggestedPrompts = [
  '创建一个采购订单管理模块',
  '创建员工请假审批流程',
  '创建一个客户信息管理表单',
];

const BuilderChat: React.FC<Props> = ({ messages, loading, onSend }) => {
  return (
    <Card
      style={{
        height: '100%',
        display: 'flex',
        flexDirection: 'column',
        borderRadius: 0,
      }}
      styles={{ body: { flex: 1, display: 'flex', flexDirection: 'column', padding: 0 } }}
    >
      <div
        style={{
          padding: '16px 20px',
          borderBottom: '1px solid #f0f0f0',
          background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
        }}
      >
        <Space>
          <ThunderboltOutlined style={{ color: '#fff', fontSize: 18 }} />
          <Title level={4} style={{ margin: 0, color: '#fff' }}>
            Builder Agent
          </Title>
          <Tag color="cyan" style={{ marginLeft: 8 }}>
            AI 辅助建模块
          </Tag>
        </Space>
        <Text style={{ color: 'rgba(255,255,255,0.75)', fontSize: 12, display: 'block', marginTop: 4 }}>
          通过自然语言描述需求，AI 自动设计模块方案
        </Text>
      </div>

      <ChatMessageList messages={messages} loading={loading} />

      {messages.length === 0 && (
        <div style={{ padding: '0 20px 12px' }}>
          <Text type="secondary" style={{ fontSize: 12, marginBottom: 8, display: 'block' }}>
            试试这样说：
          </Text>
          <Space wrap size={[8, 8]}>
            {suggestedPrompts.map((p) => (
              <Tag
                key={p}
                style={{ cursor: 'pointer', padding: '4px 10px' }}
                color="processing"
                onClick={() => !loading && onSend(p)}
              >
                {p}
              </Tag>
            ))}
          </Space>
        </div>
      )}

      <ChatInput onSend={onSend} disabled={loading} />
    </Card>
  );
};

export default BuilderChat;
