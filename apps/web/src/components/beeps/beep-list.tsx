import { Link, useNavigate } from "@tanstack/react-router";
import {
	createColumnHelper,
	type OnChangeFn,
	type SortingState,
} from "@tanstack/react-table";
import {
	Activity,
	Clock,
	Loader2,
	Plus,
	Repeat,
	Search,
	Sparkles,
} from "lucide-react";
import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Badge } from "@/components/ui/badge";
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
import {
	type Beep,
	type BeepSortDir,
	type BeepSortField,
	type BeepStatsData,
	beepSortQuery,
	fetchBeeps,
	isBeepSortField,
	type PaginationMeta,
} from "@/lib/api/beeps";
import { formatBeepScheduleTime } from "@/lib/beep-datetime";
import {
	beepRunStatusLabel,
	beepStatusLabel,
	channelLabel,
} from "@/lib/i18n-labels";
import { beepRunCount, beepRunSuccessRate } from "@/lib/run-success-rate";
import { cn } from "@/lib/utils";
import { m } from "@/locale/paraglide/messages";

export type FilterStatus =
	| "all"
	| "active"
	| "firing"
	| "recurring"
	| "completed";

const columnHelper = createColumnHelper<typeof dataTableFeatures, Beep>();

function beepStatusTone(status: Beep["status"]) {
	switch (status) {
		case "active":
			return "emerald" as const;
		case "firing":
			return "amber" as const;
		default:
			return "muted" as const;
	}
}

function formatLastRunAt(beep: Beep) {
	if (!beep.last_run_at) return null;
	return formatBeepScheduleTime(beep.last_run_at, beep.timezone, "short");
}

function LastRunCaption({ beep }: { beep: Beep }) {
	const lastRun = beep.runs?.[0];
	const lastAt = formatLastRunAt(beep);
	if (!lastRun && !lastAt) return m.common_em_dash();

	return (
		<>
			{m.beeps_last_run_at()}:{" "}
			<span className="tabular-nums">{lastAt ?? m.common_em_dash()}</span>
			{lastRun ? (
				<span className="capitalize">
					{" "}
					· {beepRunStatusLabel(lastRun.status)}
				</span>
			) : null}
		</>
	);
}

function formatChannel(channel: string) {
	if (channel === "email" || channel === "web_push" || channel === "cli") {
		return channelLabel(channel);
	}
	return channel;
}

function BeepIdText({ id, className }: { id: string; className?: string }) {
	return (
		<span
			className={cn(
				"font-mono text-[11px] text-muted-foreground select-all whitespace-nowrap",
				className,
			)}
			title={id}
		>
			#{id}
		</span>
	);
}

function useBeepColumns(slug: string, variant: "compact" | "full") {
	return useMemo(() => {
		const sortable = variant === "full";
		const fullColumns =
			variant === "full"
				? [
						columnHelper.accessor((row) => row.notification_channels, {
							id: "channels",
							enableSorting: false,
							header: m.beeps_channels(),
							cell: ({ row }) => {
								const channels = row.original.notification_channels ?? [];
								if (channels.length === 0) {
									return (
										<span className="text-xs text-muted-foreground">
											{m.beeps_channels_none()}
										</span>
									);
								}
								return (
									<div className="flex flex-wrap items-center gap-1">
										{channels.map((channel) => (
											<Badge
												key={channel}
												variant="outline"
												className="px-1.5 py-0 text-[10px] font-normal"
											>
												{formatChannel(channel)}
											</Badge>
										))}
									</div>
								);
							},
						}),
						columnHelper.accessor((row) => beepRunCount(row), {
							id: "runs",
							enableSorting: false,
							header: m.beeps_runs(),
							cell: ({ row }) => (
								<div className="flex min-w-44 flex-col gap-1">
									<div className="flex items-center gap-2">
										<span className="text-sm tabular-nums text-foreground">
											{beepRunCount(row.original)}
										</span>
										<ProgressBar value={beepRunSuccessRate(row.original)} />
									</div>
									<span className="text-[11px] text-muted-foreground">
										<LastRunCaption beep={row.original} />
									</span>
								</div>
							),
						}),
					]
				: [];

		return columnHelper.columns([
			makeSelectColumn(columnHelper),
			columnHelper.accessor("title", {
				id: "title",
				enableSorting: sortable,
				meta: { className: "w-52 min-w-52 max-w-52" },
				header: sortable
					? ({ column }) => (
							<SortableHeader
								column={column}
								label={m.term_beep_capitalized()}
							/>
						)
					: m.term_beep_capitalized(),
				cell: ({ row }) => {
					const beep = row.original;
					return (
						<div className="flex min-w-0 flex-col gap-0.5">
							<BeepIdText id={beep.id} />
							<Link
								to="/$account_slug/beeps/$beepId"
								params={{
									account_slug: slug,
									beepId: beep.id,
								}}
								className="truncate font-medium text-foreground transition-colors hover:text-primary"
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
				enableSorting: false,
				meta: { className: "min-w-56" },
				header: m.beeps_source(),
				cell: ({ row }) => {
					const beep = row.original;
					if (beep.beeper) {
						return (
							<span className="inline-flex max-w-full items-center gap-1.5 text-sm text-muted-foreground">
								<Activity className="size-3.5 shrink-0 text-primary" />
								<span className="truncate">{beep.beeper.name}</span>
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
				enableSorting: sortable,
				header: sortable
					? ({ column }) => (
							<SortableHeader column={column} label={m.common_status()} />
						)
					: m.common_status(),
				cell: ({ row }) => (
					<StatusPill
						label={beepStatusLabel(row.original.status)}
						tone={beepStatusTone(row.original.status)}
					/>
				),
			}),
			...fullColumns,
			columnHelper.accessor("created_at", {
				id: "created_at",
				enableSorting: sortable,
				meta: { className: "min-w-36 whitespace-nowrap" },
				header: sortable
					? ({ column }) => (
							<SortableHeader column={column} label={m.common_created()} />
						)
					: m.common_created(),
				cell: ({ row }) => {
					const beep = row.original;
					return (
						<span className="whitespace-nowrap text-xs tabular-nums text-muted-foreground">
							{formatBeepScheduleTime(beep.created_at, beep.timezone, "short")}
						</span>
					);
				},
			}),
		]);
	}, [slug, variant]);
}

const FILTER_TABS: {
	id: FilterStatus;
	label: () => string;
}[] = [
	{ id: "active", label: m.beeps_filter_active },
	{ id: "firing", label: m.beeps_filter_firing },
	{ id: "recurring", label: m.beeps_filter_recurring },
	{ id: "completed", label: m.beeps_filter_completed },
	{ id: "all", label: m.beeps_filter_all },
];

export function getFilterOptions(filter: FilterStatus) {
	if (filter === "recurring") return { kind: "recurring" };
	if (filter !== "all") return { status: filter };
	return {};
}

export function BeepList({
	beeps: initialBeeps,
	initialPagination,
	stats,
	slug,
	variant = "full",
	onCreateClick,
	currentTab,
	onTabChange,
	searchQuery = "",
	onSearchChange,
	sort: sortField = "created_at",
	dir: sortDir = "desc",
	onSortChange,
}: {
	beeps: Beep[];
	initialPagination?: PaginationMeta;
	stats?: BeepStatsData;
	slug: string;
	variant?: "compact" | "full";
	onCreateClick?: () => void;
	currentTab?: FilterStatus;
	onTabChange?: (tab: FilterStatus) => void;
	searchQuery?: string;
	onSearchChange?: (q: string) => void;
	sort?: BeepSortField;
	dir?: BeepSortDir;
	onSortChange?: (sort: BeepSortField, dir: BeepSortDir) => void;
}) {
	const navigate = useNavigate();
	const [items, setItems] = useState<Beep[]>(initialBeeps);
	const [pagination, setPagination] = useState<PaginationMeta | undefined>(
		initialPagination,
	);
	const [isLoadingMore, setIsLoadingMore] = useState(false);
	const [isFiltering, setIsFiltering] = useState(false);
	const isLoadingMoreRef = useRef(false);
	const filterRequestRef = useRef(0);
	const sentinelRef = useRef<HTMLDivElement | null>(null);

	const [internalStatusFilter, setInternalStatusFilter] =
		useState<FilterStatus>("active");
	const statusFilter = currentTab ?? internalStatusFilter;
	const sortQuery = useMemo(
		() => beepSortQuery(sortField, sortDir),
		[sortField, sortDir],
	);
	const sorting: SortingState = useMemo(
		() => [{ id: sortField, desc: sortDir === "desc" }],
		[sortField, sortDir],
	);

	const inputRef = useRef<HTMLInputElement>(null);
	const lastDispatchedRef = useRef(searchQuery);
	const debounceTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);
	const [searchInput, setSearchInput] = useState(searchQuery);

	useEffect(() => {
		// If searchQuery matches what we just dispatched, don't overwrite current input
		if (searchQuery === lastDispatchedRef.current) {
			return;
		}
		// If the user is currently typing/focusing in the input, don't overwrite with stale server echo
		if (document.activeElement === inputRef.current && searchQuery !== "") {
			return;
		}
		setSearchInput(searchQuery);
		lastDispatchedRef.current = searchQuery;
	}, [searchQuery]);

	useEffect(() => {
		setItems(initialBeeps);
		setPagination(initialPagination);
	}, [initialBeeps, initialPagination]);

	const dispatchSearch = useCallback(
		(trimmed: string) => {
			lastDispatchedRef.current = trimmed;
			if (onSearchChange) {
				onSearchChange(trimmed);
				return;
			}

			const requestId = ++filterRequestRef.current;
			setIsFiltering(true);
			fetchBeeps(slug, {
				...getFilterOptions(statusFilter),
				q: trimmed || undefined,
				...sortQuery,
			})
				.then((res) => {
					if (filterRequestRef.current === requestId) {
						setItems(res.beeps);
						setPagination(res.pagination);
					}
				})
				.catch((err) => {
					console.error("Failed to search beeps", err);
				})
				.finally(() => {
					if (filterRequestRef.current === requestId) {
						setIsFiltering(false);
					}
				});
		},
		[onSearchChange, slug, statusFilter, sortQuery],
	);

	useEffect(() => {
		const timer = setTimeout(() => {
			const trimmed = searchInput.trim();
			if (trimmed !== lastDispatchedRef.current) {
				dispatchSearch(trimmed);
			}
		}, 300);
		debounceTimerRef.current = timer;

		return () => clearTimeout(timer);
	}, [searchInput, dispatchSearch]);

	const handleStatusFilterChange = useCallback(
		(nextFilter: FilterStatus) => {
			if (onTabChange) {
				onTabChange(nextFilter);
				return;
			}
			setInternalStatusFilter(nextFilter);

			const requestId = ++filterRequestRef.current;
			setIsFiltering(true);
			fetchBeeps(slug, {
				...getFilterOptions(nextFilter),
				q: lastDispatchedRef.current.trim() || undefined,
				...sortQuery,
			})
				.then((res) => {
					if (filterRequestRef.current === requestId) {
						setItems(res.beeps);
						setPagination(res.pagination);
					}
				})
				.catch((err) => {
					console.error("Failed to filter beeps", err);
				})
				.finally(() => {
					if (filterRequestRef.current === requestId) {
						setIsFiltering(false);
					}
				});
		},
		[slug, onTabChange, sortQuery],
	);

	const loadMore = useCallback(async () => {
		if (
			isLoadingMoreRef.current ||
			!pagination?.has_more ||
			!pagination.next_page
		) {
			return;
		}
		isLoadingMoreRef.current = true;
		setIsLoadingMore(true);
		try {
			const res = await fetchBeeps(slug, {
				page: pagination.next_page,
				...getFilterOptions(statusFilter),
				q: lastDispatchedRef.current.trim() || undefined,
				...sortQuery,
			});
			setItems((prev) => {
				const existingIds = new Set(prev.map((b) => b.id));
				const newUnique = res.beeps.filter((b) => !existingIds.has(b.id));
				return [...prev, ...newUnique];
			});
			setPagination(res.pagination);
		} catch (err) {
			console.error("Failed to load more beeps", err);
		} finally {
			isLoadingMoreRef.current = false;
			setIsLoadingMore(false);
		}
	}, [pagination, slug, statusFilter, sortQuery]);

	const handleSortingChange = useCallback<OnChangeFn<SortingState>>(
		(updater) => {
			const next = typeof updater === "function" ? updater(sorting) : updater;
			const first = next[0];
			if (!first || !isBeepSortField(first.id)) return;
			const nextDir: BeepSortDir = first.desc ? "desc" : "asc";
			if (onSortChange) {
				onSortChange(first.id, nextDir);
				return;
			}

			const requestId = ++filterRequestRef.current;
			setIsFiltering(true);
			fetchBeeps(slug, {
				...getFilterOptions(statusFilter),
				q: lastDispatchedRef.current.trim() || undefined,
				...beepSortQuery(first.id, nextDir),
			})
				.then((res) => {
					if (filterRequestRef.current === requestId) {
						setItems(res.beeps);
						setPagination(res.pagination);
					}
				})
				.catch((err) => {
					console.error("Failed to sort beeps", err);
				})
				.finally(() => {
					if (filterRequestRef.current === requestId) {
						setIsFiltering(false);
					}
				});
		},
		[sorting, onSortChange, slug, statusFilter],
	);

	useEffect(() => {
		if (!pagination?.has_more) return;
		const node = sentinelRef.current;
		if (!node) return;

		const observer = new IntersectionObserver(
			(entries) => {
				if (entries[0]?.isIntersecting) {
					loadMore();
				}
			},
			{ rootMargin: "200px" },
		);

		observer.observe(node);
		return () => observer.disconnect();
	}, [pagination?.has_more, loadMore]);

	const columns = useBeepColumns(slug, variant);
	const filteredBeeps = items;

	const counts = useMemo(() => {
		const next = {
			all:
				stats?.all ??
				(statusFilter === "all" ? pagination?.total_count : undefined) ??
				items.length,
			active:
				stats?.active ??
				(statusFilter === "active" ? pagination?.total_count : undefined) ??
				items.filter((b) => b.status === "active").length,
			firing:
				stats?.firing ??
				(statusFilter === "firing" ? pagination?.total_count : undefined) ??
				items.filter((b) => b.status === "firing").length,
			recurring:
				stats?.recurring ??
				(statusFilter === "recurring" ? pagination?.total_count : undefined) ??
				items.filter((b) => b.kind === "recurring").length,
			completed:
				stats?.completed ??
				(statusFilter === "completed" ? pagination?.total_count : undefined) ??
				items.filter((b) => b.status === "completed").length,
		};
		if (searchQuery.trim() && pagination?.total_count != null) {
			next[statusFilter] = pagination.total_count;
		}
		return next;
	}, [stats, pagination, items, statusFilter, searchQuery]);

	const totalCount =
		stats?.all ??
		(statusFilter === "all" ? pagination?.total_count : undefined) ??
		items.length;

	if (totalCount === 0 && !isFiltering) {
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
				{onCreateClick ? (
					<Button
						type="button"
						size="sm"
						className="mt-4 gap-2"
						onClick={onCreateClick}
					>
						<Plus className="size-4" />
						{m.beeps_create_beep()}
					</Button>
				) : null}
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
								onClick={() => handleStatusFilterChange(tab.id)}
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
							ref={inputRef}
							value={searchInput}
							onChange={(event) => setSearchInput(event.target.value)}
							onKeyDown={(event) => {
								if (event.key === "Enter") {
									event.preventDefault();
									if (debounceTimerRef.current) {
										clearTimeout(debounceTimerRef.current);
										debounceTimerRef.current = null;
									}
									const trimmed = searchInput.trim();
									if (trimmed !== lastDispatchedRef.current) {
										dispatchSearch(trimmed);
									}
								}
							}}
							placeholder={m.beeps_search_placeholder()}
							className="h-8.5 pl-8 pr-8 text-xs"
						/>
						{isFiltering || searchInput.trim() !== searchQuery ? (
							<Loader2 className="absolute top-1/2 right-2.5 size-3.5 -translate-y-1/2 animate-spin text-muted-foreground" />
						) : null}
					</div>
				</div>
			) : null}

			{isFiltering ? (
				<div className="flex items-center justify-center py-12">
					<Loader2 className="size-6 animate-spin text-muted-foreground" />
				</div>
			) : filteredBeeps.length === 0 ? (
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
							setSearchInput("");
							lastDispatchedRef.current = "";
							if (onSearchChange) onSearchChange("");
							handleStatusFilterChange("all");
						}}
					>
						{m.beeps_clear_filters()}
					</Button>
				</Card>
			) : (
				<>
					{/* Mobile Card List View (< md) */}
					<div className="flex flex-col gap-3 md:hidden">
						{filteredBeeps.map((beep) => {
							const successRate = beepRunSuccessRate(beep);
							const totalRuns = beepRunCount(beep);

							return (
								<div
									key={beep.id}
									className="relative flex flex-col gap-3 rounded-xl border border-border bg-card p-4 text-left transition-all hover:border-foreground/20 hover:shadow-xs"
								>
									<Link
										to="/$account_slug/beeps/$beepId"
										params={{ account_slug: slug, beepId: beep.id }}
										className="absolute inset-0 z-0 rounded-xl"
										aria-label={beep.title}
									/>
									<div className="relative z-10 flex items-start justify-between gap-3 pointer-events-none">
										<div className="flex min-w-0 flex-col gap-0.5">
											<BeepIdText
												id={beep.id}
												className="pointer-events-auto"
											/>
											<span className="truncate text-base font-semibold text-foreground">
												{beep.title}
											</span>
										</div>
										<StatusPill
											label={beepStatusLabel(beep.status)}
											tone={beepStatusTone(beep.status)}
										/>
									</div>

									<div className="relative z-10 grid grid-cols-2 gap-2 border-y border-border/50 py-2.5 text-xs text-muted-foreground pointer-events-none">
										<div>
											<span className="block text-[11px] text-muted-foreground/80">
												{m.beeps_source()}
											</span>
											<span className="mt-0.5 inline-flex items-center gap-1 font-medium text-foreground">
												{beep.beeper ? (
													<>
														<Activity className="size-3 text-primary" />
														<span className="truncate">{beep.beeper.name}</span>
													</>
												) : beep.kind === "recurring" ? (
													<>
														<Repeat className="size-3" />
														<span>{m.beeps_kind_recurring()}</span>
													</>
												) : (
													<>
														<Clock className="size-3" />
														<span>{m.beeps_kind_once()}</span>
													</>
												)}
											</span>
										</div>
										{variant === "full" ? (
											<>
												<div>
													<span className="block text-[11px] text-muted-foreground/80">
														{m.beeps_channels()}
													</span>
													<div className="mt-0.5 flex flex-wrap gap-1">
														{beep.notification_channels?.length > 0 ? (
															beep.notification_channels.map((ch) => (
																<Badge
																	key={ch}
																	variant="outline"
																	className="px-1 py-0 text-[10px]"
																>
																	{formatChannel(ch)}
																</Badge>
															))
														) : (
															<span>{m.beeps_channels_none()}</span>
														)}
													</div>
												</div>
												<div>
													<span className="block text-[11px] text-muted-foreground/80">
														{m.beeps_runs()} ({totalRuns})
													</span>
													<div className="mt-1">
														<ProgressBar value={successRate} />
													</div>
													<span className="mt-0.5 block text-[11px] text-foreground">
														<LastRunCaption beep={beep} />
													</span>
												</div>
												<div>
													<span className="block text-[11px] text-muted-foreground/80">
														{m.common_created()}
													</span>
													<span className="mt-0.5 block font-medium text-foreground">
														{formatBeepScheduleTime(
															beep.created_at,
															beep.timezone,
															"short",
														)}
													</span>
												</div>
											</>
										) : null}
									</div>
								</div>
							);
						})}
					</div>

					{/* Desktop Table View (>= md) */}
					<div className="hidden min-w-0 md:block">
						<DataTable
							data={filteredBeeps}
							columns={columns}
							getRowId={(beep) => beep.id}
							emptyMessage={m.beeps_filter_no_match()}
							manualSorting={variant === "full"}
							sorting={sorting}
							onSortingChange={handleSortingChange}
							onRowClick={(beep) =>
								navigate({
									to: "/$account_slug/beeps/$beepId",
									params: { account_slug: slug, beepId: beep.id },
								})
							}
						/>
					</div>

					{pagination?.has_more ? (
						<div ref={sentinelRef} className="flex justify-center py-4">
							<Button
								type="button"
								variant="outline"
								size="sm"
								disabled={isLoadingMore}
								onClick={loadMore}
								className="gap-2"
							>
								{isLoadingMore ? (
									<>
										<Loader2 className="size-4 animate-spin" />
										{m.common_loading()}
									</>
								) : (
									m.common_load_more()
								)}
							</Button>
						</div>
					) : null}
				</>
			)}
		</div>
	);
}
