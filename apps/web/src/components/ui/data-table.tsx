import {
	type Column,
	type createColumnHelper,
	createSortedRowModel,
	type OnChangeFn,
	type RowData,
	type RowSelectionState,
	rowSelectionFeature,
	rowSortingFeature,
	type SortingState,
	type TableOptions,
	tableFeatures,
	useTable,
} from "@tanstack/react-table";
import { ArrowDown, ArrowUp, ArrowUpDown } from "lucide-react";
import { useState } from "react";

import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import {
	Table,
	TableBody,
	TableCell,
	TableHead,
	TableHeader,
	TableRow,
} from "@/components/ui/table";
import { cn } from "@/lib/utils";

export const dataTableFeatures = tableFeatures({
	rowSortingFeature,
	sortedRowModel: createSortedRowModel(),
	rowSelectionFeature,
});

export type DataTableFeatures = typeof dataTableFeatures;

type ColumnHelperFor<TData extends RowData> = ReturnType<
	typeof createColumnHelper<typeof dataTableFeatures, TData>
>;

export function makeSelectColumn<TData extends RowData>(
	columnHelper: ColumnHelperFor<TData>,
) {
	return columnHelper.display({
		id: "select",
		header: ({ table }) => (
			<Checkbox
				checked={table.getIsAllPageRowsSelected()}
				indeterminate={
					table.getIsSomePageRowsSelected() && !table.getIsAllPageRowsSelected()
				}
				onCheckedChange={(value) =>
					table.toggleAllPageRowsSelected(value === true)
				}
				aria-label="Select all"
			/>
		),
		cell: ({ row }) => (
			<Checkbox
				checked={row.getIsSelected()}
				onCheckedChange={(value) => row.toggleSelected(value === true)}
				aria-label="Select row"
				data-no-row-nav
			/>
		),
		enableSorting: false,
	});
}

export type DataTableColumnMeta = {
	className?: string;
};

function columnMetaClassName(meta: unknown) {
	return (meta as DataTableColumnMeta | undefined)?.className;
}

export function SortableHeader<TData extends RowData, TValue>({
	column,
	label,
}: {
	column: Column<typeof dataTableFeatures, TData, TValue>;
	label: string;
}) {
	const sorted = column.getIsSorted();
	return (
		<Button
			type="button"
			variant="ghost"
			size="sm"
			className="-ml-2 h-8 gap-1 px-2 text-xs font-medium text-muted-foreground hover:text-foreground"
			onClick={() => column.toggleSorting(sorted === "asc")}
		>
			{label}
			{sorted === "asc" ? (
				<ArrowUp className="size-3.5" />
			) : sorted === "desc" ? (
				<ArrowDown className="size-3.5" />
			) : (
				<ArrowUpDown className="size-3.5 opacity-40" />
			)}
		</Button>
	);
}

type DataTableProps<TData extends RowData> = {
	data: TData[];
	columns: TableOptions<typeof dataTableFeatures, TData>["columns"];
	getRowId?: TableOptions<typeof dataTableFeatures, TData>["getRowId"];
	className?: string;
	emptyMessage?: string;
	onRowClick?: (row: TData) => void;
	sorting?: SortingState;
	onSortingChange?: OnChangeFn<SortingState>;
	manualSorting?: boolean;
};

export function DataTable<TData extends RowData>({
	data,
	columns,
	getRowId,
	className,
	emptyMessage = "No rows to display.",
	onRowClick,
	sorting: sortingProp,
	onSortingChange,
	manualSorting = false,
}: DataTableProps<TData>) {
	const [internalSorting, setInternalSorting] = useState<SortingState>([]);
	const [rowSelection, setRowSelection] = useState<RowSelectionState>({});
	const sorting = sortingProp ?? internalSorting;

	const table = useTable<typeof dataTableFeatures, TData>({
		features: dataTableFeatures,
		data,
		columns,
		getRowId,
		manualSorting,
		state: { sorting, rowSelection },
		onSortingChange: onSortingChange ?? setInternalSorting,
		onRowSelectionChange: setRowSelection,
	});

	const rows = table.getRowModel().rows;
	const columnCount = table.getAllColumns().length;

	return (
		<div
			className={cn(
				"overflow-hidden rounded-lg border border-border bg-card",
				className,
			)}
		>
			<Table>
				<TableHeader>
					{table.getHeaderGroups().map((headerGroup) => (
						<TableRow
							key={headerGroup.id}
							className="border-b bg-muted/30 hover:bg-muted/30"
						>
							{headerGroup.headers.map((header) => (
								<TableHead
									key={header.id}
									className={cn(
										"h-10 px-3 text-xs font-medium text-muted-foreground",
										columnMetaClassName(header.column.columnDef.meta),
									)}
								>
									{header.isPlaceholder ? null : (
										<table.FlexRender header={header} />
									)}
								</TableHead>
							))}
						</TableRow>
					))}
				</TableHeader>
				<TableBody>
					{rows.length === 0 ? (
						<TableRow className="hover:bg-transparent">
							<TableCell
								colSpan={columnCount}
								className="h-24 px-3 text-center text-sm text-muted-foreground"
							>
								{emptyMessage}
							</TableCell>
						</TableRow>
					) : (
						rows.map((row) => (
							<TableRow
								key={row.id}
								data-state={row.getIsSelected() ? "selected" : undefined}
								className={cn(
									"border-b border-border/60",
									onRowClick && "cursor-pointer",
								)}
								tabIndex={onRowClick ? 0 : undefined}
								role={onRowClick ? "button" : undefined}
								onClick={
									onRowClick
										? (event) => {
												const target = event.target as HTMLElement;
												if (
													target.closest(
														'button, a, input, [role="checkbox"], [data-no-row-nav]',
													)
												) {
													return;
												}
												onRowClick(row.original);
											}
										: undefined
								}
								onKeyDown={
									onRowClick
										? (event) => {
												if (event.key !== "Enter" && event.key !== " ") {
													return;
												}
												event.preventDefault();
												onRowClick(row.original);
											}
										: undefined
								}
							>
								{row.getAllCells().map((cell) => (
									<TableCell
										key={cell.id}
										className={cn(
											"min-w-0 px-3 py-3 align-middle",
											columnMetaClassName(cell.column.columnDef.meta),
										)}
									>
										<table.FlexRender cell={cell} />
									</TableCell>
								))}
							</TableRow>
						))
					)}
				</TableBody>
			</Table>
		</div>
	);
}
