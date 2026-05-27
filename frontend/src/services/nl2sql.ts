import type { QueryResponse, HistoryItem } from '../models/query';

const API_BASE = '/api/v1/nl2sql';

export async function queryNL2SQL(question: string): Promise<QueryResponse> {
  const res = await fetch(`${API_BASE}/query`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ question }),
  });

  if (!res.ok) {
    const err = await res.text();
    throw new Error(err || '查询失败');
  }

  return res.json();
}

export async function getHistory(): Promise<HistoryItem[]> {
  const res = await fetch(`${API_BASE}/history`);

  if (!res.ok) {
    const err = await res.text();
    throw new Error(err || '获取历史记录失败');
  }

  return res.json();
}

export async function deleteHistory(id: string): Promise<void> {
  const res = await fetch(`${API_BASE}/history/${id}`, {
    method: 'DELETE',
  });

  if (!res.ok) {
    const err = await res.text();
    throw new Error(err || '删除历史记录失败');
  }
}
