import { apiFetch } from "@/lib/api/client";

export type BeeperAppInput = {
	name: string;
	label: string;
	type: "string" | "number" | "boolean" | "url" | "enum" | "secret";
	required?: boolean;
	default?: unknown;
	min?: number;
	max?: number;
	placeholder?: string;
	options?: (string | number)[];
};

export type BeeperAppMetric = {
	name: string;
	label: string;
	type: "number" | "string" | "boolean";
	unit?: string;
};

export type BeeperApp = {
	id: string;
	slug: string;
	name: string;
	version: string;
	description: string;
	default_cron?: string;
	failure_threshold?: number;
	min_interval_seconds?: number;
	capabilities?: string[];
	webhook_ping?: boolean;
	official?: boolean;
	inputs: BeeperAppInput[];
	metrics: BeeperAppMetric[];
	created_at: string;
};

export type BeeperAppsResponse = {
	beeper_apps: BeeperApp[];
};

export type BeeperRun = {
	id: string;
	scheduled_for: string;
	status:
		| "pending"
		| "running"
		| "succeeded"
		| "failed"
		| "skipped"
		| "expired";
	signal_status?: "ok" | "alerting" | "error" | null;
	signal_result?: {
		status?: string;
		title?: string;
		message?: string;
		metrics?: Record<string, unknown>;
	} | null;
	created_at: string;
};

export type Beeper = {
	id: string;
	title: string;
	body?: string | null;
	cron: string;
	timezone: string;
	status: "active" | "paused" | "completed" | "cancelled" | "firing";
	alert_state: "ok" | "alerting";
	consecutive_failures: number;
	config: Record<string, unknown>;
	signal_metadata?: Record<string, unknown>;
	notification_channels: string[];
	ping_token?: string | null;
	last_ping_at?: string | null;
	next_run_at?: string | null;
	last_run_at?: string | null;
	created_at: string;
	updated_at: string;
	beeper_app?: {
		id: string;
		slug: string;
		name: string;
		version: string;
		description?: string;
		inputs?: BeeperAppInput[];
		metrics?: BeeperAppMetric[];
	};
	runs?: BeeperRun[];
};

export type BeepersResponse = {
	beepers: Beeper[];
};

export function fetchBeeperApps() {
	return apiFetch<BeeperAppsResponse>("/api/v1/beeper_apps", {
		method: "GET",
	});
}

export function fetchBeeperApp(slug: string) {
	return apiFetch<BeeperApp>(`/api/v1/beeper_apps/${slug}`, {
		method: "GET",
	});
}

export function fetchBeepers(accountSlug: string) {
	return apiFetch<BeepersResponse>(`/api/v1/${accountSlug}/beepers`, {
		method: "GET",
	});
}

export function fetchBeeper(accountSlug: string, beeperId: string) {
	return apiFetch<Beeper>(`/api/v1/${accountSlug}/beepers/${beeperId}`, {
		method: "GET",
	});
}

export function createBeeper(
	accountSlug: string,
	body: {
		beeper_app_id?: string;
		beeper_app_slug?: string;
		title: string;
		body?: string;
		cron?: string;
		timezone?: string;
		config?: Record<string, unknown>;
		notification_channels?: string[];
	},
) {
	return apiFetch<Beeper>(`/api/v1/${accountSlug}/beepers`, {
		method: "POST",
		body,
	});
}

export function updateBeeper(
	accountSlug: string,
	beeperId: string,
	body: {
		title?: string;
		body?: string | null;
		cron?: string;
		config?: Record<string, unknown>;
		notification_channels?: string[];
	},
) {
	return apiFetch<Beeper>(`/api/v1/${accountSlug}/beepers/${beeperId}`, {
		method: "PATCH",
		body,
	});
}

export function pauseBeeper(accountSlug: string, beeperId: string) {
	return apiFetch<Beeper>(`/api/v1/${accountSlug}/beepers/${beeperId}/pause`, {
		method: "POST",
	});
}

export function resumeBeeper(accountSlug: string, beeperId: string) {
	return apiFetch<Beeper>(`/api/v1/${accountSlug}/beepers/${beeperId}/pause`, {
		method: "DELETE",
	});
}

export function deleteBeeper(accountSlug: string, beeperId: string) {
	return apiFetch<void>(`/api/v1/${accountSlug}/beepers/${beeperId}`, {
		method: "DELETE",
	});
}

export function triggerBeeperRun(accountSlug: string, beeperId: string) {
	return apiFetch<BeeperRun>(
		`/api/v1/${accountSlug}/beepers/${beeperId}/runs`,
		{
			method: "POST",
		},
	);
}
