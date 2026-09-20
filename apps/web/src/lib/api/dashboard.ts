import { apiFetch } from "@/lib/api/client";

export type DashboardExecutionStats = {
	total: number;
	success_count: number;
	failure_count: number;
	success_rate: number;
};

export type DashboardBeepStats = {
	active: number;
	due_today: number;
	firing: number;
	recurring?: number;
	completed?: number;
	all?: number;
};

export type DashboardBeeperStats = {
	total: number;
	healthy: number;
	alerting: number;
};

export type DashboardRunnerStats = {
	total: number;
	online: number;
	active_jobs: number;
};

export type DashboardStats = {
	executions: DashboardExecutionStats;
	beeps: DashboardBeepStats;
	beepers: DashboardBeeperStats;
	runners: DashboardRunnerStats;
};

export type DashboardChartPoint = {
	timestamp: string;
	label: string;
	total: number;
	success: number;
	failure: number;
};

export type DashboardChart = {
	range: "24h" | "7d";
	points: DashboardChartPoint[];
};

export type DashboardActivity = {
	id: string;
	type: "beep" | "beeper" | "runner";
	title: string;
	status: "success" | "warning" | "failure" | "running" | "pending";
	summary: string;
	occurred_at: string;
	target_path: string;
};

export type DashboardResponse = {
	stats: DashboardStats;
	chart: DashboardChart;
	activities: DashboardActivity[];
};

export function fetchDashboard(
	slug: string,
	options?: {
		range?: "24h" | "7d";
		signal?: AbortSignal;
	},
) {
	const params = new URLSearchParams();
	if (options?.range) {
		params.set("range", options.range);
	}
	const query = params.toString() ? `?${params.toString()}` : "";
	return apiFetch<DashboardResponse>(`/api/v1/${slug}/dashboard${query}`, {
		method: "GET",
		signal: options?.signal,
	});
}
