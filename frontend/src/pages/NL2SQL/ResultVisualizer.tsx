import { useState, useMemo } from 'react'
import {
  BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
  LineChart, Line, PieChart, Pie, Cell, Legend,
} from 'recharts'
import type { QueryResult, ChartType } from '../../models/nl2sql'
import { exportCSV, exportExcel } from '../../utils/export'

interface Props {
  result: QueryResult
}

const CHART_COLORS = ['#6366f1', '#22c55e', '#f59e0b', '#ef4444', '#8b5cf6', '#06b6d4', '#ec4899', '#f97316']

function detectChartType(result: QueryResult): ChartType {
  const { columns, rows } = result
  if (rows.length === 0) return 'table'
  if (rows.length === 1) return 'table'

  const numericCols = columns.filter(col =>
    rows.some(row => typeof row[col] === 'number')
  )
  const stringCols = columns.filter(col =>
    rows.some(row => typeof row[col] === 'string' && isNaN(Number(row[col])))
  )

  if (numericCols.length === 1 && stringCols.length >= 1 && rows.length <= 10) {
    return 'pie'
  }
  if (numericCols.length >= 1 && stringCols.length >= 1) {
    if (rows.length > 20) return 'line'
    return 'bar'
  }
  return 'table'
}

export default function ResultVisualizer({ result }: Props) {
  const detectedType = useMemo(() => detectChartType(result), [result])
  const chartTypes: ChartType[] = ['table', 'bar', 'line', 'pie']
  const [activeChart, setActiveChart] = useState<ChartType>(detectedType)
  const [showExport, setShowExport] = useState(false)

  const chartData = useMemo(() => {
    return result.rows.map(row => {
      const item: Record<string, unknown> = {}
      for (const col of result.columns) {
        const val = row[col]
        item[col] = typeof val === 'string' && !isNaN(Number(val)) ? Number(val) : val
      }
      return item
    })
  }, [result])

  const numCols = result.columns.filter(c =>
    result.rows.some(r => typeof r[c] === 'number')
  )
  const catCol = result.columns.find(c =>
    result.rows.some(r => typeof r[c] === 'string')
  ) || result.columns[0]

  const dataKey = numCols[0] || result.columns[1] || result.columns[0]

  return (
    <div style={styles.wrapper}>
      <div style={styles.header}>
        <div style={styles.headerLeft}>
          <span style={styles.rowCount}>{result.rowCount} rows</span>
          <span style={styles.execTime}>{result.executionMs}ms</span>
        </div>
        <div style={styles.headerRight}>
          <div style={styles.chartTabs}>
            {chartTypes.map(t => (
              <button
                key={t}
                style={{
                  ...styles.chartTab,
                  background: activeChart === t ? 'var(--accent-subtle)' : 'transparent',
                  color: activeChart === t ? 'var(--accent)' : 'var(--text-muted)',
                }}
                onClick={() => setActiveChart(t)}
              >
                {t === 'table' ? '表' : t === 'bar' ? '柱' : t === 'line' ? '线' : '饼'}
              </button>
            ))}
          </div>
          <div style={{ position: 'relative' }}>
            <button style={styles.exportBtn} onClick={() => setShowExport(!showExport)}>
              导出
            </button>
            {showExport && (
              <div style={styles.exportMenu}>
                <button style={styles.exportItem} onClick={() => { exportCSV(result); setShowExport(false) }}>
                  CSV
                </button>
                <button style={styles.exportItem} onClick={() => { exportExcel(result); setShowExport(false) }}>
                  Excel
                </button>
              </div>
            )}
          </div>
        </div>
      </div>

      <div style={styles.content}>
        {activeChart === 'table' && (
          <div style={styles.tableWrap}>
            <table style={styles.table}>
              <thead>
                <tr>
                  {result.columns.map(col => (
                    <th key={col} style={styles.th}>{col}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {result.rows.map((row, i) => (
                  <tr key={i} style={i % 2 === 0 ? { background: 'rgba(255,255,255,0.02)' } : undefined}>
                    {result.columns.map(col => (
                      <td key={col} style={styles.td}>
                        {row[col] === null ? <span style={{ color: 'var(--text-muted)' }}>NULL</span> : String(row[col])}
                      </td>
                    ))}
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {activeChart === 'bar' && (
          <ResponsiveContainer width="100%" height={350}>
            <BarChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
              <XAxis dataKey={catCol} tick={{ fill: 'var(--text-secondary)', fontSize: 12 }} stroke="var(--border)" />
              <YAxis tick={{ fill: 'var(--text-secondary)', fontSize: 12 }} stroke="var(--border)" />
              <Tooltip
                contentStyle={{
                  background: 'var(--bg-secondary)',
                  border: '1px solid var(--border)',
                  borderRadius: 8,
                  color: 'var(--text-primary)',
                }}
              />
              <Bar dataKey={dataKey} fill="var(--accent)" radius={[4, 4, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        )}

        {activeChart === 'line' && (
          <ResponsiveContainer width="100%" height={350}>
            <LineChart data={chartData}>
              <CartesianGrid strokeDasharray="3 3" stroke="var(--border)" />
              <XAxis dataKey={catCol} tick={{ fill: 'var(--text-secondary)', fontSize: 12 }} stroke="var(--border)" />
              <YAxis tick={{ fill: 'var(--text-secondary)', fontSize: 12 }} stroke="var(--border)" />
              <Tooltip
                contentStyle={{
                  background: 'var(--bg-secondary)',
                  border: '1px solid var(--border)',
                  borderRadius: 8,
                  color: 'var(--text-primary)',
                }}
              />
              <Line type="monotone" dataKey={dataKey} stroke="var(--accent)" strokeWidth={2} dot={{ fill: 'var(--accent)' }} />
            </LineChart>
          </ResponsiveContainer>
        )}

        {activeChart === 'pie' && (
          <ResponsiveContainer width="100%" height={350}>
            <PieChart>
              <Pie
                data={chartData}
                dataKey={dataKey}
                nameKey={catCol}
                cx="50%"
                cy="50%"
                outerRadius={130}
                label={({ name, value }) => `${name}: ${value}`}
              >
                {chartData.map((_, i) => (
                  <Cell key={i} fill={CHART_COLORS[i % CHART_COLORS.length]} />
                ))}
              </Pie>
              <Tooltip
                contentStyle={{
                  background: 'var(--bg-secondary)',
                  border: '1px solid var(--border)',
                  borderRadius: 8,
                  color: 'var(--text-primary)',
                }}
              />
              <Legend wrapperStyle={{ color: 'var(--text-secondary)', fontSize: 12 }} />
            </PieChart>
          </ResponsiveContainer>
        )}
      </div>
    </div>
  )
}

const styles: Record<string, React.CSSProperties> = {
  wrapper: {
    border: '1px solid var(--border)',
    borderRadius: 'var(--radius)',
    overflow: 'hidden',
    background: 'var(--bg-secondary)',
  },
  header: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    padding: '10px 14px',
    borderBottom: '1px solid var(--border)',
  },
  headerLeft: { display: 'flex', gap: '12px', alignItems: 'center' },
  rowCount: { fontSize: 12, color: 'var(--text-secondary)' },
  execTime: { fontSize: 12, color: 'var(--text-muted)' },
  headerRight: { display: 'flex', gap: '12px', alignItems: 'center' },
  chartTabs: {
    display: 'flex', gap: '2px', background: 'var(--bg-tertiary)',
    borderRadius: 'var(--radius-sm)', padding: '2px',
  },
  chartTab: {
    padding: '5px 10px', border: 'none', borderRadius: '4px',
    fontSize: 12, fontWeight: 500, cursor: 'pointer',
  },
  exportBtn: {
    padding: '6px 14px', border: '1px solid var(--border)',
    borderRadius: 'var(--radius-sm)', background: 'var(--bg-tertiary)',
    color: 'var(--text-secondary)', fontSize: 12, cursor: 'pointer',
  },
  exportMenu: {
    position: 'absolute', top: '100%', right: 0, marginTop: 4,
    background: 'var(--bg-tertiary)', border: '1px solid var(--border)',
    borderRadius: 'var(--radius-sm)', overflow: 'hidden', zIndex: 10, minWidth: 80,
  },
  exportItem: {
    display: 'block', width: '100%', padding: '8px 14px', border: 'none',
    background: 'transparent', color: 'var(--text-secondary)', fontSize: 13,
    textAlign: 'left', cursor: 'pointer',
  },
  content: { padding: '14px' },
  tableWrap: { maxHeight: 400, overflowY: 'auto' },
  table: { width: '100%', borderCollapse: 'collapse', fontSize: 13 },
  th: {
    textAlign: 'left', padding: '10px 12px', borderBottom: '2px solid var(--border)',
    color: 'var(--text-muted)', fontWeight: 600, fontSize: 12, whiteSpace: 'nowrap',
    position: 'sticky', top: 0, background: 'var(--bg-secondary)',
  },
  td: {
    padding: '8px 12px', borderBottom: '1px solid var(--border)',
    color: 'var(--text-primary)', maxWidth: 300, overflow: 'hidden',
    textOverflow: 'ellipsis', whiteSpace: 'nowrap',
  },
}
