import { apiFetch } from "@/lib/api/client";

export type PluginInput = {
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

export type PluginMetric = {
	name: string;
	label: string;
	type: "number" | "string" | "boolean";
	unit?: string;
};

export type Plugin = {
	id: string;
	slug: string;
	name: string;
	version: string;
	description: string;
	default_cron?: string;
	failure_threshold?: number;
	min_interval_seconds?: number;
	webhook_ingest?: boolean;
	inputs: PluginInput[];
	metrics: PluginMetric[];
	created_at: string;
};

export type PluginsResponse = {
	plugins: Plugin[];
};

export function fetchPlugins() {
	return apiFetch<PluginsResponse>("/api/v1/plugins", {
		method: "GET",
	});
}

export function fetchPlugin(slug: string) {
	return apiFetch<Plugin>(`/api/v1/plugins/${slug}`, {
		method: "GET",
	});
}
