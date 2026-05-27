interface Props {
  onSend: (question: string) => void;
}

const EXAMPLES = [
  '查询本月销售额前十的产品',
  '统计各部门员工数量',
  '查看最近一周的订单趋势',
  '按地区分析客户分布',
];

export function WelcomeScreen({ onSend }: Props) {
  return (
    <div className="welcome-screen">
      <div className="welcome-icon">💬</div>
      <h3>欢迎使用 AI 智能查询</h3>
      <p>用自然语言描述你想查询的数据，AI 将自动生成 SQL 并展示结果</p>
      <div className="welcome-examples">
        <span className="examples-label">试试这些：</span>
        <div className="examples-grid">
          {EXAMPLES.map((q) => (
            <button
              key={q}
              className="example-chip"
              onClick={() => onSend(q)}
            >
              {q}
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}
