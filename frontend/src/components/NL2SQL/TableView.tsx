import type { QueryResponse } from '../models/query';

interface Props {
  data: QueryResponse;
}

export function TableView({ data }: Props) {
  if (data.columns.length === 0) {
    return <p className="empty-result">查询无返回列</p>;
  }

  return (
    <div className="table-view">
      <div className="table-scroll">
        <table>
          <thead>
            <tr>
              <th className="row-num">#</th>
              {data.columns.map((col) => (
                <th key={col.name}>
                  <div className="th-content">
                    <span>{col.name}</span>
                    <span className="col-type">{col.type}</span>
                  </div>
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {data.rows.length === 0 ? (
              <tr>
                <td colSpan={data.columns.length + 1} className="empty-cell">
                  暂无数据
                </td>
              </tr>
            ) : (
              data.rows.map((row, i) => (
                <tr key={i}>
                  <td className="row-num">{i + 1}</td>
                  {data.columns.map((col) => (
                    <td key={col.name}>{formatCell(row[col.name])}</td>
                  ))}
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function formatCell(value: unknown): string {
  if (value === null || value === undefined) return '-';
  return String(value);
}
