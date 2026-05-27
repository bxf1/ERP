export interface QueryMessage {
  id: string
  role: 'user' | 'assistant'
  content: string
  sql?: string
  result?: QueryResult
  timestamp: string
}

export interface QueryResult {
  columns: string[]
  rows: Record<string, unknown>[]
  rowCount: number
  executionMs: number
}

export interface QueryHistoryItem {
  id: string
  query: string
  sql: string
  timestamp: string
}

export type ChartType = 'table' | 'bar' | 'line' | 'pie'

export interface NL2SQLRequest {
  query: string
}

export interface NL2SQLResponse {
  sql: string
  result: QueryResult
  explanation: string
}
