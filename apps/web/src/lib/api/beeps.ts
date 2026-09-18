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
	intent?: string | null;
	metadata?: Record<string, unknown> | null;
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

export const BEEP_SORT_FIELDS = [
	"title",
	"status",
	"next_run_at",
	"created_at",
] as const;

export type BeepSortField = (typeof BEEP_SORT_FIELDS)[number];
export type BeepSortDir = "asc" | "desc";

export function isBeepSortField(val: string): val is BeepSortField {
	return (BEEP_SORT_FIELDS as readonly string[]).includes(val);
}

export const DEFAULT_BEEP_SORT: { sort: BeepSortField; dir: BeepSortDir } = {
	sort: "created_at",
	dir: "desc",
};

export function parseBeepSort(
	sort?: string,
	dir?: string,
): { sort: BeepSortField; dir: BeepSortDir } {
	const field =
		sort === "schedule" || sort === "scheduled_at"
			? "next_run_at"
			: sort && isBeepSortField(sort)
				? sort
				: DEFAULT_BEEP_SORT.sort;
	const direction: BeepSortDir = dir === "asc" ? "asc" : "desc";
	return { sort: field, dir: direction };
}

export function beepSortQuery(sort: BeepSortField, dir: BeepSortDir) {
	if (sort === DEFAULT_BEEP_SORT.sort && dir === DEFAULT_BEEP_SORT.dir) {
		return {};
	}
	return { sort, dir };
}

export function fetchBeeps(
	slug: string,
	options?: {
		page?: string | null;
		status?: string | null;
		kind?: string | null;
		q?: string | null;
		sort?: string | null;
		dir?: string | null;
		signal?: AbortSignal;
	},
) {
	const params = new URLSearchParams();
	if (options?.page) params.set("page", options.page);
	if (options?.status) params.set("status", options.status);
	if (options?.kind) params.set("kind", options.kind);
	if (options?.q) params.set("q", options.q);
	if (options?.sort) params.set("sort", options.sort);
	if (options?.dir) params.set("dir", options.dir);
	const query = params.toString() ? `?${params.toString()}` : "";
	return apiFetch<BeepsResponse>(`/api/v1/${slug}/beeps${query}`, {
		method: "GET",
		signal: options?.signal,
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
		intent?: string | null;
		metadata?: Record<string, unknown> | null;
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
	action?: "create" | "other";
	intent: string | null;
	metadata?: Record<string, unknown> | null;
	kind?: "once" | "recurring";
	title: string | null;
	body: string | null;
	run_at: string | null;
	cron: string | null;
	timezone: string;
	notification_channels?: string[] | null;
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
