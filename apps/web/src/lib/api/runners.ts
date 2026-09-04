import { apiFetch } from "@/lib/api/client";

export type RunnerStatus = "online" | "idle" | "offline";

export type Runner = {
	id: string;
	name: string;
	token_prefix: string;
	status: RunnerStatus;
	tags: string[];
	version?: string | null;
	os?: string | null;
	arch?: string | null;
	hostname?: string | null;
	ip_address?: string | null;
	last_seen_at?: string | null;
	is_online?: boolean;
	jobs_count?: number;
	created_at: string;
	updated_at: string;
};

export type RunnerWithToken = Runner & {
	token?: string;
};

export type RunnersResponse = {
	runners: Runner[];
};

export type RunnerResponse = {
	runner: Runner;
};

export type RunnerWithTokenResponse = {
	runner: RunnerWithToken;
};

export function fetchRunners(accountSlug: string) {
	return apiFetch<RunnersResponse>(`/api/v1/${accountSlug}/runners`, {
		method: "GET",
	});
}

export function fetchRunner(accountSlug: string, runnerId: string) {
	return apiFetch<RunnerResponse>(
		`/api/v1/${accountSlug}/runners/${runnerId}`,
		{
			method: "GET",
		},
	);
}

export function createRunner(
	accountSlug: string,
	body: {
		name: string;
		tags?: string[];
	},
) {
	return apiFetch<RunnerWithTokenResponse>(`/api/v1/${accountSlug}/runners`, {
		method: "POST",
		body: { runner: body },
	});
}

export function updateRunner(
	accountSlug: string,
	runnerId: string,
	body: {
		name?: string;
		tags?: string[];
	},
) {
	return apiFetch<RunnerResponse>(
		`/api/v1/${accountSlug}/runners/${runnerId}`,
		{
			method: "PUT",
			body: { runner: body },
		},
	);
}

export function deleteRunner(accountSlug: string, runnerId: string) {
	return apiFetch<void>(`/api/v1/${accountSlug}/runners/${runnerId}`, {
		method: "DELETE",
	});
}

export function regenerateRunnerToken(accountSlug: string, runnerId: string) {
	return apiFetch<RunnerWithTokenResponse>(
		`/api/v1/${accountSlug}/runners/${runnerId}/regenerate_token`,
		{
			method: "POST",
		},
	);
}
