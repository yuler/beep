import {
	createFileRoute,
	getRouteApi,
	notFound,
	useRouter,
} from "@tanstack/react-router";
import {
	Check,
	Clock,
	Copy,
	Edit,
	Pause,
	Play,
	Plus,
	Server,
	Trash2,
} from "lucide-react";
import { useEffect, useState } from "react";
import { DashboardHeader } from "@/components/dashboard/dashboard-header";
import { RunnerJobFormDialog } from "@/components/runners/runner-job-form-dialog";
import {
	formatRunnerLastSeen,
	getRunnerStatusBadge,
} from "@/components/runners/runner-list";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { ApiError } from "@/lib/api/client";
import {
	deleteRunnerJob,
	fetchRunnerJobRun,
	fetchRunnerJobRuns,
	fetchRunnerJobs,
	pauseRunnerJob,
	type RunnerJob,
	type RunnerRun,
	resumeRunnerJob,
	triggerRunnerJobRun,
} from "@/lib/api/runner-jobs";
import { fetchRunner } from "@/lib/api/runners";
import { withAuthRedirects } from "@/lib/auth/guards";
import { translateError } from "@/lib/i18n-labels";
import { m } from "@/locale/paraglide/messages";

const accountRoute = getRouteApi("/$account_slug");

export const Route = createFileRoute("/$account_slug/runners_/$runnerId")({
	loader: withAuthRedirects(async ({ params }) => {
		const slug = params?.account_slug ?? "";
		const runnerId = params?.runnerId ?? "";
		try {
			const [{ runner }, { jobs }] = await Promise.all([
				fetchRunner(slug, runnerId),
				fetchRunnerJobs(slug, runnerId),
			]);
			return { runner, jobs };
		} catch (err) {
			if (err instanceof ApiError && err.status === 404) {
				throw notFound();
			}
			throw err;
		}
	}),
	component: RunnerDetailPage,
});

function RunnerDetailPage() {
	const { account_slug: slug } = accountRoute.useParams();
	const router = useRouter();
	const { runner, jobs } = Route.useLoaderData();
	const [selectedJobId, setSelectedJobId] = useState<string | null>(
		jobs[0]?.id ?? null,
	);
	const [runs, setRuns] = useState<RunnerRun[]>([]);
	const [selectedRun, setSelectedRun] = useState<RunnerRun | null>(null);
	const [error, setError] = useState<string | null>(null);
	const [isJobDialogOpen, setIsJobDialogOpen] = useState(false);
	const [editingJob, setEditingJob] = useState<RunnerJob | null>(null);
	const [copiedJobId, setCopiedJobId] = useState<string | null>(null);

	const selectedJob = jobs.find((job) => job.id === selectedJobId) ?? null;

	function handleCopyJobId(id: string, event: React.MouseEvent) {
		event.stopPropagation();
		void navigator.clipboard.writeText(id);
		setCopiedJobId(id);
		setTimeout(() => setCopiedJobId(null), 2000);
	}

	function handleOpenCreateJob() {
		setEditingJob(null);
		setIsJobDialogOpen(true);
	}

	function handleOpenEditJob(job: RunnerJob) {
		setEditingJob(job);
		setIsJobDialogOpen(true);
	}

	useEffect(() => {
		if (!selectedJob) {
			setRuns([]);
			setSelectedRun(null);
			return;
		}
		void fetchRunnerJobRuns(slug, runner.id, selectedJob.id)
			.then((res) => {
				setRuns(res.runs);
				setSelectedRun(res.runs[0] ?? null);
			})
			.catch((err) => {
				setError(
					err instanceof ApiError
						? err.message
						: translateError(err) || m.runners_jobs_load_failed(),
				);
			});
	}, [selectedJob, slug, runner.id]);

	useEffect(() => {
		if (!selectedJob || !selectedRun || selectedRun.status !== "running") {
			return;
		}
		const timer = window.setInterval(() => {
			void fetchRunnerJobRun(
				slug,
				runner.id,
				selectedJob.id,
				selectedRun.id,
			).then(setSelectedRun);
		}, 2000);
		return () => window.clearInterval(timer);
	}, [selectedJob, selectedRun, slug, runner.id]);

	async function handleTrigger(job: RunnerJob) {
		setError(null);
		try {
			const run = await triggerRunnerJobRun(slug, runner.id, job.id);
			const res = await fetchRunnerJobRuns(slug, runner.id, job.id);
			setRuns(res.runs);
			setSelectedRun(run);
			await router.invalidate();
		} catch (err) {
			setError(
				err instanceof ApiError
					? err.message
					: translateError(err) || m.runners_jobs_trigger_failed(),
			);
		}
	}

	async function handleToggle(job: RunnerJob) {
		setError(null);
		try {
			if (job.status === "paused") {
				await resumeRunnerJob(slug, runner.id, job.id);
			} else {
				await pauseRunnerJob(slug, runner.id, job.id);
			}
			await router.invalidate();
		} catch (err) {
			setError(
				err instanceof ApiError
					? err.message
					: translateError(err) || m.runners_jobs_status_failed(),
			);
		}
	}

	async function handleDelete(job: RunnerJob) {
		if (!window.confirm(m.runners_jobs_delete_confirm({ name: job.name }))) {
			return;
		}
		setError(null);
		try {
			await deleteRunnerJob(slug, runner.id, job.id);
			setSelectedJobId(null);
			await router.invalidate();
		} catch (err) {
			setError(
				err instanceof ApiError
					? err.message
					: translateError(err) || m.runners_jobs_delete_failed(),
			);
		}
	}

	return (
		<>
			<DashboardHeader
				breadcrumbs={[
					{
						label: m.nav_home(),
						to: "/$account_slug",
						params: { account_slug: slug },
					},
					{
						label: m.nav_runners(),
						to: "/$account_slug/runners",
						params: { account_slug: slug },
					},
					{ label: runner.name, isCurrentPage: true },
				]}
			/>

			<div className="flex flex-1 flex-col gap-6 p-4 md:p-6">
				<div className="flex flex-wrap items-center justify-between gap-4">
					<div className="flex items-center gap-3">
						<div className="flex size-10 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary">
							<Server className="size-5" />
						</div>
						<div className="flex flex-col gap-1">
							<div className="flex items-center gap-2.5">
								<h1 className="font-heading text-2xl font-bold tracking-tight">
									{runner.name}
								</h1>
								{getRunnerStatusBadge(runner)}
							</div>
							<div className="flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-muted-foreground font-mono">
								<span>{runner.token_prefix}••••</span>
								{runner.hostname ? <span>· {runner.hostname}</span> : null}
								{runner.version ? <span>· v{runner.version}</span> : null}
								{runner.os && runner.arch ? (
									<span>
										· {runner.os}/{runner.arch}
									</span>
								) : null}
								<span className="flex items-center gap-1 font-sans">
									<Clock className="size-3 shrink-0" />
									{formatRunnerLastSeen(runner.last_seen_at)}
								</span>
							</div>
						</div>
					</div>

					<Button size="sm" onClick={handleOpenCreateJob}>
						<Plus data-icon="inline-start" />
						{m.runners_jobs_add()}
					</Button>
				</div>

				{error ? (
					<p className="text-sm text-destructive" role="alert">
						{error}
					</p>
				) : null}

				<div className="grid gap-6 lg:grid-cols-2">
					<Card>
						<CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
							<CardTitle>{m.runners_jobs_title()}</CardTitle>
							<Button size="sm" variant="outline" onClick={handleOpenCreateJob}>
								<Plus data-icon="inline-start" />
								{m.runners_jobs_add()}
							</Button>
						</CardHeader>
						<CardContent className="flex flex-col gap-2">
							{jobs.length === 0 ? (
								<div className="flex flex-col items-center justify-center gap-3 py-8 text-center">
									<p className="text-sm text-muted-foreground">
										{m.runners_jobs_empty()}
									</p>
									<Button
										size="sm"
										variant="outline"
										onClick={handleOpenCreateJob}
									>
										<Plus data-icon="inline-start" />
										{m.runners_jobs_add()}
									</Button>
								</div>
							) : (
								jobs.map((job) => (
									<div
										key={job.id}
										className={`flex flex-col gap-2 rounded-lg border p-3 ${
											selectedJobId === job.id
												? "border-primary bg-primary/5"
												: "border-border"
										}`}
									>
										<button
											type="button"
											className="text-left focus-visible:outline-none"
											onClick={() => setSelectedJobId(job.id)}
										>
											<div className="flex items-center justify-between gap-2">
												<span className="font-medium">{job.name}</span>
												<Badge variant="outline">{job.status}</Badge>
											</div>
											<p className="mt-1 font-mono text-xs text-muted-foreground">
												{job.slug} · {job.cron}
											</p>
										</button>

										<div className="flex items-center gap-1.5 font-mono text-xs text-muted-foreground">
											<span className="text-foreground/80 font-medium">
												ID:
											</span>
											<span className="select-all text-foreground/80">
												{job.id}
											</span>
											<button
												type="button"
												className="inline-flex size-4 items-center justify-center text-muted-foreground hover:text-foreground focus-visible:outline-none"
												onClick={(e) => handleCopyJobId(job.id, e)}
												title="Copy Job ID"
												aria-label="Copy Job ID"
											>
												{copiedJobId === job.id ? (
													<Check className="size-3 text-emerald-500" />
												) : (
													<Copy className="size-3" />
												)}
											</button>
										</div>

										<div className="flex flex-wrap gap-2">
											<Button
												size="sm"
												variant="outline"
												onClick={() => void handleTrigger(job)}
											>
												{m.beepers_trigger_run()}
											</Button>
											<Button
												size="sm"
												variant="outline"
												onClick={() => handleOpenEditJob(job)}
											>
												<Edit data-icon="inline-start" />
												{m.common_edit()}
											</Button>
											<Button
												size="sm"
												variant="outline"
												onClick={() => void handleToggle(job)}
											>
												{job.status === "paused" ? (
													<Play data-icon="inline-start" />
												) : (
													<Pause data-icon="inline-start" />
												)}
												{job.status === "paused"
													? m.beepers_resume()
													: m.beepers_pause()}
											</Button>
											<Button
												size="sm"
												variant="ghost"
												onClick={() => void handleDelete(job)}
											>
												<Trash2 data-icon="inline-start" />
											</Button>
										</div>
									</div>
								))
							)}
						</CardContent>
					</Card>

					<Card>
						<CardHeader>
							<CardTitle>{m.runners_runs_title()}</CardTitle>
							<CardDescription className="font-mono text-xs">
								{selectedJob ? (
									<span>
										{selectedJob.name} ({selectedJob.id}) · {selectedJob.slug}
									</span>
								) : (
									m.runners_runs_select_job()
								)}
							</CardDescription>
						</CardHeader>
						<CardContent className="flex flex-col gap-3">
							<div className="flex flex-col gap-1">
								{runs.map((run) => (
									<button
										key={run.id}
										type="button"
										className="flex items-center justify-between rounded-md border px-3 py-2 text-left text-xs"
										onClick={() => {
											if (!selectedJob) return;
											void fetchRunnerJobRun(
												slug,
												runner.id,
												selectedJob.id,
												run.id,
											).then(setSelectedRun);
										}}
									>
										<span className="font-mono">
											{new Date(run.scheduled_for).toLocaleString()}
										</span>
										<Badge variant="outline">
											{run.result_status || run.status}
										</Badge>
									</button>
								))}
							</div>
							<pre className="max-h-80 overflow-auto rounded-lg bg-muted/60 p-3 font-mono text-[11px] whitespace-pre-wrap">
								{selectedRun?.log ||
									selectedRun?.log_preview ||
									m.runners_runs_no_log()}
							</pre>
						</CardContent>
					</Card>
				</div>
			</div>

			<RunnerJobFormDialog
				slug={slug}
				runnerId={runner.id}
				job={editingJob}
				open={isJobDialogOpen}
				onOpenChange={setIsJobDialogOpen}
				onSuccess={async (job) => {
					await router.invalidate();
					setSelectedJobId(job.id);
				}}
			/>
		</>
	);
}
