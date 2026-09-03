import { Link, useNavigate, useRouter } from "@tanstack/react-router";
import { createColumnHelper } from "@tanstack/react-table";
import { Edit } from "lucide-react";
import { useMemo, useState } from "react";
import { EditBeeperDialog } from "@/components/beepers/edit-beeper-dialog";
import { Badge } from "@/components/ui/badge";
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
import {
	beepStatusLabel,
	channelLabel,
	healthStatusLabel,
} from "@/lib/i18n-labels";
import type { NotificationChannel } from "@/lib/notification-channels";
import { runSuccessRate } from "@/lib/run-success-rate";
import { shortId } from "@/lib/short-id";
import { m } from "@/locale/paraglide/messages";

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
						<SortableHeader column={column} label={m.term_beeper()} />
					),
					cell: ({ row }) => {
						const beeper = row.original;
						return (
							<div className="flex min-w-0 max-w-md flex-col gap-0.5">
								<span className="font-mono text-[11px] text-muted-foreground">
									#{shortId(beeper.id)}
								</span>
								<Link
									to="/$account_slug/beepers/$beeperId"
									params={{
										account_slug: slug,
										beeperId: beeper.id,
									}}
									className="truncate font-medium text-foreground transition-colors hover:text-primary"
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
						<SortableHeader column={column} label={m.beepers_app()} />
					),
					cell: ({ row }) => (
						<span className="text-sm text-muted-foreground">
							{row.original.beeper_app?.name ?? m.common_em_dash()}
						</span>
					),
				}),
				columnHelper.accessor("status", {
					id: "status",
					header: ({ column }) => (
						<SortableHeader column={column} label={m.common_status()} />
					),
					cell: ({ row }) => (
						<StatusPill
							label={beepStatusLabel(row.original.status)}
							tone={beeperStatusTone(row.original.status)}
						/>
					),
				}),
				columnHelper.accessor((row) => beeperHealthLabel(row), {
					id: "health",
					header: ({ column }) => (
						<SortableHeader column={column} label={m.beepers_health()} />
					),
					cell: ({ row }) => {
						const beeper = row.original;
						return (
							<StatusPill
								label={healthStatusLabel(
									beeperHealthLabel(beeper),
								).toUpperCase()}
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
						<SortableHeader column={column} label={m.beeps_schedule()} />
					),
					cell: ({ row }) => (
						<span className="font-mono text-xs text-muted-foreground">
							{row.original.cron}
						</span>
					),
				}),
				columnHelper.accessor(
					(row) => row.notification_channels?.join(",") ?? "",
					{
						id: "channels",
						header: m.beeps_channels(),
						cell: ({ row }) => {
							const channels = row.original.notification_channels ?? [];
							if (channels.length === 0) {
								return (
									<span className="text-sm text-muted-foreground">
										{m.beepers_default_channels()}
									</span>
								);
							}
							return (
								<div className="flex flex-wrap gap-1">
									{channels.map((channel) => (
										<Badge
											key={channel}
											variant="outline"
											className="text-[11px] font-normal"
										>
											{channelLabel(channel as NotificationChannel)}
										</Badge>
									))}
								</div>
							);
						},
						enableSorting: false,
					},
				),
				columnHelper.accessor((row) => runSuccessRate(row.runs ?? []), {
					id: "run_success",
					header: ({ column }) => (
						<SortableHeader column={column} label={m.beepers_run_success()} />
					),
					cell: ({ row }) => (
						<ProgressBar value={runSuccessRate(row.original.runs ?? [])} />
					),
				}),
				columnHelper.accessor((row) => row.runs?.length ?? 0, {
					id: "runs",
					header: ({ column }) => (
						<SortableHeader column={column} label={m.beepers_runs()} />
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
						<SortableHeader column={column} label={m.beepers_next_run()} />
					),
					cell: ({ row }) => {
						const beeper = row.original;
						if (!beeper.next_run_at) {
							return (
								<span className="text-sm text-muted-foreground">
									{m.common_em_dash()}
								</span>
							);
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
					header: () => <span className="sr-only">{m.common_actions()}</span>,
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
									{m.beepers_edit()}
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
			{/* Mobile Card List View (< md) */}
			<div className="flex flex-col gap-3 md:hidden">
				{beepers.length === 0 ? (
					<div className="rounded-lg border bg-card p-6 text-center text-sm text-muted-foreground">
						{m.beepers_empty_no_beepers()}
					</div>
				) : (
					beepers.map((beeper) => {
						const channels = beeper.notification_channels ?? [];
						const successRate = runSuccessRate(beeper.runs ?? []);

						return (
							<div
								key={beeper.id}
								className="relative flex flex-col gap-3 rounded-xl border border-border bg-card p-4 text-left transition-all hover:border-foreground/20 hover:shadow-xs"
							>
								<Link
									to="/$account_slug/beepers/$beeperId"
									params={{ account_slug: slug, beeperId: beeper.id }}
									className="absolute inset-0 z-0 rounded-xl"
									aria-label={beeper.title}
								/>
								<div className="relative z-10 flex items-start justify-between gap-3 pointer-events-none">
									<div className="flex min-w-0 flex-col gap-0.5">
										<span className="font-mono text-[11px] text-muted-foreground">
											#{shortId(beeper.id)}
										</span>
										<span className="truncate text-base font-semibold text-foreground">
											{beeper.title}
										</span>
									</div>
									<div className="flex shrink-0 items-center gap-1.5">
										<StatusPill
											label={healthStatusLabel(
												beeperHealthLabel(beeper),
											).toUpperCase()}
											tone={
												beeperHealthIsDestructive(beeper)
													? "rose"
													: healthTone(beeper)
											}
										/>
										<StatusPill
											label={beepStatusLabel(beeper.status)}
											tone={beeperStatusTone(beeper.status)}
										/>
									</div>
								</div>

								<div className="relative z-10 grid grid-cols-2 gap-2 border-y border-border/50 py-2.5 text-xs text-muted-foreground pointer-events-none">
									<div>
										<span className="block text-[11px] text-muted-foreground/80">
											{m.beepers_app()}
										</span>
										<span className="font-medium text-foreground">
											{beeper.beeper_app?.name ?? m.common_em_dash()}
										</span>
									</div>
									<div>
										<span className="block text-[11px] text-muted-foreground/80">
											{m.beeps_schedule()}
										</span>
										<span className="font-mono font-medium text-foreground">
											{beeper.cron}
										</span>
									</div>
									<div>
										<span className="block text-[11px] text-muted-foreground/80">
											{m.beepers_next_run()}
										</span>
										<span className="text-foreground">
											{beeper.next_run_at
												? formatBeepScheduleTime(
														beeper.next_run_at,
														beeper.timezone,
														"short",
													)
												: m.common_em_dash()}
										</span>
									</div>
									<div>
										<span className="block text-[11px] text-muted-foreground/80">
											{m.beepers_run_success()} ({beeper.runs?.length ?? 0})
										</span>
										<div className="mt-1">
											<ProgressBar value={successRate} />
										</div>
									</div>
								</div>

								<div className="relative z-10 flex items-center justify-between gap-2 pt-0.5">
									<div className="flex min-w-0 flex-wrap gap-1 pointer-events-none">
										{channels.length > 0 ? (
											channels.map((channel) => (
												<Badge
													key={channel}
													variant="outline"
													className="px-1.5 py-0 text-[10px] font-normal"
												>
													{channelLabel(channel as NotificationChannel)}
												</Badge>
											))
										) : (
											<span className="text-[11px] text-muted-foreground">
												{m.beepers_default_channels()}
											</span>
										)}
									</div>
									<Button
										type="button"
										variant="ghost"
										size="xs"
										className="relative z-10 h-7 shrink-0 px-2 text-xs text-muted-foreground hover:text-foreground"
										onClick={(e) => {
											e.preventDefault();
											e.stopPropagation();
											setEditingBeeper(beeper);
										}}
									>
										<Edit className="size-3.5" data-icon="inline-start" />
										{m.beepers_edit()}
									</Button>
								</div>
							</div>
						);
					})
				)}
			</div>

			{/* Desktop Table View (>= md) */}
			<div className="hidden md:block">
				<DataTable
					data={beepers}
					columns={columns}
					getRowId={(beeper) => beeper.id}
					emptyMessage={m.beepers_empty_no_beepers()}
					onRowClick={(beeper) =>
						navigate({
							to: "/$account_slug/beepers/$beeperId",
							params: { account_slug: slug, beeperId: beeper.id },
						})
					}
				/>
			</div>

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
