import {
  PieChart,
  Pie,
  Cell,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from 'recharts';
import type { QueryResponse } from '../models/query';
import { getChartColors } from '../utils/chartHelper';

interface Props {
  data: QueryResponse;
}

export function PieChartView({ data }: Props) {
  const colors = getChartColors();
  const categoryCol = data.columns.find((c) => c.type !== 'number' && c.type !== 'integer' && c.type !== 'float' && c.type !== 'numeric');
  const numericCol = data.columns.find((c) => c.type === 'number' || c.type === 'integer' || c.type === 'float' || c.type === 'numeric');

  const pieData = data.rows.map((row) => {
    const name = categoryCol ? String(row[categoryCol.name] ?? '') : '';
    const value = numericCol
      ? (typeof row[numericCol.name] === 'number'
          ? row[numericCol.name]
          : Number(row[numericCol.name]) || 0)
      : 0;
    return { name, value: value as number };
  });

  return (
    <div className="chart-view chart-pie">
      <ResponsiveContainer width="100%" height={360}>
        <PieChart>
          <Pie
            data={pieData}
            cx="50%"
            cy="50%"
            outerRadius={120}
            dataKey="value"
            label={({ name, percent }) =>
              `${name} ${(percent * 100).toFixed(0)}%`
            }
          >
            {pieData.map((_, i) => (
              <Cell key={i} fill={colors[i % colors.length]} />
            ))}
          </Pie>
          <Tooltip />
          <Legend />
        </PieChart>
      </ResponsiveContainer>
    </div>
  );
}
