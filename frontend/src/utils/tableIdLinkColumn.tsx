import { Tooltip } from 'antd';
import type { ColumnType } from 'antd/es/table';
import { TABLE_COL_WIDTH } from './tableActionColumn';
import { TABLE_COL_ID, cellClassNames } from './tableEllipsisColumn';

/** 长 ID 列：等宽字体 + 省略，点击跳转 */
export function tableIdLinkColumn<T>(
  title: string,
  dataIndex: string,
  onClick: (id: string) => void,
): ColumnType<T> {
  return {
    title,
    dataIndex,
    key: dataIndex,
    width: TABLE_COL_WIDTH.id,
    ellipsis: { showTitle: false },
    ...cellClassNames(TABLE_COL_ID),
    render: (value: string) => (
      <Tooltip title={value}>
        <button type="button" className="ota-table-id-link" onClick={() => onClick(value)}>
          {value}
        </button>
      </Tooltip>
    ),
  };
}

/** 计算字段的可点击 ID 列 */
export function tableIdLinkRenderColumn<T>(
  title: string,
  key: string,
  getValue: (record: T) => string,
  onClick: (record: T) => void,
): ColumnType<T> {
  return {
    title,
    key,
    width: TABLE_COL_WIDTH.id,
    ellipsis: { showTitle: false },
    ...cellClassNames(TABLE_COL_ID),
    render: (_: unknown, record: T) => {
      const value = getValue(record);
      return (
        <Tooltip title={value}>
          <button type="button" className="ota-table-id-link" onClick={() => onClick(record)}>
            {value}
          </button>
        </Tooltip>
      );
    },
  };
}
