import { Link, useNavigate } from "@tanstack/react-router";
import { createColumnHelper } from "@tanstack/react-table";
import { useMemo } from "react";

import { Checkbox } from "@/components/ui/checkbox";
import {
	DataTable,
	type dataTableFeatures,
	SortableHeader,
} from "@/components/ui/data-table";
import { ProgressBar, StatusPill } from "@/components/ui/status-pill";
import type { Beeper } from "@/lib/api/beepers";
import { formatBeepScheduleTime } from "@/lib/beep-datetime";
import {
	beeperHealthIsDestructive,
	beeperHealthLabel,
} from "@/lib/beeper-health";
import { shortId } from "@/lib/short-id";

const columnHelper = createColumnHelper<typeof dataTableFeatures, Beeper>();

function beeperStatusTone(status: Beeper["status"]) {
	switch (status) {
		case "active":
			return "emerald" as const;
		case "firing":
			return "amber" as const;
		case "paused":
			return "muted" as const;
		default:
			return "outline" as const;
	}
}

function healthTone(beeper: Beeper) {
	const label = beeperHealthLabel(beeper);
	if (label === "ok") return "emerald" as const;
	if (label === "alerting") return "rose" as const;
	return "amber" as const;
}

function runSuccessRate(beeper: Beeper) {
	const runs = beeper.runs ?? [];
	if (runs.length === 0) return 0;
	const succeeded = runs.filter((run) => run.status === "succeeded").length;
	return Math.round((succeeded / runs.length) * 100);
}

function useBeeperColumns(slug: string) {
	return useMemo(
		() =>
			columnHelper.columns([
				columnHelper.display({
					id: "select",
					header: ({ table }) => (
						<Checkbox
							checked={
								table.getIsAllPageRowsSelected()
									? true
									: table.getIsSomePageRowsSelected()
										? "indeterminate"
										: false
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
				}),
				columnHelper.accessor("title", {
					id: "title",
					header: ({ column }) => (
						<SortableHeader column={column} label="Beeper" />
					),
					cell: ({ row }) => {
						const beeper = row.original;
						return (
							<div className="flex min-w-48 flex-col gap-0.5">
								<span className="font-mono text-[11px] text-muted-foreground">
									#{shortId(beeper.id)}
								</span>
								<Link
									to="/$account_slug/beepers/$beeperId"
									params={{
										account_slug: slug,
										beeperId: beeper.id,
									}}
									className="font-medium text-foreground hover:text-primary"
									onClick={(event) => event.stopPropagation()}
									data-no-row-nav
								>
									{beeper.title}
								</Link>
							</div>
						);
					},
				}),
				columnHelper.accessor((row) => row.beeper_app?.name ?? "—", {
					id: "app",
					header: ({ column }) => (
						<SortableHeader column={column} label="App" />
					),
					cell: ({ row }) => (
						<span className="text-sm text-muted-foreground">
							{row.original.beeper_app?.name ?? "—"}
						</span>
					),
				}),
				columnHelper.accessor("status", {
					id: "status",
					header: ({ column }) => (
						<SortableHeader column={column} label="Status" />
					),
					cell: ({ row }) => (
						<StatusPill
							label={row.original.status}
							tone={beeperStatusTone(row.original.status)}
						/>
					),
				}),
				columnHelper.accessor((row) => beeperHealthLabel(row), {
					id: "health",
					header: ({ column }) => (
						<SortableHeader column={column} label="Health" />
					),
					cell: ({ row }) => {
						const beeper = row.original;
						return (
							<StatusPill
								label={beeperHealthLabel(beeper).toUpperCase()}
								tone={
									beeperHealthIsDestructive(beeper)
										? "rose"
										: healthTone(beeper)
								}
							/>
						);
					},
				}),
				columnHelper.accessor("cron", {
					id: "cron",
					header: ({ column }) => (
						<SortableHeader column={column} label="Schedule" />
					),
					cell: ({ row }) => (
						<span className="font-mono text-xs text-muted-foreground">
							{row.original.cron}
						</span>
					),
				}),
				columnHelper.accessor((row) => runSuccessRate(row), {
					id: "progress",
					header: ({ column }) => (
						<SortableHeader column={column} label="Progress" />
					),
					cell: ({ row }) => (
						<ProgressBar value={runSuccessRate(row.original)} />
					),
				}),
				columnHelper.accessor((row) => row.runs?.length ?? 0, {
					id: "runs",
					header: ({ column }) => (
						<SortableHeader column={column} label="Runs" />
					),
					cell: ({ row }) => (
						<span className="tabular-nums text-sm">
							{row.original.runs?.length ?? 0}
						</span>
					),
				}),
				columnHelper.accessor((row) => row.next_run_at ?? "", {
					id: "next_run",
					header: ({ column }) => (
						<SortableHeader column={column} label="Next run" />
					),
					cell: ({ row }) => {
						const beeper = row.original;
						if (!beeper.next_run_at) {
							return <span className="text-sm text-muted-foreground">—</span>;
						}
						return (
							<span className="text-sm tabular-nums text-foreground">
								{formatBeepScheduleTime(
									beeper.next_run_at,
									beeper.timezone,
									"short",
								)}
							</span>
						);
					},
				}),
			]),
		[slug],
	);
}

export function BeeperList({
	beepers,
	slug,
}: {
	beepers: Beeper[];
	slug: string;
}) {
	const navigate = useNavigate();
	const columns = useBeeperColumns(slug);

	return (
		<DataTable
			data={beepers}
			columns={columns}
			getRowId={(beeper) => beeper.id}
			emptyMessage="No beepers installed yet."
			onRowClick={(beeper) =>
				navigate({
					to: "/$account_slug/beepers/$beeperId",
					params: { account_slug: slug, beeperId: beeper.id },
				})
			}
		/>
	);
}
