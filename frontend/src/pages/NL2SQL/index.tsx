import { useState, useCallback } from 'react'
import ChatInterface from './ChatInterface'
import QueryHistory from './QueryHistory'
import { sendQuery } from '../../services/api'
import type { QueryMessage } from '../../models/nl2sql'

let msgId = 0
function nextId() { return `msg-${++msgId}-${Date.now()}` }

export default function NL2SQLPage() {
  const [messages, setMessages] = useState<QueryMessage[]>([])
  const [loading, setLoading] = useState(false)
  const [refreshKey, setRefreshKey] = useState(0)

  const handleSend = useCallback(async (text: string) => {
    const userMsg: QueryMessage = {
      id: nextId(),
      role: 'user',
      content: text,
      timestamp: new Date().toISOString(),
    }
    setMessages(prev => [...prev, userMsg])
    setLoading(true)

    try {
      const response = await sendQuery({ query: text })
      const assistantMsg: QueryMessage = {
        id: nextId(),
        role: 'assistant',
        content: response.explanation,
        sql: response.sql,
        result: response.result,
        timestamp: new Date().toISOString(),
      }
      setMessages(prev => [...prev, assistantMsg])
      setRefreshKey(k => k + 1)
    } catch (err) {
      const errorMsg: QueryMessage = {
        id: nextId(),
        role: 'assistant',
        content: `查询失败: ${err instanceof Error ? err.message : '未知错误'}`,
        timestamp: new Date().toISOString(),
      }
      setMessages(prev => [...prev, errorMsg])
    } finally {
      setLoading(false)
    }
  }, [])

  return (
    <div style={styles.page}>
      <QueryHistory onSelect={handleSend} refreshKey={refreshKey} />
      <div style={styles.chatArea}>
        <ChatInterface
          messages={messages}
          onSend={handleSend}
          loading={loading}
        />
      </div>
    </div>
  )
}

const styles: Record<string, React.CSSProperties> = {
  page: {
    display: 'flex',
    height: '100%',
  },
  chatArea: {
    flex: 1,
    minWidth: 0,
  },
}
