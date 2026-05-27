import { useState, useRef, useEffect } from 'react'
import type { QueryMessage } from '../../models/nl2sql'
import ResultVisualizer from './ResultVisualizer'

interface Props {
  messages: QueryMessage[]
  onSend: (text: string) => void
  loading: boolean
}

export default function ChatInterface({ messages, onSend, loading }: Props) {
  const [input, setInput] = useState('')
  const bottomRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLTextAreaElement>(null)

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  useEffect(() => {
    if (!loading) inputRef.current?.focus()
  }, [loading])

  const handleSubmit = () => {
    const trimmed = input.trim()
    if (!trimmed || loading) return
    onSend(trimmed)
    setInput('')
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSubmit()
    }
  }

  return (
    <div style={styles.container}>
      <div style={styles.messageList}>
        {messages.length === 0 && (
          <div style={styles.empty}>
            <div style={styles.emptyIcon}>💬</div>
            <div style={styles.emptyTitle}>智能数据查询</div>
            <div style={styles.emptyHint}>
              用自然语言描述你想查询的数据，AI 将自动生成 SQL 并展示结果
            </div>
            <div style={styles.suggestions}>
              {['查询所有客户信息', '统计本月销售额', '按地区分组显示订单数量', '找出库存不足的商品'].map(s => (
                <button key={s} style={styles.chip} onClick={() => onSend(s)}>
                  {s}
                </button>
              ))}
            </div>
          </div>
        )}
        {messages.map(msg => (
          <div key={msg.id} style={{
            ...styles.message,
            flexDirection: msg.role === 'user' ? 'row-reverse' : 'row',
          }}>
            <div style={{
              ...styles.avatar,
              background: msg.role === 'user' ? 'var(--accent)' : 'var(--bg-tertiary)',
            }}>
              {msg.role === 'user' ? 'U' : 'AI'}
            </div>
            <div style={{ maxWidth: '75%' }}>
              <div style={{
                ...styles.bubble,
                background: msg.role === 'user' ? 'var(--accent)' : 'var(--bg-tertiary)',
                borderRadius: msg.role === 'user'
                  ? '16px 4px 16px 16px'
                  : '4px 16px 16px 16px',
              }}>
                <div style={styles.bubbleText}>{msg.content}</div>
                {msg.sql && (
                  <div style={styles.sqlBlock}>
                    <div style={styles.sqlLabel}>SQL</div>
                    <pre style={styles.sqlCode}>{msg.sql}</pre>
                  </div>
                )}
              </div>
              {msg.result && (
                <div style={styles.resultWrapper}>
                  <ResultVisualizer result={msg.result} />
                </div>
              )}
            </div>
          </div>
        ))}
        {loading && (
          <div style={{ ...styles.message, flexDirection: 'row' }}>
            <div style={{ ...styles.avatar, background: 'var(--bg-tertiary)' }}>AI</div>
            <div style={{ ...styles.bubble, background: 'var(--bg-tertiary)', borderRadius: '4px 16px 16px 16px' }}>
              <div style={styles.typing}>
                <span style={styles.dot} />
                <span style={styles.dot} />
                <span style={styles.dot} />
              </div>
            </div>
          </div>
        )}
        <div ref={bottomRef} />
      </div>
      <div style={styles.inputArea}>
        <textarea
          ref={inputRef}
          style={styles.input}
          value={input}
          onChange={e => setInput(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="输入自然语言查询，例如：查询上个月的销售总额..."
          rows={1}
          disabled={loading}
        />
        <button
          style={{
            ...styles.sendBtn,
            opacity: input.trim() && !loading ? 1 : 0.4,
          }}
          onClick={handleSubmit}
          disabled={!input.trim() || loading}
        >
          ↑
        </button>
      </div>
    </div>
  )
}

const styles: Record<string, React.CSSProperties> = {
  container: {
    display: 'flex',
    flexDirection: 'column',
    height: '100%',
  },
  messageList: {
    flex: 1,
    overflowY: 'auto',
    padding: '20px',
    display: 'flex',
    flexDirection: 'column',
    gap: '20px',
  },
  empty: {
    display: 'flex',
    flexDirection: 'column',
    alignItems: 'center',
    justifyContent: 'center',
    height: '100%',
    gap: '12px',
  },
  emptyIcon: { fontSize: '48px' },
  emptyTitle: { fontSize: '22px', fontWeight: 600, color: 'var(--text-primary)' },
  emptyHint: { fontSize: '14px', color: 'var(--text-secondary)', textAlign: 'center', maxWidth: 400 },
  suggestions: { display: 'flex', gap: '8px', flexWrap: 'wrap', justifyContent: 'center', marginTop: 12 },
  chip: {
    padding: '8px 16px',
    borderRadius: '20px',
    border: '1px solid var(--border)',
    background: 'var(--bg-tertiary)',
    color: 'var(--text-secondary)',
    fontSize: '13px',
    cursor: 'pointer',
    transition: 'all 0.15s',
  },
  message: {
    display: 'flex',
    gap: '10px',
    alignItems: 'flex-start',
  },
  avatar: {
    width: 32,
    height: 32,
    borderRadius: '50%',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    fontSize: 12,
    fontWeight: 700,
    flexShrink: 0,
    color: 'var(--text-primary)',
  },
  bubble: {
    padding: '12px 16px',
    color: 'var(--text-primary)',
    fontSize: 14,
    lineHeight: 1.6,
  },
  bubbleText: { whiteSpace: 'pre-wrap' },
  sqlBlock: {
    marginTop: 10,
    background: 'rgba(0,0,0,0.3)',
    borderRadius: 'var(--radius-sm)',
    overflow: 'hidden',
  },
  sqlLabel: {
    fontSize: 11,
    fontWeight: 600,
    color: 'var(--text-muted)',
    padding: '6px 12px 0',
    textTransform: 'uppercase',
    letterSpacing: '0.5px',
  },
  sqlCode: {
    margin: 0,
    padding: '8px 12px 10px',
    fontSize: 13,
    fontFamily: "'JetBrains Mono', 'Fira Code', monospace",
    color: 'var(--success)',
    whiteSpace: 'pre-wrap',
    wordBreak: 'break-all',
  },
  resultWrapper: { marginTop: 12 },
  typing: {
    display: 'flex',
    gap: 5,
    padding: '4px 0',
  },
  dot: {
    width: 7,
    height: 7,
    borderRadius: '50%',
    background: 'var(--text-muted)',
    display: 'inline-block',
  },
  inputArea: {
    padding: '16px 20px',
    borderTop: '1px solid var(--border)',
    display: 'flex',
    gap: '10px',
    alignItems: 'flex-end',
  },
  input: {
    flex: 1,
    padding: '12px 16px',
    borderRadius: 'var(--radius)',
    border: '1px solid var(--border)',
    background: 'var(--bg-tertiary)',
    color: 'var(--text-primary)',
    fontSize: 14,
    resize: 'none',
    outline: 'none',
    maxHeight: 120,
  },
  sendBtn: {
    width: 40,
    height: 40,
    borderRadius: '50%',
    border: 'none',
    background: 'var(--accent)',
    color: '#fff',
    fontSize: 18,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    flexShrink: 0,
    transition: 'opacity 0.15s',
  },
}
