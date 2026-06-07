import { Tooltip } from 'antd';
import type { ColumnType } from 'antd/es/table';
import type { ReactNode } from 'react';
import { TABLE_COL_WIDTH } from './tableActionColumn';

export const TABLE_COL_FLEX = 'ota-table-col-flex';
export const TABLE_COL_COMPACT = 'ota-table-col-compact';
export const TABLE_COL_ID = 'ota-table-col-id';

const TEXT_CLASS = 'ota-table-cell-text';

type WidthPreset = keyof typeof TABLE_COL_WIDTH;

export function cellClassNames(...names: string[]) {
  return {
    className: names.join(' '),
    onCell: () => ({ className: names.join(' ') }),
  };
}

export function ellipsisCell(text: string | null | undefined, options?: { monospace?: boolean }): ReactNode {
  const value = text?.trim() ? text : '-';
  if (value === '-') {
    return value;
  }
  return (
    <Tooltip title={value}>
      <span className={`${TEXT_CLASS}${options?.monospace ? ' ota-table-cell-mono' : ''}`}>{value}</span>
    </Tooltip>
  );
}

type EllipsisOptions<T> = {
  size?: WidthPreset;
  monospace?: boolean;
  render?: (value: unknown, record: T) => string;
};

function ellipsisColumnBase<T>(
  title: string,
  key: string,
  colClass: string,
  width: number,
  render: ColumnType<T>['render'],
  dataIndex?: string,
): ColumnType<T> {
  return {
    title,
    key,
    dataIndex,
    width,
    ellipsis: { showTitle: false },
    ...cellClassNames(colClass),
    render,
  };
}

/** 文本列：单行 ellipsis + Tooltip（无自定义 render 时用 Ant Design 原生截断） */
export function tableEllipsisColumn<T>(
  title: string,
  dataIndex: string,
  options?: EllipsisOptions<T>,
): ColumnType<T> {
  const width = TABLE_COL_WIDTH[options?.size ?? 'text'];
  const shared = {
    title,
    dataIndex,
    key: dataIndex,
    width,
    ...cellClassNames(TABLE_COL_FLEX),
  };
  if (options?.render) {
    return {
      ...shared,
      ellipsis: { showTitle: false },
      render: (_: unknown, record: T) =>
        ellipsisCell(options.render!(_, record), { monospace: options.monospace }),
    };
  }
  return {
    ...shared,
    ellipsis: { showTitle: true },
  };
}

/** 计算字段列：单行 ellipsis + Tooltip */
export function tableEllipsisRenderColumn<T>(
  title: string,
  key: string,
  getText: (record: T) => string,
  options?: Pick<EllipsisOptions<T>, 'size' | 'monospace'>,
): ColumnType<T> {
  const width = TABLE_COL_WIDTH[options?.size ?? 'wide'];
  return ellipsisColumnBase(
    title,
    key,
    TABLE_COL_FLEX,
    width,
    (_, record) => ellipsisCell(getText(record), { monospace: options?.monospace }),
  );
}

/** Tag / 状态等短内容列 */
export function tableCompactColumn<T>(
  title: string,
  key: string,
  render: NonNullable<ColumnType<T>['render']>,
  dataIndex?: string,
): ColumnType<T> {
  return {
    title,
    key,
    dataIndex,
    width: TABLE_COL_WIDTH.tag,
    ...cellClassNames(TABLE_COL_COMPACT),
    render,
  };
}
