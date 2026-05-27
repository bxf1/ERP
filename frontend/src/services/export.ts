import * as XLSX from 'xlsx';
import type { QueryResponse } from '../models/query';

export function exportToCSV(data: QueryResponse, filename: string): void {
  const csv = rowsToCSV(data.columns.map((c) => c.name), data.rows);
  downloadFile(csv, `${filename}.csv`, 'text/csv;charset=utf-8');
}

export function exportToExcel(data: QueryResponse, filename: string): void {
  const sheetData = [
    data.columns.map((c) => c.name),
    ...data.rows.map((row) => data.columns.map((c) => String(row[c.name] ?? ''))),
  ];

  const ws = XLSX.utils.aoa_to_sheet(sheetData);
  const wb = XLSX.utils.book_new();
  XLSX.utils.book_append_sheet(wb, ws, 'Query Result');
  XLSX.writeFile(wb, `${filename}.xlsx`);
}

function rowsToCSV(headers: string[], rows: Record<string, unknown>[]): string {
  const escape = (v: string) => {
    if (v.includes(',') || v.includes('"') || v.includes('\n')) {
      return `"${v.replace(/"/g, '""')}"`;
    }
    return v;
  };

  const headerLine = headers.map(escape).join(',');
  const dataLines = rows.map((row) =>
    headers.map((h) => escape(String(row[h] ?? ''))).join(',')
  );

  return '﻿' + [headerLine, ...dataLines].join('\n');
}

function downloadFile(content: string, filename: string, mimeType: string): void {
  const blob = new Blob([content], { type: mimeType });
  const url = URL.createObjectURL(blob);
  const a = document.createElement('a');
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}
