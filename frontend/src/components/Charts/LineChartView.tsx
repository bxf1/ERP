import {
  LineChart,
  Line,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from 'recharts';
import type { QueryResponse } from '../models/query';
import { getChartColors } from '../utils/chartHelper';

interface Props {
  data: QueryResponse;
}

export function LineChartView({ data }: Props) {
  const colors = getChartColors();
  const categoryCol = data.columns.find((c) => c.type !== 'number' && c.type !== 'integer' && c.type !== 'float' && c.type !== 'numeric');
  const numericCols = data.columns.filter((c) => c.type === 'number' || c.type === 'integer' || c.type === 'float' || c.type === 'numeric');

  const chartData = data.rows.map((row) => {
    const entry: Record<string, unknown> = {};
    if (categoryCol) {
      entry.name = row[categoryCol.name];
    }
    numericCols.forEach((col) => {
      const val = row[col.name];
      entry[col.name] = typeof val === 'number' ? val : Number(val) || 0;
    });
    return entry;
  });

  return (
    <div className="chart-view">
      <ResponsiveContainer width="100%" height={360}>
        <LineChart data={chartData} margin={{ top: 8, right: 16, left: 8, bottom: 8 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="#e5e7eb" />
          <XAxis
            dataKey={categoryCol?.name ? undefined : 'name'}
            tick={{ fontSize: 12 }}
            stroke="#6b7280"
          />
          <YAxis tick={{ fontSize: 12 }} stroke="#6b7280" />
          <Tooltip />
          {numericCols.length > 1 && <Legend />}
          {numericCols.map((col, i) => (
            <Line
              key={col.name}
              type="monotone"
              dataKey={col.name}
              stroke={colors[i % colors.length]}
              strokeWidth={2}
              dot={{ r: 3 }}
              activeDot={{ r: 5 }}
            />
          ))}
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
