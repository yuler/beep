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
import { cn } from "@/lib/utils";

const adminRoute = getRouteApi("/admin");

export const Route = createFileRoute("/admin/jobs")({
	loader: withAuthRedirects(fetchAdminJobs),
	component: JobsPage,
});

function JobsPage() {
	const { account } = adminRoute.useRouteContext();
	const data = Route.useLoaderData();

	return (
		<>
			<DashboardHeader
				breadcrumbs={[
					{
						label: "Home",
						to: "/$account_slug",
						params: { account_slug: account.slug },
					},
					{ label: "Admin" },
					{ label: "Jobs", isCurrentPage: true },
				]}
			/>

			<div className="flex flex-1 flex-col gap-6 p-4 md:p-6">
				<div className="flex flex-wrap items-start justify-between gap-3">
					<div>
						<h1 className="font-heading text-2xl font-semibold tracking-tight">
							Jobs
						</h1>
						<p className="mt-1 text-sm text-muted-foreground">
							Background job queue status from core.
						</p>
					</div>
					<a
						href={coreAppUrl("/admin/jobs")}
						target="_blank"
						rel="noreferrer"
						className={cn(buttonVariants({ variant: "outline" }))}
					>
						Open jobs
						<ExternalLink data-icon="inline-end" />
					</a>
				</div>

				<div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
					<Card size="sm">
						<CardHeader>
							<CardDescription>Adapter</CardDescription>
							<CardTitle className="font-mono text-lg">
								{data.adapter}
							</CardTitle>
						</CardHeader>
					</Card>
					{data.counts ? (
						<>
							<StatCard label="Pending" value={data.counts.pending} />
							<StatCard label="Ready" value={data.counts.ready} />
							<StatCard label="Scheduled" value={data.counts.scheduled} />
							<StatCard label="Failed" value={data.counts.failed} />
							<StatCard label="Finished" value={data.counts.finished} />
							<StatCard label="Total" value={data.counts.total} />
						</>
					) : (
						<Card size="sm" className="sm:col-span-2">
							<CardHeader>
								<CardTitle>Solid Queue unavailable</CardTitle>
								<CardDescription>
									{data.available
										? "Could not load queue counts."
										: data.adapter === "solid_queue"
											? "Queue database is missing tables — run bin/rails db:prepare in core."
											: `Active Job adapter is “${data.adapter}”. Enable Solid Queue in this environment to see counts.`}
								</CardDescription>
							</CardHeader>
						</Card>
					)}
				</div>

				<Card>
					<CardHeader>
						<CardTitle>Recent jobs</CardTitle>
						<CardDescription>Latest 20 jobs from the queue.</CardDescription>
					</CardHeader>
					<CardContent>
						{data.recent.length === 0 ? (
							<p className="text-sm text-muted-foreground">No jobs yet.</p>
						) : (
							<div className="overflow-x-auto">
								<table className="w-full text-left text-sm">
									<thead className="border-b text-muted-foreground">
										<tr>
											<th className="px-2 py-2 font-medium">Class</th>
											<th className="px-2 py-2 font-medium">Queue</th>
											<th className="px-2 py-2 font-medium">Status</th>
											<th className="px-2 py-2 font-medium">Created</th>
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
													{job.failed
														? "Failed"
														: job.finished_at
															? "Finished"
															: "Pending"}
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
