import { useState, useEffect } from 'react'
import type { QueryHistoryItem } from '../../models/nl2sql'
import { fetchHistory } from '../../services/api'

interface Props {
  onSelect: (query: string) => void
  refreshKey: number
}

export default function QueryHistory({ onSelect, refreshKey }: Props) {
  const [history, setHistory] = useState<QueryHistoryItem[]>([])
  const [search, setSearch] = useState('')
  const [collapsed, setCollapsed] = useState(false)

  useEffect(() => {
    fetchHistory()
      .then(setHistory)
      .catch(() => setHistory([]))
  }, [refreshKey])

  const filtered = history.filter(h =>
    h.query.toLowerCase().includes(search.toLowerCase())
  )

  if (collapsed) {
    return (
      <div style={styles.collapsedBar}>
        <button style={styles.toggleBtn} onClick={() => setCollapsed(false)} title="展开历史记录">
          ☰
        </button>
      </div>
    )
  }

  return (
    <div style={styles.sidebar}>
      <div style={styles.header}>
        <span style={styles.title}>历史记录</span>
        <button style={styles.toggleBtn} onClick={() => setCollapsed(true)}>✕</button>
      </div>
      <input
        style={styles.search}
        placeholder="搜索历史..."
        value={search}
        onChange={e => setSearch(e.target.value)}
      />
      <div style={styles.list}>
        {filtered.length === 0 && (
          <div style={styles.empty}>
            {history.length === 0 ? '暂无查询记录' : '无匹配结果'}
          </div>
        )}
        {filtered.map(item => (
          <div
            key={item.id}
            style={styles.item}
            onClick={() => onSelect(item.query)}
          >
            <div style={styles.itemQuery}>{item.query}</div>
            <div style={styles.itemMeta}>
              <span style={styles.itemTime}>{formatTime(item.timestamp)}</span>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

function formatTime(ts: string): string {
  const d = new Date(ts)
  const now = new Date()
  const diffMs = now.getTime() - d.getTime()
  const diffMin = Math.floor(diffMs / 60000)
  if (diffMin < 1) return '刚刚'
  if (diffMin < 60) return `${diffMin}分钟前`
  const diffHour = Math.floor(diffMin / 60)
  if (diffHour < 24) return `${diffHour}小时前`
  return d.toLocaleDateString('zh-CN')
}

const styles: Record<string, React.CSSProperties> = {
  sidebar: {
    width: 280, height: '100%', borderRight: '1px solid var(--border)',
    background: 'var(--bg-secondary)', display: 'flex', flexDirection: 'column', flexShrink: 0,
  },
  collapsedBar: {
    width: 40, height: '100%', borderRight: '1px solid var(--border)',
    display: 'flex', alignItems: 'flex-start', justifyContent: 'center', paddingTop: 12, flexShrink: 0,
  },
  header: {
    display: 'flex', justifyContent: 'space-between', alignItems: 'center',
    padding: '16px', borderBottom: '1px solid var(--border)',
  },
  title: { fontSize: 15, fontWeight: 600 },
  toggleBtn: {
    background: 'none', border: 'none', color: 'var(--text-muted)',
    fontSize: 14, padding: '4px 8px', borderRadius: 4,
  },
  search: {
    margin: '10px 12px', padding: '8px 12px', borderRadius: 'var(--radius-sm)',
    border: '1px solid var(--border)', background: 'var(--bg-tertiary)',
    color: 'var(--text-primary)', fontSize: 13, outline: 'none',
  },
  list: { flex: 1, overflowY: 'auto', padding: '0 8px 8px' },
  empty: { textAlign: 'center', color: 'var(--text-muted)', fontSize: 13, padding: '40px 16px' },
  item: {
    padding: '10px 12px', borderRadius: 'var(--radius-sm)', cursor: 'pointer',
    marginBottom: 2, transition: 'background 0.1s',
  },
  itemQuery: {
    fontSize: 13, color: 'var(--text-primary)', lineHeight: 1.5,
    overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap',
  },
  itemMeta: { marginTop: 4, display: 'flex', gap: '8px' },
  itemTime: { fontSize: 11, color: 'var(--text-muted)' },
}
