import React, { useRef, useEffect } from 'react';
import { Avatar, Typography, Space } from 'antd';
import { UserOutlined, RobotOutlined } from '@ant-design/icons';
import type { ChatMessage } from '@/types';

const { Text, Paragraph } = Typography;

interface Props {
  messages: ChatMessage[];
  loading: boolean;
}

const MessageBubble: React.FC<{ message: ChatMessage }> = ({ message }) => {
  const isUser = message.role === 'user';
  return (
    <div
      style={{
        display: 'flex',
        flexDirection: isUser ? 'row-reverse' : 'row',
        gap: 12,
        marginBottom: 20,
        alignItems: 'flex-start',
      }}
    >
      <Avatar
        icon={isUser ? <UserOutlined /> : <RobotOutlined />}
        style={{
          backgroundColor: isUser ? '#1677ff' : '#52c41a',
          flexShrink: 0,
        }}
      />
      <div
        style={{
          maxWidth: '75%',
          background: isUser ? '#e6f4ff' : '#f6ffed',
          borderRadius: 12,
          padding: '12px 16px',
          border: `1px solid ${isUser ? '#91caff' : '#b7eb8f'}`,
        }}
      >
        <Paragraph style={{ margin: 0, whiteSpace: 'pre-wrap' }}>
          {message.content}
        </Paragraph>
        <Text
          type="secondary"
          style={{ fontSize: 11, display: 'block', marginTop: 6 }}
        >
          {new Date(message.timestamp).toLocaleTimeString()}
        </Text>
      </div>
    </div>
  );
};

const LoadingIndicator: React.FC = () => (
  <div style={{ display: 'flex', gap: 12, marginBottom: 20, alignItems: 'center' }}>
    <Avatar icon={<RobotOutlined />} style={{ backgroundColor: '#52c41a' }} />
    <div style={{ display: 'flex', gap: 6 }}>
      <span
        style={{
          width: 8,
          height: 8,
          borderRadius: '50%',
          background: '#52c41a',
          animation: 'pulse 1.4s ease-in-out infinite',
        }}
      />
      <span
        style={{
          width: 8,
          height: 8,
          borderRadius: '50%',
          background: '#52c41a',
          animation: 'pulse 1.4s ease-in-out 0.2s infinite',
        }}
      />
      <span
        style={{
          width: 8,
          height: 8,
          borderRadius: '50%',
          background: '#52c41a',
          animation: 'pulse 1.4s ease-in-out 0.4s infinite',
        }}
      />
    </div>
  </div>
);

const ChatMessageList: React.FC<Props> = ({ messages, loading }) => {
  const bottomRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages, loading]);

  return (
    <div
      style={{
        flex: 1,
        overflowY: 'auto',
        padding: '16px 20px',
        display: 'flex',
        flexDirection: 'column',
      }}
    >
      {messages.length === 0 && !loading && (
        <div
          style={{
            flex: 1,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
          }}
        >
          <Text type="secondary" style={{ fontSize: 14 }}>
            描述你想要创建的模块，AI 将为你设计方案
          </Text>
        </div>
      )}
      {messages.map((msg) => (
        <MessageBubble key={msg.id} message={msg} />
      ))}
      {loading && <LoadingIndicator />}
      <div ref={bottomRef} />
    </div>
  );
};

export default ChatMessageList;
