import { useState, useMemo } from 'react';
import type { ChartType, QueryResponse } from '../models/query';
import { inferChartType } from '../utils/chartHelper';
import { TableView } from './TableView';
import { BarChartView } from './BarChartView';
import { LineChartView } from './LineChartView';
import { PieChartView } from './PieChartView';

interface Props {
  data: QueryResponse;
}

const VIEW_TABS: { key: ChartType; label: string; icon: string }[] = [
  { key: 'table', label: '表格', icon: '📋' },
  { key: 'bar', label: '柱状图', icon: '📊' },
  { key: 'line', label: '折线图', icon: '📈' },
  { key: 'pie', label: '饼图', icon: '🥧' },
];

export function ResultVisualizer({ data }: Props) {
  const defaultView = useMemo(() => inferChartType(data), [data]);
  const [activeView, setActiveView] = useState<ChartType>(defaultView);

  return (
    <div className="result-visualizer">
      <div className="view-tabs">
        {VIEW_TABS.map((tab) => (
          <button
            key={tab.key}
            className={`view-tab ${activeView === tab.key ? 'active' : ''}`}
            onClick={() => setActiveView(tab.key)}
          >
            <span className="tab-icon">{tab.icon}</span>
            <span className="tab-label">{tab.label}</span>
          </button>
        ))}
      </div>

      <div className="view-content">
        {activeView === 'table' && <TableView data={data} />}
        {activeView === 'bar' && <BarChartView data={data} />}
        {activeView === 'line' && <LineChartView data={data} />}
        {activeView === 'pie' && <PieChartView data={data} />}
      </div>
    </div>
  );
}
