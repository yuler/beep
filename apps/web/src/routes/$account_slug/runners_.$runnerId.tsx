import {
	createFileRoute,
	getRouteApi,
	Link,
	notFound,
	useNavigate,
	useRouter,
} from "@tanstack/react-router";
import {
	Check,
	ChevronRight,
	Clock,
	Copy,
	Edit,
	Pause,
	Play,
	Plus,
	Server,
	Trash2,
} from "lucide-react";
import { useState } from "react";
import { confirm } from "@/components/confirm-dialog";
import { DashboardHeader } from "@/components/dashboard/dashboard-header";
import { RunnerJobFormDialog } from "@/components/runners/runner-job-form-dialog";
import {
	formatRunnerLastSeen,
	getRunnerStatusBadge,
} from "@/components/runners/runner-list";
import { RunnerJobStatusBadge } from "@/components/runners/runner-run-status";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { ApiError } from "@/lib/api/client";
import {
	deleteRunnerJob,
	fetchRunnerJobs,
	pauseRunnerJob,
	type RunnerJob,
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
	const navigate = useNavigate();
	const { runner, jobs } = Route.useLoaderData();
	const [error, setError] = useState<string | null>(null);
	const [isJobDialogOpen, setIsJobDialogOpen] = useState(false);
	const [editingJob, setEditingJob] = useState<RunnerJob | null>(null);
	const [copiedJobId, setCopiedJobId] = useState<string | null>(null);

	function handleCopyJobId(id: string) {
		void navigator.clipboard.writeText(id);
		setCopiedJobId(id);
		window.setTimeout(() => setCopiedJobId(null), 2000);
	}

	function handleOpenCreateJob() {
		setEditingJob(null);
		setIsJobDialogOpen(true);
	}

	function handleOpenEditJob(job: RunnerJob) {
		setEditingJob(job);
		setIsJobDialogOpen(true);
	}

	async function handleTrigger(job: RunnerJob) {
		setError(null);
		try {
			const run = await triggerRunnerJobRun(slug, runner.id, job.id);
			await router.invalidate();
			await navigate({
				to: "/$account_slug/runners/$runnerId/jobs/$jobId/runs/$runId",
				params: {
					account_slug: slug,
					runnerId: runner.id,
					jobId: job.id,
					runId: run.id,
				},
			});
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
		if (!(await confirm(m.runners_jobs_delete_confirm({ name: job.name })))) {
			return;
		}
		setError(null);
		try {
			await deleteRunnerJob(slug, runner.id, job.id);
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

			<div className="flex flex-1 flex-col gap-6">
				<div className="flex flex-wrap items-center justify-between gap-4">
					<div className="flex items-center gap-3">
						<div className="flex size-10 shrink-0 items-center justify-center rounded-md border border-line text-ink">
							<Server className="size-5" />
						</div>
						<div className="flex flex-col gap-1">
							<div className="flex items-center gap-2.5">
								<h1 className="font-heading text-2xl font-bold tracking-tight">
									{runner.name}
								</h1>
								{getRunnerStatusBadge(runner)}
							</div>
							<div className="flex flex-wrap items-center gap-x-3 gap-y-1 font-mono text-xs text-gray">
								<span>{runner.masked_token}</span>
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

				{!runner.is_online && runner.status === "offline" ? (
					<Alert variant="warning">
						<AlertDescription className="text-xs">
							{m.runners_workspace_hint()}
						</AlertDescription>
					</Alert>
				) : null}

				{error ? (
					<p className="text-sm text-danger" role="alert">
						{error}
					</p>
				) : null}

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
								<p className="text-sm text-gray">{m.runners_jobs_empty()}</p>
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
									className="flex flex-col gap-2 rounded-md border border-line p-3 hover:bg-wash"
								>
									<Link
										to="/$account_slug/runners/$runnerId/jobs/$jobId"
										params={{
											account_slug: slug,
											runnerId: runner.id,
											jobId: job.id,
										}}
										className="min-w-0 text-left"
									>
										<div className="flex items-center justify-between gap-2">
											<span className="font-medium text-ink">{job.name}</span>
											<span className="flex items-center gap-1.5">
												<RunnerJobStatusBadge status={job.status} />
												<ChevronRight className="size-4 text-gray" />
											</span>
										</div>
										<p className="mt-1 font-mono text-xs text-gray">
											{job.slug} · {job.cron}
										</p>
										<div className="mt-1.5 flex items-center gap-1.5 font-mono text-xs text-gray">
											<span className="font-medium text-ink">ID:</span>
											<span className="select-all text-ink">{job.id}</span>
										</div>
									</Link>

									<div className="flex flex-wrap gap-2">
										<Button
											size="sm"
											variant="outline"
											onClick={() => handleCopyJobId(job.id)}
											title={m.common_copy()}
											aria-label={m.common_copy()}
										>
											{copiedJobId === job.id ? (
												<Check className="size-3 text-success" />
											) : (
												<Copy className="size-3" />
											)}
										</Button>
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
											variant="destructive"
											onClick={() => void handleDelete(job)}
											aria-label={m.common_delete()}
										>
											<Trash2 data-icon="inline-start" />
										</Button>
									</div>
								</div>
							))
						)}
					</CardContent>
				</Card>
			</div>

			<RunnerJobFormDialog
				slug={slug}
				runnerId={runner.id}
				job={editingJob}
				open={isJobDialogOpen}
				onOpenChange={setIsJobDialogOpen}
				onSuccess={async (job) => {
					await router.invalidate();
					await navigate({
						to: "/$account_slug/runners/$runnerId/jobs/$jobId",
						params: {
							account_slug: slug,
							runnerId: runner.id,
							jobId: job.id,
						},
					});
				}}
			/>
		</>
	);
}
