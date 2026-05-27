import * as XLSX from 'xlsx'
import type { QueryResult } from '../models/nl2sql'

export function exportCSV(result: QueryResult, filename: string = 'query_result.csv'): void {
  const header = result.columns.join(',')
  const body = result.rows.map(row =>
    result.columns.map(col => {
      const val = row[col]
      if (val === null || val === undefined) return ''
      const str = String(val)
      return str.includes(',') || str.includes('"') || str.includes('\n')
        ? `"${str.replace(/"/g, '""')}"`
        : str
    }).join(',')
  ).join('\n')
  const csv = '﻿' + header + '\n' + body
  downloadFile(csv, filename, 'text/csv;charset=utf-8')
}

export function exportExcel(result: QueryResult, filename: string = 'query_result.xlsx'): void {
  const worksheet = XLSX.utils.json_to_sheet(result.rows, { header: result.columns })
  const workbook = XLSX.utils.book_new()
  XLSX.utils.book_append_sheet(workbook, worksheet, 'Results')
  XLSX.writeFile(workbook, filename)
}

function downloadFile(content: string, filename: string, mimeType: string): void {
  const blob = new Blob([content], { type: mimeType })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
  URL.revokeObjectURL(url)
}
