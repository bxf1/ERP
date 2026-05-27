import { useState, useEffect, useCallback } from 'react';
import type { HistoryItem } from '../models/query';
import { getHistory, deleteHistory } from '../services/nl2sql';

interface Props {
  refreshKey: number;
  onSelect: (question: string) => void;
}

export function QueryHistory({ refreshKey, onSelect }: Props) {
  const [items, setItems] = useState<HistoryItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');

  const fetchHistory = useCallback(async () => {
    try {
      setError('');
      const data = await getHistory();
      setItems(data);
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : '加载失败');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchHistory();
  }, [fetchHistory, refreshKey]);

  const handleDelete = async (e: React.MouseEvent, id: string) => {
    e.stopPropagation();
    try {
      await deleteHistory(id);
      setItems((prev) => prev.filter((item) => item.id !== id));
    } catch {
      // silently fail for delete
    }
  };

  return (
    <div className="query-history">
      <div className="history-header">
        <h3>📜 查询历史</h3>
        <button className="history-refresh" onClick={fetchHistory}>
          🔄
        </button>
      </div>

      <div className="history-list">
        {loading && <p className="history-empty">加载中...</p>}
        {error && <p className="history-error">{error}</p>}
        {!loading && !error && items.length === 0 && (
          <p className="history-empty">暂无查询记录</p>
        )}
        {items.map((item) => (
          <div
            key={item.id}
            className="history-item"
            onClick={() => onSelect(item.question)}
          >
            <div className="history-item-content">
              <p className="history-question">{item.question}</p>
              <p className="history-sql">{item.sql}</p>
              <span className="history-time">{formatTime(item.created_at)}</span>
            </div>
            <button
              className="history-delete"
              onClick={(e) => handleDelete(e, item.id)}
              title="删除"
            >
              🗑
            </button>
          </div>
        ))}
      </div>
    </div>
  );
}

function formatTime(ts: string): string {
  const d = new Date(ts);
  if (isNaN(d.getTime())) return ts;
  const now = new Date();
  const diff = now.getTime() - d.getTime();

  if (diff < 60_000) return '刚刚';
  if (diff < 3_600_000) return `${Math.floor(diff / 60_000)} 分钟前`;
  if (diff < 86_400_000) return `${Math.floor(diff / 3_600_000)} 小时前`;
  return d.toLocaleDateString('zh-CN');
}
