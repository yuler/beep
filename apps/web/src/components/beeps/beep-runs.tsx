import { ChevronRight, Loader2 } from "lucide-react";
import { useCallback, useEffect, useRef, useState } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
	type BeepRun,
	fetchBeepRuns,
	type PaginationMeta,
} from "@/lib/api/beeps";
import { formatBeepScheduleTime } from "@/lib/beep-datetime";
import { beepRunStatusLabel } from "@/lib/i18n-labels";
import { m } from "@/locale/paraglide/messages";

const RUN_STATUS_VARIANT: Record<
	string,
	"default" | "secondary" | "outline" | "destructive"
> = {
	pending: "secondary",
	running: "default",
	succeeded: "outline",
	failed: "destructive",
	skipped: "secondary",
	expired: "destructive",
};

export function BeepRuns({
	runs: initialRuns,
	initialPagination,
	slug,
	beepId,
	timezone,
}: {
	runs: BeepRun[];
	initialPagination?: PaginationMeta;
	slug?: string;
	beepId?: string;
	timezone: string;
}) {
	const [runs, setRuns] = useState<BeepRun[]>(initialRuns);
	const [pagination, setPagination] = useState<PaginationMeta | undefined>(
		initialPagination,
	);
	const [isLoadingMore, setIsLoadingMore] = useState(false);
	const isLoadingMoreRef = useRef(false);
	const [isOpen, setIsOpen] = useState(true);
	const sentinelRef = useRef<HTMLDivElement | null>(null);

	useEffect(() => {
		setRuns(initialRuns);
		setPagination(initialPagination);
	}, [initialRuns, initialPagination]);

	const loadMore = useCallback(async () => {
		if (
			isLoadingMoreRef.current ||
			!pagination?.has_more ||
			!pagination.next_page ||
			!slug ||
			!beepId
		) {
			return;
		}

		isLoadingMoreRef.current = true;
		setIsLoadingMore(true);
		try {
			const res = await fetchBeepRuns(slug, beepId, {
				page: pagination.next_page,
			});
			setRuns((prev) => {
				const existingIds = new Set(prev.map((r) => r.id));
				const newUnique = res.runs.filter((r) => !existingIds.has(r.id));
				return [...prev, ...newUnique];
			});
			setPagination(res.pagination);
		} catch (err) {
			console.error("Failed to load more beep runs", err);
		} finally {
			isLoadingMoreRef.current = false;
			setIsLoadingMore(false);
		}
	}, [pagination, slug, beepId]);

	useEffect(() => {
		if (!isOpen || !pagination?.has_more) return;
		const node = sentinelRef.current;
		if (!node) return;

		const observer = new IntersectionObserver(
			(entries) => {
				if (entries[0]?.isIntersecting) {
					loadMore();
				}
			},
			{ rootMargin: "150px" },
		);

		observer.observe(node);
		return () => observer.disconnect();
	}, [isOpen, pagination?.has_more, loadMore]);

	return (
		<details
			open
			className="group/runs rounded-lg border bg-muted/20 text-sm"
			onToggle={(e) => setIsOpen(e.currentTarget.open)}
		>
			<summary className="flex cursor-pointer list-none items-center gap-2 px-3 py-2 marker:hidden [&::-webkit-details-marker]:hidden">
				<ChevronRight className="size-3.5 shrink-0 text-muted-foreground transition-transform group-open/runs:rotate-90" />
				<span className="font-medium">{m.beeps_execution_runs()}</span>
				<span className="text-muted-foreground">
					{pagination?.total_count ?? runs.length}
				</span>
			</summary>
			{runs.length === 0 ? (
				<p className="border-t px-3 py-2 text-xs text-muted-foreground">
					{m.beeps_no_runs()}
				</p>
			) : (
				<ul className="flex flex-col gap-2 border-t px-3 py-2">
					{runs.map((run) => (
						<li
							key={run.id}
							className="flex flex-col gap-2 rounded-md bg-background p-2.5 ring-1 ring-foreground/10"
						>
							<div className="flex flex-wrap items-center justify-between gap-2">
								<span className="tabular-nums text-xs text-muted-foreground">
									{formatBeepScheduleTime(run.scheduled_for, timezone)}
								</span>
								<div className="flex items-center gap-1.5">
									<Badge
										variant={RUN_STATUS_VARIANT[run.status] ?? "secondary"}
									>
										{m.beeps_delivery_status({
											status: beepRunStatusLabel(run.status),
										})}
									</Badge>
								</div>
							</div>

							{run.result && Object.keys(run.result).length > 0 ? (
								<pre className="max-h-36 overflow-auto rounded bg-muted/50 p-2 text-[11px] leading-snug whitespace-pre-wrap">
									{JSON.stringify(run.result, null, 2)}
								</pre>
							) : null}
						</li>
					))}
				</ul>
			)}
			{pagination?.has_more ? (
				<div ref={sentinelRef} className="flex justify-center border-t p-2">
					<Button
						type="button"
						variant="ghost"
						size="sm"
						disabled={isLoadingMore}
						onClick={loadMore}
						className="gap-2 text-xs"
					>
						{isLoadingMore ? (
							<>
								<Loader2 className="size-3.5 animate-spin" />
								{m.common_loading()}
							</>
						) : (
							m.common_load_more()
						)}
					</Button>
				</div>
			) : null}
		</details>
	);
}
