import { Link, useNavigate, useRouter } from "@tanstack/react-router";
import { ChevronLeft, ChevronRight, Eraser, Trash2 } from "lucide-react";
import { useEffect, useState } from "react";
import { confirm } from "@/components/confirm-dialog";
import {
	isActiveRunnerRun,
	RunnerRunStatusBadge,
} from "@/components/runners/runner-run-status";
import { Button, buttonVariants } from "@/components/ui/button";
import { ApiError } from "@/lib/api/client";
import {
	clearRunnerJobRuns,
	deleteRunnerJobRun,
	fetchRunnerJobRun,
	fetchRunnerJobRuns,
	type RunnerJob,
	type RunnerRun,
} from "@/lib/api/runner-jobs";
import { formatBeepScheduleTime } from "@/lib/beep-datetime";
import { translateError } from "@/lib/i18n-labels";
import { cn } from "@/lib/utils";
import { m } from "@/locale/paraglide/messages";

export function RunnerJobWorkspace({
	slug,
	runnerId,
	job,
	initialRuns,
	selectedRun: selectedRunProp,
	onError,
}: {
	slug: string;
	runnerId: string;
	job: RunnerJob;
	initialRuns: RunnerRun[];
	selectedRun: RunnerRun | null;
	onError: (message: string | null) => void;
}) {
	const router = useRouter();
	const navigate = useNavigate();
	const [runs, setRuns] = useState(initialRuns);
	const [selectedRun, setSelectedRun] = useState(selectedRunProp);

	useEffect(() => {
		setRuns(initialRuns);
	}, [initialRuns]);

	useEffect(() => {
		setSelectedRun(selectedRunProp);
	}, [selectedRunProp]);

	const selectedRunId = selectedRun?.id;
	const hasActiveRun = runs.some(isActiveRunnerRun);
	const selectedIsActive = selectedRun ? isActiveRunnerRun(selectedRun) : false;

	useEffect(() => {
		if (!hasActiveRun) return;
		let active = true;
		const timer = window.setInterval(() => {
			void fetchRunnerJobRuns(slug, runnerId, job.id).then((res) => {
				if (!active) return;
				setRuns(res.runs);
			});
		}, 2000);
		return () => {
			active = false;
			window.clearInterval(timer);
		};
	}, [hasActiveRun, slug, runnerId, job.id]);

	useEffect(() => {
		if (!selectedRunId || !selectedIsActive) return;
		let active = true;
		const timer = window.setInterval(() => {
			void fetchRunnerJobRun(slug, runnerId, job.id, selectedRunId)
				.then((updated) => {
					if (!active) return;
					setSelectedRun(updated);
					if (!isActiveRunnerRun(updated)) {
						void router.invalidate();
					}
				})
				.catch((err) => {
					if (!active) return;
					if (err instanceof ApiError && err.status === 404) {
						void navigate({
							to: "/$account_slug/runners/$runnerId/jobs/$jobId",
							params: {
								account_slug: slug,
								runnerId,
								jobId: job.id,
							},
						});
					}
				});
		}, 2000);
		return () => {
			active = false;
			window.clearInterval(timer);
		};
	}, [
		selectedIsActive,
		selectedRunId,
		slug,
		runnerId,
		job.id,
		router,
		navigate,
	]);

	async function goToRuns(nextRuns: RunnerRun[]) {
		const next = nextRuns[0];
		if (next) {
			await navigate({
				to: "/$account_slug/runners/$runnerId/jobs/$jobId/runs/$runId",
				params: {
					account_slug: slug,
					runnerId,
					jobId: job.id,
					runId: next.id,
				},
			});
		} else {
			await navigate({
				to: "/$account_slug/runners/$runnerId/jobs/$jobId",
				params: {
					account_slug: slug,
					runnerId,
					jobId: job.id,
				},
			});
		}
	}

	async function handleDeleteRun(run: RunnerRun) {
		if (!(await confirm(m.runners_runs_delete_confirm()))) {
			return;
		}
		onError(null);
		try {
			await deleteRunnerJobRun(slug, runnerId, job.id, run.id);
			const res = await fetchRunnerJobRuns(slug, runnerId, job.id);
			setRuns(res.runs);
			if (selectedRunId === run.id) {
				await goToRuns(res.runs);
			}
			await router.invalidate();
		} catch (err) {
			onError(
				err instanceof ApiError
					? err.message
					: translateError(err) || m.runners_runs_delete_failed(),
			);
		}
	}

	async function handleClearRuns() {
		if (!(await confirm(m.runners_runs_clear_confirm()))) {
			return;
		}
		onError(null);
		try {
			await clearRunnerJobRuns(slug, runnerId, job.id);
			const res = await fetchRunnerJobRuns(slug, runnerId, job.id);
			setRuns(res.runs);
			await navigate({
				to: "/$account_slug/runners/$runnerId/jobs/$jobId",
				params: {
					account_slug: slug,
					runnerId,
					jobId: job.id,
				},
			});
			await router.invalidate();
		} catch (err) {
			onError(
				err instanceof ApiError
					? err.message
					: translateError(err) || m.runners_runs_clear_failed(),
			);
		}
	}

	return (
		<div className="flex min-h-[24rem] flex-1 flex-col overflow-hidden rounded-md border border-line md:h-[calc(100dvh_-_14rem)] md:max-h-[calc(100dvh_-_14rem)] md:min-h-[28rem] md:flex-row">
			<aside
				className={cn(
					"flex min-h-0 w-full flex-col overflow-hidden md:w-[min(22rem,42%)] md:shrink-0 md:border-r md:border-line",
					selectedRunId ? "max-md:hidden" : "",
				)}
			>
				<div className="flex items-center justify-between gap-2 border-b border-line px-3 py-2.5">
					<h2 className="text-sm font-medium text-ink">
						{m.runners_runs_title()}
					</h2>
					{runs.length > 0 ? (
						<Button
							size="xs"
							variant="outline"
							onClick={() => void handleClearRuns()}
						>
							<Eraser data-icon="inline-start" />
							{m.runners_runs_clear()}
						</Button>
					) : null}
				</div>
				<div className="min-h-0 flex-1 overflow-auto p-2">
					{runs.length === 0 ? (
						<p className="px-2 py-8 text-center text-sm text-gray">
							{m.runners_runs_empty()}
						</p>
					) : (
						<ul className="flex flex-col gap-1">
							{runs.map((run) => {
								const active = isActiveRunnerRun(run);
								const selected = selectedRunId === run.id;
								return (
									<li key={run.id}>
										<div
											className={cn(
												"flex items-center gap-1 rounded-sm border pr-1",
												selected
													? "border-ink bg-wash"
													: "border-transparent hover:bg-wash",
											)}
										>
											<Link
												to="/$account_slug/runners/$runnerId/jobs/$jobId/runs/$runId"
												params={{
													account_slug: slug,
													runnerId,
													jobId: job.id,
													runId: run.id,
												}}
												className="flex min-w-0 flex-1 items-center justify-between gap-2 px-2.5 py-2 text-left"
											>
												<span className="truncate font-mono text-xs text-ink">
													{formatBeepScheduleTime(
														run.scheduled_for,
														job.timezone,
													)}
												</span>
												<span className="flex shrink-0 items-center gap-1">
													<RunnerRunStatusBadge run={run} />
													<ChevronRight className="size-3.5 text-gray md:hidden" />
												</span>
											</Link>
											<Button
												size="icon-xs"
												variant="destructive"
												disabled={active}
												onClick={() => void handleDeleteRun(run)}
												title={m.runners_runs_delete()}
												aria-label={m.runners_runs_delete()}
											>
												<Trash2 className="size-3.5" />
											</Button>
										</div>
									</li>
								);
							})}
						</ul>
					)}
				</div>
			</aside>

			<section
				className={cn(
					"flex min-h-0 min-w-0 flex-1 flex-col overflow-hidden bg-ground",
					selectedRunId ? "" : "max-md:hidden",
				)}
			>
				{selectedRun ? (
					<>
						<div className="flex items-center justify-between gap-2 border-b border-line px-3 py-2">
							<div className="flex min-w-0 items-center gap-2">
								<Link
									to="/$account_slug/runners/$runnerId/jobs/$jobId"
									params={{
										account_slug: slug,
										runnerId,
										jobId: job.id,
									}}
									className={cn(
										buttonVariants({ size: "xs", variant: "ghost" }),
										"md:hidden",
									)}
								>
									<ChevronLeft data-icon="inline-start" />
									{m.common_back()}
								</Link>
								<span className="truncate font-mono text-xs text-ink">
									{formatBeepScheduleTime(
										selectedRun.scheduled_for,
										job.timezone,
									)}
								</span>
								<RunnerRunStatusBadge run={selectedRun} />
							</div>
							<Button
								size="icon-xs"
								variant="destructive"
								disabled={selectedIsActive}
								onClick={() => void handleDeleteRun(selectedRun)}
								title={m.runners_runs_delete()}
								aria-label={m.runners_runs_delete()}
							>
								<Trash2 className="size-3.5" />
							</Button>
						</div>
						<pre className="min-h-0 flex-1 overflow-auto bg-wash p-3 font-mono text-[11px] whitespace-pre-wrap text-ink">
							{selectedRun.log ||
								selectedRun.log_preview ||
								m.runners_runs_no_log()}
						</pre>
					</>
				) : (
					<div className="flex h-full items-center justify-center p-6 text-sm text-gray">
						{runs.length === 0
							? m.runners_runs_empty()
							: m.runners_runs_select_run()}
					</div>
				)}
			</section>
		</div>
	);
}
