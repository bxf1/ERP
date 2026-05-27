import { Table, Button, Input, InputNumber, DatePicker } from 'antd';
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons';
import { useState } from 'react';
import type { SubTableColumn } from '../types';
import dayjs from 'dayjs';

interface SubTableProps {
  columns?: SubTableColumn[];
  value?: Record<string, unknown>[];
  onChange?: (value: Record<string, unknown>[]) => void;
  disabled?: boolean;
  maxRows?: number;
  minRows?: number;
}

export function SubTable({ columns = [], value = [], onChange, disabled, maxRows, minRows = 0 }: SubTableProps) {
  const [dataSource, setDataSource] = useState<{ key: string; [k: string]: unknown }[]>(
    (value || []).map((row, i) => ({ key: String(i), ...row }))
  );

  const updateValue = (rows: { key: string; [k: string]: unknown }[]) => {
    setDataSource(rows);
    const clean = rows.map(({ key: _key, ...rest }) => rest);
    onChange?.(clean);
  };

  const handleAdd = () => {
    if (maxRows && dataSource.length >= maxRows) return;
    const newRow: Record<string, unknown> = { key: String(Date.now()) };
    columns.forEach((col) => {
      newRow[col.key] = undefined;
    });
    updateValue([...dataSource, newRow as { key: string; [k: string]: unknown }]);
  };

  const handleDelete = (key: string) => {
    if (dataSource.length <= minRows) return;
    updateValue(dataSource.filter((row) => row.key !== key));
  };

  const handleCellChange = (key: string, colKey: string, val: unknown) => {
    updateValue(dataSource.map((row) => (row.key === key ? { ...row, [colKey]: val } : row)));
  };

  const renderCell = (col: SubTableColumn, record: { key: string; [k: string]: unknown }) => {
    const val = record[col.key];
    const isDisabled = disabled;

    switch (col.type) {
      case 'text':
        return (
          <Input
            value={val as string}
            onChange={(e) => handleCellChange(record.key, col.key, e.target.value)}
            disabled={isDisabled}
            size="small"
          />
        );
      case 'number':
        return (
          <InputNumber
            value={val as number}
            onChange={(v) => handleCellChange(record.key, col.key, v)}
            disabled={isDisabled}
            size="small"
            style={{ width: '100%' }}
          />
        );
      case 'date':
        return (
          <DatePicker
            value={val ? dayjs(val as string) : null}
            onChange={(d) => handleCellChange(record.key, col.key, d?.toISOString())}
            disabled={isDisabled}
            size="small"
            style={{ width: '100%' }}
          />
        );
      default:
        return (
          <Input
            value={val as string}
            onChange={(e) => handleCellChange(record.key, col.key, e.target.value)}
            disabled={isDisabled}
            size="small"
          />
        );
    }
  };

  const tableColumns = [
    ...columns.map((col) => ({
      key: col.key,
      title: col.title,
      dataIndex: col.key,
      width: col.width,
      render: (_: unknown, record: { key: string; [k: string]: unknown }) => renderCell(col, record),
    })),
    ...(disabled
      ? []
      : [
          {
            key: 'actions',
            title: '操作',
            width: 80,
            render: (_: unknown, record: { key: string; [k: string]: unknown }) => (
              <Button
                type="link"
                danger
                size="small"
                icon={<DeleteOutlined />}
                onClick={() => handleDelete(record.key)}
                disabled={dataSource.length <= minRows}
              />
            ),
          },
        ]),
  ];

  return (
    <div>
      <Table
        columns={tableColumns}
        dataSource={dataSource}
        pagination={false}
        size="small"
        bordered
        locale={{ emptyText: '暂无数据' }}
      />
      {!disabled && (
        <Button
          type="dashed"
          onClick={handleAdd}
          block
          icon={<PlusOutlined />}
          style={{ marginTop: 8 }}
          disabled={!!maxRows && dataSource.length >= maxRows}
        >
          添加行
        </Button>
      )}
    </div>
  );
}

export default SubTable;
