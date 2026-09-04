import {
	createFileRoute,
	getRouteApi,
	notFound,
	useRouter,
} from "@tanstack/react-router";
import { Pause, Play, Plus, Trash2 } from "lucide-react";
import { useEffect, useState } from "react";
import { DashboardHeader } from "@/components/dashboard/dashboard-header";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { ApiError } from "@/lib/api/client";
import {
	createRunnerJob,
	deleteRunnerJob,
	fetchRunnerJobRun,
	fetchRunnerJobRuns,
	fetchRunnerJobs,
	pauseRunnerJob,
	resumeRunnerJob,
	type RunnerJob,
	type RunnerRun,
	triggerRunnerJobRun,
} from "@/lib/api/runner-jobs";
import { fetchRunner } from "@/lib/api/runners";
import { withAuthRedirects } from "@/lib/auth/guards";
import { translateError } from "@/lib/i18n-labels";
import { browserTimezone } from "@/lib/timezone";
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
	const [name, setName] = useState("");
	const [jobSlug, setJobSlug] = useState("");
	const [cron, setCron] = useState("*/5 * * * *");
	const [submitting, setSubmitting] = useState(false);

	const selectedJob = jobs.find((job) => job.id === selectedJobId) ?? null;

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

	async function handleCreateJob(event: React.FormEvent) {
		event.preventDefault();
		setSubmitting(true);
		setError(null);
		try {
			const job = await createRunnerJob(slug, runner.id, {
				name: name.trim(),
				slug: jobSlug.trim(),
				cron: cron.trim(),
				timezone: browserTimezone(),
			});
			setName("");
			setJobSlug("");
			await router.invalidate();
			setSelectedJobId(job.id);
		} catch (err) {
			setError(
				err instanceof ApiError
					? err.message
					: translateError(err) || m.runners_jobs_create_failed(),
			);
		} finally {
			setSubmitting(false);
		}
	}

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
				<div>
					<h1 className="font-heading text-2xl font-bold tracking-tight">
						{runner.name}
					</h1>
					<p className="text-sm text-muted-foreground">
						{m.runners_workspace_hint()}
					</p>
				</div>

				{error ? (
					<p className="text-sm text-destructive" role="alert">
						{error}
					</p>
				) : null}

				<Card>
					<CardHeader>
						<CardTitle>{m.runners_jobs_add()}</CardTitle>
						<CardDescription>{m.runners_jobs_add_hint()}</CardDescription>
					</CardHeader>
					<CardContent>
						<form
							onSubmit={(event) => void handleCreateJob(event)}
							className="grid gap-3 sm:grid-cols-4"
						>
							<div className="flex flex-col gap-1.5 sm:col-span-1">
								<Label htmlFor="job-name">{m.runners_jobs_name()}</Label>
								<Input
									id="job-name"
									required
									value={name}
									onChange={(e) => setName(e.target.value)}
									placeholder="Intranet HTTP"
								/>
							</div>
							<div className="flex flex-col gap-1.5 sm:col-span-1">
								<Label htmlFor="job-slug">{m.runners_jobs_slug()}</Label>
								<Input
									id="job-slug"
									required
									value={jobSlug}
									onChange={(e) => setJobSlug(e.target.value)}
									placeholder="intranet-http"
									className="font-mono"
								/>
							</div>
							<div className="flex flex-col gap-1.5 sm:col-span-1">
								<Label htmlFor="job-cron">{m.beeps_cron()}</Label>
								<Input
									id="job-cron"
									required
									value={cron}
									onChange={(e) => setCron(e.target.value)}
									className="font-mono"
								/>
							</div>
							<div className="flex items-end">
								<Button type="submit" size="sm" disabled={submitting}>
									<Plus data-icon="inline-start" />
									{m.runners_jobs_add()}
								</Button>
							</div>
						</form>
					</CardContent>
				</Card>

				<div className="grid gap-6 lg:grid-cols-2">
					<Card>
						<CardHeader>
							<CardTitle>{m.runners_jobs_title()}</CardTitle>
						</CardHeader>
						<CardContent className="flex flex-col gap-2">
							{jobs.length === 0 ? (
								<p className="text-sm text-muted-foreground">
									{m.runners_jobs_empty()}
								</p>
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
											className="text-left"
											onClick={() => setSelectedJobId(job.id)}
										>
											<div className="flex items-center justify-between gap-2">
												<span className="font-medium">{job.name}</span>
												<Badge variant="outline">{job.status}</Badge>
											</div>
											<p className="font-mono text-xs text-muted-foreground">
												{job.slug} · {job.cron}
											</p>
										</button>
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
							<CardDescription>
								{selectedJob
									? selectedJob.slug
									: m.runners_runs_select_job()}
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
		</>
	);
}
