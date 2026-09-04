import { apiFetch } from "@/lib/api/client";

export type RunnerJobStatus = "active" | "paused" | "firing";
export type RunnerRunStatus =
	| "pending"
	| "running"
	| "succeeded"
	| "failed"
	| "expired";
export type RunnerResultStatus = "ok" | "alerting" | "error";

export type RunnerJob = {
	id: string;
	runner_id: string;
	name: string;
	slug: string;
	cron: string;
	timezone: string;
	status: RunnerJobStatus;
	timeout_seconds: number;
	config: Record<string, unknown>;
	next_run_at?: string | null;
	last_run_at?: string | null;
	runner_online?: boolean;
	created_at: string;
	updated_at: string;
};

export type RunnerRun = {
	id: string;
	runner_job_id: string;
	runner_id: string;
	scheduled_for: string;
	status: RunnerRunStatus;
	claimed_at?: string | null;
	result_status?: RunnerResultStatus | null;
	result?: {
		status?: string;
		title?: string;
		message?: string;
		metrics?: Record<string, unknown>;
	} | null;
	log?: string;
	log_preview?: string;
	created_at: string;
	updated_at: string;
};

export function fetchRunnerJobs(accountSlug: string, runnerId: string) {
	return apiFetch<{ jobs: RunnerJob[] }>(
		`/api/v1/${accountSlug}/runners/${runnerId}/jobs`,
		{ method: "GET" },
	);
}

export function createRunnerJob(
	accountSlug: string,
	runnerId: string,
	body: {
		name: string;
		slug: string;
		cron: string;
		timeout_seconds?: number;
		timezone?: string;
		config?: Record<string, unknown>;
	},
) {
	return apiFetch<RunnerJob>(
		`/api/v1/${accountSlug}/runners/${runnerId}/jobs`,
		{ method: "POST", body },
	);
}

export function updateRunnerJob(
	accountSlug: string,
	runnerId: string,
	jobId: string,
	body: {
		name?: string;
		cron?: string;
		timeout_seconds?: number;
		config?: Record<string, unknown>;
	},
) {
	return apiFetch<RunnerJob>(
		`/api/v1/${accountSlug}/runners/${runnerId}/jobs/${jobId}`,
		{ method: "PATCH", body },
	);
}

export function deleteRunnerJob(
	accountSlug: string,
	runnerId: string,
	jobId: string,
) {
	return apiFetch<void>(
		`/api/v1/${accountSlug}/runners/${runnerId}/jobs/${jobId}`,
		{ method: "DELETE" },
	);
}

export function pauseRunnerJob(
	accountSlug: string,
	runnerId: string,
	jobId: string,
) {
	return apiFetch<RunnerJob>(
		`/api/v1/${accountSlug}/runners/${runnerId}/jobs/${jobId}/pause`,
		{ method: "POST" },
	);
}

export function resumeRunnerJob(
	accountSlug: string,
	runnerId: string,
	jobId: string,
) {
	return apiFetch<RunnerJob>(
		`/api/v1/${accountSlug}/runners/${runnerId}/jobs/${jobId}/pause`,
		{ method: "DELETE" },
	);
}

export function fetchRunnerJobRuns(
	accountSlug: string,
	runnerId: string,
	jobId: string,
) {
	return apiFetch<{ runs: RunnerRun[] }>(
		`/api/v1/${accountSlug}/runners/${runnerId}/jobs/${jobId}/runs`,
		{ method: "GET" },
	);
}

export function fetchRunnerJobRun(
	accountSlug: string,
	runnerId: string,
	jobId: string,
	runId: string,
) {
	return apiFetch<RunnerRun>(
		`/api/v1/${accountSlug}/runners/${runnerId}/jobs/${jobId}/runs/${runId}`,
		{ method: "GET" },
	);
}

export function triggerRunnerJobRun(
	accountSlug: string,
	runnerId: string,
	jobId: string,
) {
	return apiFetch<RunnerRun>(
		`/api/v1/${accountSlug}/runners/${runnerId}/jobs/${jobId}/runs`,
		{ method: "POST" },
	);
}
