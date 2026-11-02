import { useNavigate, useRouter } from "@tanstack/react-router";
import { Edit, Pause, Play, Trash2 } from "lucide-react";
import { useState } from "react";
import { confirm } from "@/components/confirm-dialog";
import { DashboardHeader } from "@/components/dashboard/dashboard-header";
import { RunnerJobFormDialog } from "@/components/runners/runner-job-form-dialog";
import { RunnerJobWorkspace } from "@/components/runners/runner-job-workspace";
import { RunnerJobStatusBadge } from "@/components/runners/runner-run-status";
import { Alert, AlertDescription } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { ApiError } from "@/lib/api/client";
import {
	deleteRunnerJob,
	pauseRunnerJob,
	type RunnerJob,
	type RunnerRun,
	resumeRunnerJob,
	triggerRunnerJobRun,
} from "@/lib/api/runner-jobs";
import type { Runner } from "@/lib/api/runners";
import { formatBeepScheduleTime } from "@/lib/beep-datetime";
import { translateError } from "@/lib/i18n-labels";
import { m } from "@/locale/paraglide/messages";

export function RunnerJobScreen({
	slug,
	runner,
	job,
	runs,
	selectedRun,
}: {
	slug: string;
	runner: Runner;
	job: RunnerJob;
	runs: RunnerRun[];
	selectedRun: RunnerRun | null;
}) {
	const router = useRouter();
	const navigate = useNavigate();
	const [error, setError] = useState<string | null>(null);
	const [isJobDialogOpen, setIsJobDialogOpen] = useState(false);

	async function handleTrigger() {
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

	async function handleToggle() {
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

	async function handleDelete() {
		if (!(await confirm(m.runners_jobs_delete_confirm({ name: job.name })))) {
			return;
		}
		setError(null);
		try {
			await deleteRunnerJob(slug, runner.id, job.id);
			await navigate({
				to: "/$account_slug/runners/$runnerId",
				params: { account_slug: slug, runnerId: runner.id },
			});
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
					{
						label: runner.name,
						to: "/$account_slug/runners/$runnerId",
						params: { account_slug: slug, runnerId: runner.id },
					},
					...(selectedRun
						? [
								{
									label: job.name,
									to: "/$account_slug/runners/$runnerId/jobs/$jobId" as const,
									params: {
										account_slug: slug,
										runnerId: runner.id,
										jobId: job.id,
									},
								},
								{
									label: formatBeepScheduleTime(
										selectedRun.scheduled_for,
										job.timezone,
									),
									isCurrentPage: true,
								},
							]
						: [{ label: job.name, isCurrentPage: true }]),
				]}
			/>

			<div className="flex flex-1 flex-col gap-4">
				<div className="flex flex-wrap items-start justify-between gap-3">
					<div className="flex min-w-0 flex-col gap-1.5">
						<div className="flex flex-wrap items-center gap-2">
							<h1 className="font-heading text-2xl font-bold tracking-tight">
								{job.name}
							</h1>
							<RunnerJobStatusBadge status={job.status} />
						</div>
						<p className="font-mono text-xs text-gray">
							{job.slug} · {job.cron} · {job.timezone}
						</p>
					</div>
					<div className="flex flex-wrap items-center gap-2">
						<Button
							size="sm"
							variant="outline"
							onClick={() => void handleTrigger()}
						>
							<Play data-icon="inline-start" />
							{m.beepers_trigger_run()}
						</Button>
						<Button
							size="sm"
							variant="outline"
							onClick={() => setIsJobDialogOpen(true)}
						>
							<Edit data-icon="inline-start" />
							{m.common_edit()}
						</Button>
						<Button
							size="sm"
							variant="outline"
							onClick={() => void handleToggle()}
						>
							{job.status === "paused" ? (
								<Play data-icon="inline-start" />
							) : (
								<Pause data-icon="inline-start" />
							)}
							{job.status === "paused" ? m.beepers_resume() : m.beepers_pause()}
						</Button>
						<Button
							size="sm"
							variant="destructive"
							onClick={() => void handleDelete()}
						>
							<Trash2 data-icon="inline-start" />
							{m.common_delete()}
						</Button>
					</div>
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

				<RunnerJobWorkspace
					slug={slug}
					runnerId={runner.id}
					job={job}
					initialRuns={runs}
					selectedRun={selectedRun}
					onError={setError}
				/>
			</div>

			<RunnerJobFormDialog
				slug={slug}
				runnerId={runner.id}
				job={job}
				open={isJobDialogOpen}
				onOpenChange={setIsJobDialogOpen}
				onSuccess={async () => {
					await router.invalidate();
				}}
			/>
		</>
	);
}
