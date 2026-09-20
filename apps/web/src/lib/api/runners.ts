import { apiFetch } from "@/lib/api/client";

export type RunnerStatus = "online" | "offline";

export type Runner = {
	id: string;
	name: string;
	masked_token: string;
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
		`/api/v1/${accountSlug}/runners/${runnerId}/token`,
		{
			method: "POST",
		},
	);
}

export type RunnerDeviceAuthInfo = {
	user_code: string;
	runner_name?: string;
	tags?: string[];
	metadata?: {
		os?: string;
		arch?: string;
		hostname?: string;
		version?: string;
	};
	status: string;
	expires_in: number;
};

export async function verifyRunnerDeviceCode(
	accountSlug: string,
	userCode: string,
): Promise<RunnerDeviceAuthInfo> {
	return apiFetch<RunnerDeviceAuthInfo>(
		`/api/v1/${encodeURIComponent(accountSlug)}/runners/authorizations/${encodeURIComponent(userCode)}`,
		{
			method: "GET",
		},
	);
}

export async function approveRunnerDeviceAuth(
	accountSlug: string,
	data: {
		user_code: string;
		runner_name?: string;
		tags?: string[];
	},
): Promise<{
	status: string;
	runner: {
		id: string;
		name: string;
		tags: string[];
		status: string;
		masked_token?: string;
	};
}> {
	return apiFetch<{
		status: string;
		runner: {
			id: string;
			name: string;
			tags: string[];
			status: string;
			masked_token?: string;
		};
	}>(
		`/api/v1/${encodeURIComponent(accountSlug)}/runners/authorizations/${encodeURIComponent(data.user_code)}`,
		{
			method: "PATCH",
			body: {
				runner_name: data.runner_name,
				tags: data.tags,
			},
		},
	);
}

export async function denyRunnerDeviceAuth(
	accountSlug: string,
	user_code: string,
): Promise<void> {
	await apiFetch(
		`/api/v1/${encodeURIComponent(accountSlug)}/runners/authorizations/${encodeURIComponent(user_code)}`,
		{
			method: "DELETE",
		},
	);
}
