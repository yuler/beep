import { createFileRoute, getRouteApi } from "@tanstack/react-router";

import { DashboardHeader } from "@/components/dashboard/dashboard-header";
import { StatCard } from "@/components/dashboard/stat-card";
import { Badge } from "@/components/ui/badge";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { fetchAdminStats } from "@/lib/api/admin";
import { withAuthRedirects } from "@/lib/auth/guards";

const adminRoute = getRouteApi("/admin");

export const Route = createFileRoute("/admin/stats")({
	loader: withAuthRedirects(fetchAdminStats),
	component: StatsPage,
});

function StatsPage() {
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
					{ label: "Stats", isCurrentPage: true },
				]}
			/>

			<div className="flex flex-1 flex-col gap-6 p-4 md:p-6">
				<div>
					<h1 className="font-heading text-2xl font-semibold tracking-tight">
						Stats
					</h1>
					<p className="mt-1 text-sm text-muted-foreground">
						Account and identity growth from core.
					</p>
				</div>

				<section className="grid gap-4 sm:grid-cols-3">
					<StatCard label="Accounts total" value={data.accounts.total} />
					<StatCard
						label="Accounts · 7 days"
						value={data.accounts.last_7_days}
					/>
					<StatCard
						label="Accounts · 24 hours"
						value={data.accounts.last_24_hours}
					/>
				</section>

				<section className="grid gap-4 sm:grid-cols-3">
					<StatCard label="Identities total" value={data.identities.total} />
					<StatCard
						label="Identities · 7 days"
						value={data.identities.last_7_days}
					/>
					<StatCard
						label="Identities · 24 hours"
						value={data.identities.last_24_hours}
					/>
				</section>

				<Card>
					<CardHeader>
						<CardTitle>Recent accounts</CardTitle>
						<CardDescription>Newest 10 accounts.</CardDescription>
					</CardHeader>
					<CardContent className="flex flex-col gap-3">
						{data.recent_accounts.length === 0 ? (
							<p className="text-sm text-muted-foreground">No accounts.</p>
						) : (
							data.recent_accounts.map((accountRow) => (
								<div
									key={accountRow.id}
									className="flex items-center justify-between gap-3 rounded-lg border border-border px-3 py-2"
								>
									<div className="min-w-0">
										<p className="truncate font-medium">{accountRow.name}</p>
										<p className="truncate text-xs text-muted-foreground">
											/{accountRow.slug} ·{" "}
											{new Date(accountRow.created_at).toLocaleString()}
										</p>
									</div>
									{accountRow.personal ? (
										<Badge variant="secondary">Personal</Badge>
									) : (
										<Badge variant="outline">Team</Badge>
									)}
								</div>
							))
						)}
					</CardContent>
				</Card>
			</div>
		</>
	);
}
