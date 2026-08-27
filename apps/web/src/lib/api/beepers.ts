import { apiFetch } from "@/lib/api/client";

export type BeeperInput = {
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

export type BeeperMetric = {
	name: string;
	label: string;
	type: "number" | "string" | "boolean";
	unit?: string;
};

export type Beeper = {
	id: string;
	slug: string;
	name: string;
	version: string;
	description: string;
	default_cron?: string;
	failure_threshold?: number;
	min_interval_seconds?: number;
	webhook_ingest?: boolean;
	inputs: BeeperInput[];
	metrics: BeeperMetric[];
	created_at: string;
};

export type BeepersResponse = {
	beepers: Beeper[];
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

export type BeeperInstall = {
	id: string;
	title: string;
	cron: string;
	timezone: string;
	status: "active" | "paused" | "completed" | "cancelled" | "firing";
	alert_state: "ok" | "alerting";
	consecutive_failures: number;
	config: Record<string, unknown>;
	notification_channels: string[];
	ping_token?: string | null;
	last_ping_at?: string | null;
	next_run_at?: string | null;
	last_run_at?: string | null;
	created_at: string;
	updated_at: string;
	beeper?: {
		id: string;
		slug: string;
		name: string;
		version: string;
		description?: string;
		inputs?: BeeperInput[];
		metrics?: BeeperMetric[];
	};
	runs?: BeeperRun[];
};

export type BeeperInstallsResponse = {
	beeper_installs: BeeperInstall[];
};

export function fetchBeepers() {
	return apiFetch<BeepersResponse>("/api/v1/beepers", {
		method: "GET",
	});
}

export function fetchBeeper(slug: string) {
	return apiFetch<Beeper>(`/api/v1/beepers/${slug}`, {
		method: "GET",
	});
}

export function fetchBeeperInstalls(accountSlug: string) {
	return apiFetch<BeeperInstallsResponse>(
		`/api/v1/${accountSlug}/beeper_installs`,
		{
			method: "GET",
		},
	);
}

export function fetchBeeperInstall(accountSlug: string, installId: string) {
	return apiFetch<BeeperInstall>(
		`/api/v1/${accountSlug}/beeper_installs/${installId}`,
		{
			method: "GET",
		},
	);
}

export function createBeeperInstall(
	accountSlug: string,
	body: {
		beeper_id?: string;
		beeper_slug?: string;
		title: string;
		cron?: string;
		timezone?: string;
		config?: Record<string, unknown>;
		notification_channels?: string[];
	},
) {
	return apiFetch<BeeperInstall>(`/api/v1/${accountSlug}/beeper_installs`, {
		method: "POST",
		body,
	});
}

export function pauseBeeperInstall(accountSlug: string, installId: string) {
	return apiFetch<BeeperInstall>(
		`/api/v1/${accountSlug}/beeper_installs/${installId}/pause`,
		{
			method: "POST",
		},
	);
}

export function resumeBeeperInstall(accountSlug: string, installId: string) {
	return apiFetch<BeeperInstall>(
		`/api/v1/${accountSlug}/beeper_installs/${installId}/pause`,
		{
			method: "DELETE",
		},
	);
}

export function deleteBeeperInstall(accountSlug: string, installId: string) {
	return apiFetch<void>(`/api/v1/${accountSlug}/beeper_installs/${installId}`, {
		method: "DELETE",
	});
}

export function triggerBeeperInstallRun(
	accountSlug: string,
	installId: string,
) {
	return apiFetch<BeeperRun>(
		`/api/v1/${accountSlug}/beeper_installs/${installId}/runs`,
		{
			method: "POST",
		},
	);
}
