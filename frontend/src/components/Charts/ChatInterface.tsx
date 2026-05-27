import { useState, useRef, useEffect, useCallback } from 'react';
import type { ChatMessage, QueryResponse } from '../models/query';
import { queryNL2SQL } from '../services/nl2sql';
import { ChatMessage as ChatMessageCmp } from './ChatMessage';
import { ChatInput } from './ChatInput';
import { WelcomeScreen } from './WelcomeScreen';

let msgId = 0;
function nextId(): string {
  return `msg-${++msgId}-${Date.now()}`;
}

interface Props {
  onHistoryUpdate: () => void;
}

export function ChatInterface({ onHistoryUpdate }: Props) {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [loading, setLoading] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const scrollToBottom = useCallback(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, []);

  useEffect(() => {
    scrollToBottom();
  }, [messages, scrollToBottom]);

  const handleSend = async (question: string) => {
    const userMsg: ChatMessage = {
      id: nextId(),
      role: 'user',
      segments: [{ type: 'text', content: question }],
    };
    setMessages((prev) => [...prev, userMsg]);
    setLoading(true);

    try {
      const data: QueryResponse = await queryNL2SQL(question);

      const assistantMsg: ChatMessage = {
        id: nextId(),
        role: 'assistant',
        segments: [
          { type: 'text', content: `查询：${data.question}` },
          { type: 'sql', content: data.sql },
          ...(data.columns.length > 0
            ? [{ type: 'result' as const, data }]
            : []),
        ],
        data,
      };
      setMessages((prev) => [...prev, assistantMsg]);
      onHistoryUpdate();
    } catch (err: unknown) {
      const errorMsg = err instanceof Error ? err.message : '查询失败，请稍后重试';
      const assistantMsg: ChatMessage = {
        id: nextId(),
        role: 'assistant',
        segments: [{ type: 'text', content: `❌ ${errorMsg}` }],
      };
      setMessages((prev) => [...prev, assistantMsg]);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="chat-interface">
      <div className="chat-header">
        <h2>AI 智能查询</h2>
        <span className="chat-subtitle">用自然语言查询数据库</span>
      </div>

      <div className="chat-messages">
        {messages.length === 0 && <WelcomeScreen onSend={handleSend} />}
        {messages.map((msg) => (
          <ChatMessageCmp key={msg.id} message={msg} />
        ))}
        {loading && (
          <div className="chat-message assistant">
            <div className="message-avatar">🤖</div>
            <div className="message-body">
              <div className="typing-indicator">
                <span className="typing-dot" />
                <span className="typing-dot" />
                <span className="typing-dot" />
              </div>
            </div>
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>

      <ChatInput onSend={handleSend} disabled={loading} />
    </div>
  );
}
