import { createFileRoute, getRouteApi } from "@tanstack/react-router";
import { ExternalLink } from "lucide-react";

import { DashboardHeader } from "@/components/dashboard/dashboard-header";
import { StatCard } from "@/components/dashboard/stat-card";
import { buttonVariants } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { coreAppUrl } from "@/config";
import { fetchAdminJobs } from "@/lib/api/admin";
import { withAuthRedirects } from "@/lib/auth/guards";
import { useTranslation } from "@/lib/i18n";
import { cn } from "@/lib/utils";

const adminRoute = getRouteApi("/admin");

export const Route = createFileRoute("/admin/jobs")({
	loader: withAuthRedirects(fetchAdminJobs),
	component: JobsPage,
});

function jobStatusLabel(
	t: ReturnType<typeof useTranslation>["t"],
	job: { failed: boolean; finished_at: string | null },
): string {
	if (job.failed) return t("status.job.failed");
	if (job.finished_at) return t("status.job.finished");
	return t("status.job.pending");
}

function JobsPage() {
	const { t } = useTranslation();
	const { account } = adminRoute.useRouteContext();
	const data = Route.useLoaderData();
	const missionControlUrl = coreAppUrl("/admin/jobs");

	return (
		<>
			<DashboardHeader
				breadcrumbs={[
					{
						label: t("nav.home"),
						to: "/$account_slug",
						params: { account_slug: account.slug },
					},
					{ label: t("admin.admin") },
					{ label: t("admin.jobs"), isCurrentPage: true },
				]}
			/>

			<div className="flex flex-1 flex-col gap-6 p-4 md:p-6">
				<div className="flex flex-wrap items-start justify-between gap-3">
					<div>
						<h1 className="font-heading text-2xl font-semibold tracking-tight">
							{t("admin.jobs")}
						</h1>
						<p className="mt-1 text-sm text-muted-foreground">
							{t("admin.jobs_description")}
						</p>
					</div>
					{missionControlUrl ? (
						<a
							href={missionControlUrl}
							target="_blank"
							rel="noreferrer"
							className={cn(buttonVariants({ variant: "outline" }))}
						>
							{t("admin.jobs_open")}
							<ExternalLink data-icon="inline-end" />
						</a>
					) : null}
				</div>

				<div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
					<Card size="sm">
						<CardHeader>
							<CardDescription>{t("admin.adapter")}</CardDescription>
							<CardTitle className="font-mono text-lg">
								{data.adapter}
							</CardTitle>
						</CardHeader>
					</Card>
					{data.counts ? (
						<>
							<StatCard label={t("admin.queue_pending")} value={data.counts.pending} />
							<StatCard label={t("admin.queue_ready")} value={data.counts.ready} />
							<StatCard
								label={t("admin.queue_scheduled")}
								value={data.counts.scheduled}
							/>
							<StatCard label={t("admin.queue_failed")} value={data.counts.failed} />
							<StatCard
								label={t("admin.queue_finished")}
								value={data.counts.finished}
							/>
							<StatCard label={t("admin.queue_total")} value={data.counts.total} />
						</>
					) : (
						<Card size="sm" className="sm:col-span-2">
							<CardHeader>
								<CardTitle>{t("admin.queue_unavailable")}</CardTitle>
								<CardDescription>
									{data.available
										? t("admin.queue_counts_failed")
										: data.adapter === "solid_queue"
											? t("admin.queue_db_missing")
											: t("admin.queue_enable_solid", {
													adapter: data.adapter,
												})}
								</CardDescription>
							</CardHeader>
						</Card>
					)}
				</div>

				<Card>
					<CardHeader>
						<CardTitle>{t("admin.jobs_recent_title")}</CardTitle>
						<CardDescription>{t("admin.jobs_recent_description")}</CardDescription>
					</CardHeader>
					<CardContent>
						{data.recent.length === 0 ? (
							<p className="text-sm text-muted-foreground">
								{t("admin.jobs_empty")}
							</p>
						) : (
							<div className="overflow-x-auto">
								<table className="w-full text-left text-sm">
									<thead className="border-b text-muted-foreground">
										<tr>
											<th className="px-2 py-2 font-medium">
												{t("admin.jobs_class")}
											</th>
											<th className="px-2 py-2 font-medium">
												{t("admin.jobs_queue")}
											</th>
											<th className="px-2 py-2 font-medium">
												{t("admin.jobs_status")}
											</th>
											<th className="px-2 py-2 font-medium">
												{t("admin.jobs_created")}
											</th>
										</tr>
									</thead>
									<tbody>
										{data.recent.map((job) => (
											<tr
												key={String(job.id)}
												className="border-b last:border-0"
											>
												<td className="px-2 py-2 font-mono text-xs">
													{job.class_name}
												</td>
												<td className="px-2 py-2">{job.queue_name}</td>
												<td className="px-2 py-2">
													{jobStatusLabel(t, job)}
												</td>
												<td className="px-2 py-2 text-muted-foreground">
													{new Date(job.created_at).toLocaleString()}
												</td>
											</tr>
										))}
									</tbody>
								</table>
							</div>
						)}
					</CardContent>
				</Card>
			</div>
		</>
	);
}
