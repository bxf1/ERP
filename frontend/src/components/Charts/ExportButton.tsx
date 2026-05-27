import { useState, useRef, useEffect } from 'react';
import type { QueryResponse } from '../models/query';
import { exportToCSV, exportToExcel } from '../services/export';

interface Props {
  data: QueryResponse;
}

export function ExportButton({ data }: Props) {
  const [open, setOpen] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (ref.current && !ref.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const filename = `query_${new Date().toISOString().slice(0, 10)}`;

  return (
    <div className="export-button" ref={ref}>
      <button className="export-trigger" onClick={() => setOpen(!open)}>
        ⬇️ 导出
      </button>
      {open && (
        <div className="export-menu">
          <button onClick={() => { exportToCSV(data, filename); setOpen(false); }}>
            CSV
          </button>
          <button onClick={() => { exportToExcel(data, filename); setOpen(false); }}>
            Excel
          </button>
        </div>
      )}
    </div>
  );
}
