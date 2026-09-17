import { apiFetch } from "@/lib/api/client";

export type BeepRun = {
	id: string;
	scheduled_for: string;
	status:
		| "pending"
		| "running"
		| "succeeded"
		| "failed"
		| "skipped"
		| "expired";
	result: Record<string, unknown> | null;
	created_at: string;
};

export type Beep = {
	id: string;
	title: string;
	body: string | null;
	kind: "once" | "recurring";
	status: "active" | "paused" | "completed" | "cancelled" | "firing";
	cron: string | null;
	run_at: string | null;
	next_run_at: string | null;
	last_run_at: string | null;
	timezone: string;
	notification_channels: string[];
	beeper_id?: string | null;
	beeper?: {
		slug: string;
		name: string;
	} | null;
	created_at: string;
	run_stats?: {
		total: number;
		succeeded: number;
	};
	runs: BeepRun[];
};

export type PaginationMeta = {
	page: number;
	next_page: string | null;
	has_more: boolean;
	total_count?: number;
};

export type BeepsResponse = {
	beeps: Beep[];
	pagination?: PaginationMeta;
};

export type BeepStatsData = {
	active: number;
	due_today: number;
	firing: number;
	recurring?: number;
	completed?: number;
	all?: number;
};

export type BeepStatsResponse = {
	stats: BeepStatsData;
};

export function fetchBeeps(
	slug: string,
	options?: {
		page?: string | null;
		status?: string | null;
		kind?: string | null;
	},
) {
	const params = new URLSearchParams();
	if (options?.page) params.set("page", options.page);
	if (options?.status) params.set("status", options.status);
	if (options?.kind) params.set("kind", options.kind);
	const query = params.toString() ? `?${params.toString()}` : "";
	return apiFetch<BeepsResponse>(`/api/v1/${slug}/beeps${query}`, {
		method: "GET",
	});
}

export function fetchBeepStats(slug: string) {
	return apiFetch<BeepStatsResponse>(`/api/v1/${slug}/beeps/stats`, {
		method: "GET",
	});
}

export function fetchBeep(slug: string, beepId: string) {
	return apiFetch<Beep>(`/api/v1/${slug}/beeps/${beepId}`, {
		method: "GET",
	});
}

export type BeepRunsResponse = {
	runs: BeepRun[];
	pagination?: PaginationMeta;
};

export function fetchBeepRuns(
	slug: string,
	beepId: string,
	options?: { page?: string | null },
) {
	const query = options?.page
		? `?page=${encodeURIComponent(options.page)}`
		: "";
	return apiFetch<BeepRunsResponse>(
		`/api/v1/${slug}/beeps/${beepId}/runs${query}`,
		{
			method: "GET",
		},
	);
}

export function createBeep(
	slug: string,
	body: {
		title: string;
		body?: string | null;
		kind?: "once" | "recurring";
		run_at?: string | null;
		cron?: string | null;
		timezone?: string;
		notification_channels?: string[];
	},
) {
	return apiFetch<Beep>(`/api/v1/${slug}/beeps`, {
		method: "POST",
		body,
	});
}

export function triggerBeepRun(slug: string, beepId: string) {
	return apiFetch<BeepRun>(`/api/v1/${slug}/beeps/${beepId}/runs`, {
		method: "POST",
	});
}

export function pauseBeep(slug: string, beepId: string) {
	return apiFetch<Beep>(`/api/v1/${slug}/beeps/${beepId}/pause`, {
		method: "POST",
	});
}

export function resumeBeep(slug: string, beepId: string) {
	return apiFetch<Beep>(`/api/v1/${slug}/beeps/${beepId}/pause`, {
		method: "DELETE",
	});
}

export function deleteBeep(slug: string, beepId: string) {
	return apiFetch<void>(`/api/v1/${slug}/beeps/${beepId}`, {
		method: "DELETE",
	});
}

export type BeepProposal = {
	intent: "create" | "other";
	kind?: "once" | "recurring";
	title: string | null;
	body: string | null;
	run_at: string | null;
	cron: string | null;
	timezone: string;
	errors: {
		title?: string;
		body?: string;
		run_at?: string;
		cron?: string;
	};
	confirmable: boolean;
	message: string | null;
};

export function createBeepProposal(
	slug: string,
	prompt: string,
	timezone?: string,
) {
	return apiFetch<BeepProposal>(`/api/v1/${slug}/beep_proposals`, {
		method: "POST",
		body: { prompt, timezone },
	});
}
