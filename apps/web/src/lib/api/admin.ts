import {
	fetchAdminJobs as serverFetchAdminJobs,
	fetchAdminStats as serverFetchAdminStats,
} from "@/server/admin";

export type { AdminJobsResponse, AdminStatsResponse } from "@/server/admin";

export function fetchAdminStats() {
	return serverFetchAdminStats({});
}

export function fetchAdminJobs() {
	return serverFetchAdminJobs({});
}
