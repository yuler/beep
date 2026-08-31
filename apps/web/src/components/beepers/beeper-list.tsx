import { Link, useNavigate, useRouter } from "@tanstack/react-router";
import { createColumnHelper } from "@tanstack/react-table";
import { Edit } from "lucide-react";
import { useMemo, useState } from "react";

import { EditBeeperDialog } from "@/components/beepers/edit-beeper-dialog";
import { Button } from "@/components/ui/button";
import {
	DataTable,
	type dataTableFeatures,
	makeSelectColumn,
	SortableHeader,
} from "@/components/ui/data-table";
import { ProgressBar, StatusPill } from "@/components/ui/status-pill";
import type { Beeper } from "@/lib/api/beepers";
import { formatBeepScheduleTime } from "@/lib/beep-datetime";
import {
	beeperHealthIsDestructive,
	beeperHealthLabel,
} from "@/lib/beeper-health";
import { runSuccessRate } from "@/lib/run-success-rate";
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

function useBeeperColumns(slug: string, onEdit: (beeper: Beeper) => void) {
	return useMemo(
		() =>
			columnHelper.columns([
				makeSelectColumn(columnHelper),
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
				columnHelper.accessor((row) => runSuccessRate(row.runs ?? []), {
					id: "run_success",
					header: ({ column }) => (
						<SortableHeader column={column} label="Run success" />
					),
					cell: ({ row }) => (
						<ProgressBar value={runSuccessRate(row.original.runs ?? [])} />
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
				columnHelper.display({
					id: "actions",
					header: () => <span className="sr-only">Actions</span>,
					cell: ({ row }) => {
						const beeper = row.original;
						return (
							<div className="flex items-center justify-end">
								<Button
									type="button"
									variant="ghost"
									size="xs"
									className="h-7 px-2 text-xs text-muted-foreground hover:text-foreground"
									aria-label={`Edit beeper ${beeper.title}`}
									onClick={(e) => {
										e.stopPropagation();
										onEdit(beeper);
									}}
									data-no-row-nav
								>
									<Edit className="size-3.5" data-icon="inline-start" />
									Edit
								</Button>
							</div>
						);
					},
					enableSorting: false,
				}),
			]),
		[slug, onEdit],
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
	const router = useRouter();
	const [editingBeeper, setEditingBeeper] = useState<Beeper | null>(null);

	const columns = useBeeperColumns(slug, (beeper) => {
		setEditingBeeper(beeper);
	});

	return (
		<>
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

			{editingBeeper ? (
				<EditBeeperDialog
					beeper={editingBeeper}
					slug={slug}
					open={editingBeeper !== null}
					onOpenChange={(open) => {
						if (!open) {
							setEditingBeeper(null);
						}
					}}
					onSuccess={async () => {
						await router.invalidate();
					}}
				/>
			) : null}
		</>
	);
}
