import { apiFetch } from "@/lib/api/client";

export type AdminStatsResponse = {
	accounts: {
		total: number;
		last_7_days: number;
		last_24_hours: number;
	};
	identities: {
		total: number;
		last_7_days: number;
		last_24_hours: number;
	};
	recent_accounts: Array<{
		id: string;
		name: string;
		slug: string;
		personal: boolean;
		created_at: string;
	}>;
};

export type AdminJobsResponse = {
	adapter: string;
	available: boolean;
	counts: {
		total: number;
		finished: number;
		pending: number;
		failed: number;
		ready: number;
		scheduled: number;
	} | null;
	recent: Array<{
		id: number | string;
		class_name: string;
		queue_name: string;
		finished_at: string | null;
		created_at: string;
		failed: boolean;
	}>;
};

export function fetchAdminStats() {
	return apiFetch<AdminStatsResponse>("/api/v1/admin/stats", {
		method: "GET",
	});
}

export function fetchAdminJobs() {
	return apiFetch<AdminJobsResponse>("/api/v1/admin/jobs", {
		method: "GET",
	});
}
