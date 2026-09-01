import { Link, useNavigate } from "@tanstack/react-router";
import { createColumnHelper } from "@tanstack/react-table";
import { Activity, Clock, Repeat, Search, Sparkles } from "lucide-react";
import { useMemo, useState } from "react";
import { Button } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import {
	DataTable,
	type dataTableFeatures,
	makeSelectColumn,
	SortableHeader,
} from "@/components/ui/data-table";
import { Input } from "@/components/ui/input";
import { ProgressBar, StatusPill } from "@/components/ui/status-pill";
import type { Beep } from "@/lib/api/beeps";
import { formatBeepScheduleTime } from "@/lib/beep-datetime";
import { beepRunAt } from "@/lib/beep-stats";
import { beepRunStatusLabel, beepStatusLabel } from "@/lib/i18n-labels";
import { runSuccessRate } from "@/lib/run-success-rate";
import { shortId } from "@/lib/short-id";
import { cn } from "@/lib/utils";
import * as m from "@/locale/paraglide/messages";

type FilterStatus = "all" | "active" | "firing" | "recurring" | "completed";

const columnHelper = createColumnHelper<typeof dataTableFeatures, Beep>();

function beepStatusTone(status: Beep["status"]) {
	switch (status) {
		case "active":
			return "emerald" as const;
		case "firing":
			return "amber" as const;
		case "cancelled":
			return "rose" as const;
		default:
			return "muted" as const;
	}
}

function formatScheduleLabel(beep: Beep) {
	if (beep.status === "completed") return m.beeps_schedule_completed();
	if (beep.kind === "recurring") return beep.cron ?? m.beeps_schedule();
	const nextRun = beepRunAt(beep);
	if (!nextRun) return m.common_em_dash();
	return formatBeepScheduleTime(nextRun, beep.timezone, "short");
}

function useBeepColumns(slug: string, variant: "compact" | "full") {
	return useMemo(() => {
		const fullColumns =
			variant === "full"
				? [
						columnHelper.accessor((row) => runSuccessRate(row.runs), {
							id: "run_success",
							header: ({ column }) => (
								<SortableHeader column={column} label={m.beeps_run_success()} />
							),
							cell: ({ row }) => (
								<ProgressBar value={runSuccessRate(row.original.runs)} />
							),
						}),
						columnHelper.accessor((row) => row.runs.length, {
							id: "runs",
							header: ({ column }) => (
								<SortableHeader column={column} label={m.beeps_runs()} />
							),
							cell: ({ row }) => {
								const beep = row.original;
								const lastRun = beep.runs[beep.runs.length - 1];
								return (
									<div className="flex flex-col gap-0.5 text-sm">
										<span className="tabular-nums text-foreground">
											{beep.runs.length}
										</span>
										{lastRun ? (
											<span className="text-[11px] text-muted-foreground capitalize">
												{m.beeps_last()}: {beepRunStatusLabel(lastRun.status)}
											</span>
										) : (
											<span className="text-[11px] text-muted-foreground">
												{m.common_em_dash()}
											</span>
										)}
									</div>
								);
							},
						}),
					]
				: [];

		return columnHelper.columns([
			makeSelectColumn(columnHelper),
			columnHelper.accessor("title", {
				id: "title",
				header: ({ column }) => (
					<SortableHeader column={column} label={m.term_beep_capitalized()} />
				),
				cell: ({ row }) => {
					const beep = row.original;
					return (
						<div className="flex min-w-48 flex-col gap-0.5">
							<span className="font-mono text-[11px] text-muted-foreground">
								#{shortId(beep.id)}
							</span>
							<Link
								to="/$account_slug/beeps/$beepId"
								params={{
									account_slug: slug,
									beepId: beep.id,
								}}
								className="font-medium text-foreground hover:text-primary"
								onClick={(event) => event.stopPropagation()}
								data-no-row-nav
							>
								{beep.title}
							</Link>
						</div>
					);
				},
			}),
			columnHelper.accessor((row) => row.beeper?.name ?? row.kind, {
				id: "source",
				header: m.beeps_source(),
				cell: ({ row }) => {
					const beep = row.original;
					if (beep.beeper) {
						return (
							<span className="inline-flex items-center gap-1.5 text-sm text-muted-foreground">
								<Activity className="size-3.5 text-primary" />
								{beep.beeper.name}
							</span>
						);
					}
					return (
						<span className="inline-flex items-center gap-1.5 text-sm text-muted-foreground">
							{beep.kind === "recurring" ? (
								<>
									<Repeat className="size-3.5" />
									{m.beeps_kind_recurring()}
								</>
							) : (
								<>
									<Clock className="size-3.5" />
									{m.beeps_kind_once()}
								</>
							)}
						</span>
					);
				},
			}),
			columnHelper.accessor("status", {
				id: "status",
				header: ({ column }) => (
					<SortableHeader column={column} label={m.common_status()} />
				),
				cell: ({ row }) => (
					<StatusPill
						label={beepStatusLabel(row.original.status)}
						tone={beepStatusTone(row.original.status)}
					/>
				),
			}),
			columnHelper.accessor((row) => beepRunAt(row)?.toString() ?? "", {
				id: "schedule",
				header: ({ column }) => (
					<SortableHeader column={column} label={m.beeps_schedule()} />
				),
				cell: ({ row }) => {
					const beep = row.original;
					return (
						<div className="flex flex-col gap-0.5 text-sm">
							<span className="tabular-nums text-foreground">
								{formatScheduleLabel(beep)}
							</span>
							{variant === "full" ? (
								<span className="text-[11px] text-muted-foreground">
									{beep.timezone}
								</span>
							) : null}
						</div>
					);
				},
			}),
			...fullColumns,
		]);
	}, [slug, variant]);
}

const FILTER_TABS: {
	id: FilterStatus;
	label: () => string;
}[] = [
	{ id: "all", label: m.beeps_filter_all },
	{ id: "active", label: m.beeps_filter_active },
	{ id: "firing", label: m.beeps_filter_firing },
	{ id: "recurring", label: m.beeps_filter_recurring },
	{ id: "completed", label: m.beeps_filter_completed },
];

export function BeepList({
	beeps,
	slug,
	variant = "full",
}: {
	beeps: Beep[];
	slug: string;
	variant?: "compact" | "full";
}) {
	const navigate = useNavigate();
	const [search, setSearch] = useState("");
	const [statusFilter, setStatusFilter] = useState<FilterStatus>("all");
	const columns = useBeepColumns(slug, variant);

	const filteredBeeps = useMemo(() => {
		return beeps.filter((beep) => {
			if (statusFilter === "active" && beep.status !== "active") return false;
			if (statusFilter === "firing" && beep.status !== "firing") return false;
			if (statusFilter === "completed" && beep.status !== "completed")
				return false;
			if (statusFilter === "recurring" && beep.kind !== "recurring")
				return false;

			if (search.trim()) {
				const query = search.toLowerCase();
				const matchTitle = beep.title.toLowerCase().includes(query);
				const matchBody = beep.body?.toLowerCase().includes(query);
				const matchBeeper = beep.beeper?.name.toLowerCase().includes(query);
				if (!matchTitle && !matchBody && !matchBeeper) return false;
			}

			return true;
		});
	}, [beeps, statusFilter, search]);

	const counts = useMemo(() => {
		return {
			all: beeps.length,
			active: beeps.filter((b) => b.status === "active").length,
			firing: beeps.filter((b) => b.status === "firing").length,
			recurring: beeps.filter((b) => b.kind === "recurring").length,
			completed: beeps.filter((b) => b.status === "completed").length,
		};
	}, [beeps]);

	if (beeps.length === 0) {
		return (
			<Card className="flex flex-col items-center justify-center p-8 text-center">
				<div className="flex size-12 items-center justify-center rounded-full bg-primary/10 text-primary">
					<Sparkles className="size-6" />
				</div>
				<h3 className="mt-3 font-heading text-base font-semibold">
					{m.beeps_empty_no_beeps()}
				</h3>
				<p className="mt-1 max-w-sm text-sm text-muted-foreground">
					{m.beeps_empty_create_hint()}
				</p>
			</Card>
		);
	}

	return (
		<div className="flex flex-col gap-4">
			{variant === "full" ? (
				<div className="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
					<div className="flex flex-nowrap items-center gap-1 overflow-x-auto rounded-lg border border-input bg-muted/30 p-1 [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
						{FILTER_TABS.map((tab) => (
							<Button
								key={tab.id}
								type="button"
								size="xs"
								variant={statusFilter === tab.id ? "default" : "ghost"}
								className={cn(
									"h-7 shrink-0 text-xs font-medium whitespace-nowrap transition-colors",
									statusFilter === tab.id
										? "bg-background text-foreground shadow-sm dark:bg-card dark:text-foreground"
										: "text-muted-foreground hover:text-foreground",
								)}
								onClick={() => setStatusFilter(tab.id)}
							>
								{tab.label()}
								<span
									className={cn(
										"ml-1 rounded-full px-1.5 py-0.2 text-[10px] tabular-nums",
										statusFilter === tab.id
											? "bg-primary/10 text-primary dark:bg-primary/20"
											: "bg-muted text-muted-foreground",
									)}
								>
									{counts[tab.id]}
								</span>
							</Button>
						))}
					</div>

					<div className="relative w-full sm:w-64">
						<Search className="absolute top-1/2 left-2.5 size-4 -translate-y-1/2 text-muted-foreground" />
						<Input
							value={search}
							onChange={(event) => setSearch(event.target.value)}
							placeholder={m.beeps_search_placeholder()}
							className="h-8.5 pl-8 text-xs"
						/>
					</div>
				</div>
			) : null}

			{filteredBeeps.length === 0 ? (
				<Card className="flex flex-col items-center justify-center p-8 text-center">
					<p className="text-sm font-medium text-muted-foreground">
						{m.beeps_filter_no_match()}
					</p>
					<Button
						type="button"
						variant="ghost"
						size="sm"
						className="mt-2"
						onClick={() => {
							setStatusFilter("all");
							setSearch("");
						}}
					>
						{m.beeps_clear_filters()}
					</Button>
				</Card>
			) : (
				<DataTable
					data={filteredBeeps}
					columns={columns}
					getRowId={(beep) => beep.id}
					emptyMessage={m.beeps_filter_no_match()}
					onRowClick={(beep) =>
						navigate({
							to: "/$account_slug/beeps/$beepId",
							params: { account_slug: slug, beepId: beep.id },
						})
					}
				/>
			)}
		</div>
	);
}
