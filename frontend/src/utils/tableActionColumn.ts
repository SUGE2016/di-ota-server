import type { ColumnType } from 'antd/es/table';
import type { TableProps } from 'antd';

const FIXED_ACTION_CLASS = 'ota-table-fixed-col';

export type TableActionColumnOptions = {
  /** 该表可能出现的操作按钮文案，用于估算列宽 */
  actionLabels: string[];
  /** 同一行最多并排几个按钮（如任务页最多 3 个） */
  maxButtonsPerRow?: number;
};

/** 按按钮文案估算冻结操作列宽度（Ant Design 冻结列仍需数值 width，但不写死统一像素） */
export function measureActionColumnWidth(options: TableActionColumnOptions): number {
  const { actionLabels, maxButtonsPerRow } = options;
  if (actionLabels.length === 0) {
    return 88;
  }
  const charPx = 13;
  const btnPad = 10;
  const gap = 4;
  const cellPad = 16;
  const perRow = Math.min(maxButtonsPerRow ?? actionLabels.length, actionLabels.length);
  const row = [...actionLabels].sort((a, b) => b.length - a.length).slice(0, perRow);
  const content = row.reduce(
    (sum, label, index) => sum + label.length * charPx + btnPad + (index > 0 ? gap : 0),
    0,
  );
  return Math.max(72, Math.ceil(content + cellPad));
}

export function tableActionColumn<T>(
  render: NonNullable<ColumnType<T>['render']>,
  options: TableActionColumnOptions,
): ColumnType<T> {
  return {
    title: '操作',
    key: 'actions',
    width: measureActionColumnWidth(options),
    fixed: 'right',
    align: 'center',
    className: FIXED_ACTION_CLASS,
    onHeaderCell: () => ({ className: FIXED_ACTION_CLASS }),
    onCell: () => ({ className: FIXED_ACTION_CLASS }),
    render,
  };
}

/** 带冻结列的列表表格通用属性（数据列 auto 布局，仅操作列按内容给定 width） */
export const listTableProps: Pick<TableProps, 'size' | 'tableLayout' | 'className'> = {
  size: 'middle',
  tableLayout: 'auto',
  className: 'ota-list-table',
};

export function listTableScroll(minWidth: number, rowCount: number): TableProps['scroll'] {
  return rowCount > 0 ? { x: minWidth } : undefined;
}
