import type { ChartType, QueryResponse } from '../models/query';

/**
 * Infer the best chart type from the query result data.
 * Heuristic:
 * - 1 numeric column + 0-1 text columns → pie
 * - 1 text/datetime column + 1+ numeric columns, few rows → bar
 * - 1 text/datetime column + 1+ numeric columns, many rows → line
 * - Otherwise → table
 */
export function inferChartType(data: QueryResponse): ChartType {
  if (data.chartType && data.chartType !== 'table') {
    return data.chartType;
  }

  const columns = data.columns;
  const rows = data.rows;

  if (rows.length === 0 || columns.length === 0) return 'table';

  const numericCols = columns.filter((c) => c.type === 'number' || c.type === 'integer' || c.type === 'float' || c.type === 'numeric');
  const textCols = columns.filter((c) => c.type !== 'number' && c.type !== 'integer' && c.type !== 'float' && c.type !== 'numeric');

  const hasNumeric = numericCols.length > 0;

  if (!hasNumeric) return 'table';

  if (textCols.length <= 1 && numericCols.length === 1 && rows.length <= 10) {
    return 'pie';
  }

  if (rows.length <= 30) return 'bar';

  return 'line';
}

export function getChartColors(): string[] {
  return [
    '#2563eb', '#dc2626', '#16a34a', '#ca8a04', '#9333ea',
    '#0891b2', '#db2777', '#ea580c', '#4f46e5', '#65a30d',
  ];
}
