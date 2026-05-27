import { useState } from 'react';
import type { ChatMessage as ChatMessageType } from '../models/query';
import { ResultVisualizer } from './ResultVisualizer';
import { ExportButton } from './ExportButton';

interface Props {
  message: ChatMessageType;
}

export function ChatMessage({ message }: Props) {
  return (
    <div className={`chat-message ${message.role}`}>
      <div className="message-avatar">
        {message.role === 'user' ? '👤' : '🤖'}
      </div>
      <div className="message-body">
        {message.segments.map((seg, i) => (
          <MessageSegment key={i} segment={seg} message={message} />
        ))}
      </div>
    </div>
  );
}

function MessageSegment({
  segment,
}: {
  segment: ChatMessageType['segments'][number];
  message: ChatMessageType;
}) {
  const [expanded, setExpanded] = useState(segment.type !== 'sql');

  switch (segment.type) {
    case 'text':
      return <p className="segment-text">{segment.content}</p>;

    case 'sql': {
      if (segment.type !== 'sql') return null;
      return (
        <div className="segment-sql">
          <div
            className="sql-header"
            onClick={() => setExpanded(!expanded)}
          >
            <span>🔍 生成的 SQL</span>
            <span className="sql-toggle">{expanded ? '收起' : '展开'}</span>
          </div>
          {expanded && (
            <pre className="sql-code"><code>{segment.content}</code></pre>
          )}
        </div>
      );
    }

    case 'result': {
      if (segment.type !== 'result') return null;
      return (
        <div className="segment-result">
          <div className="result-header">
            <span>📊 查询结果 ({segment.data.rows.length} 行)</span>
            <ExportButton data={segment.data} />
          </div>
          <ResultVisualizer data={segment.data} />
        </div>
      );
    }

    default:
      return null;
  }
}
