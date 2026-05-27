export interface ColumnDef {
  name: string;
  type: string;
}

export interface QueryResponse {
  question: string;
  sql: string;
  columns: ColumnDef[];
  rows: Record<string, unknown>[];
  chartType: ChartType;
}

export type ChartType = 'table' | 'bar' | 'line' | 'pie';

export interface HistoryItem {
  id: string;
  question: string;
  sql: string;
  created_at: string;
}

export interface ChatMessage {
  id: string;
  role: 'user' | 'assistant';
  segments: MessageSegment[];
  data?: QueryResponse;
}

export type MessageSegment =
  | { type: 'text'; content: string }
  | { type: 'sql'; content: string }
  | { type: 'result'; data: QueryResponse };
