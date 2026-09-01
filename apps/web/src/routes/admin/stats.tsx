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
import { m } from "@/locale/paraglide/messages";

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
						label: m.nav_home(),
						to: "/$account_slug",
						params: { account_slug: account.slug },
					},
					{ label: m.admin_admin() },
					{ label: m.admin_stats(), isCurrentPage: true },
				]}
			/>

			<div className="flex flex-1 flex-col gap-6 p-4 md:p-6">
				<div>
					<h1 className="font-heading text-2xl font-semibold tracking-tight">
						{m.admin_stats()}
					</h1>
					<p className="mt-1 text-sm text-muted-foreground">
						{m.admin_growth_description()}
					</p>
				</div>

				<section className="grid gap-4 sm:grid-cols-3">
					<StatCard
						label={m.admin_accounts_total()}
						value={data.accounts.total}
					/>
					<StatCard
						label={m.admin_accounts_7d()}
						value={data.accounts.last_7_days}
					/>
					<StatCard
						label={m.admin_accounts_24h()}
						value={data.accounts.last_24_hours}
					/>
				</section>

				<section className="grid gap-4 sm:grid-cols-3">
					<StatCard
						label={m.admin_identities_total()}
						value={data.identities.total}
					/>
					<StatCard
						label={m.admin_identities_7d()}
						value={data.identities.last_7_days}
					/>
					<StatCard
						label={m.admin_identities_24h()}
						value={data.identities.last_24_hours}
					/>
				</section>

				<Card>
					<CardHeader>
						<CardTitle>{m.admin_recent_accounts_title()}</CardTitle>
						<CardDescription>
							{m.admin_recent_accounts_description()}
						</CardDescription>
					</CardHeader>
					<CardContent className="flex flex-col gap-3">
						{data.recent_accounts.length === 0 ? (
							<p className="text-sm text-muted-foreground">
								{m.admin_no_accounts()}
							</p>
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
										<Badge variant="secondary">
											{m.account_type_personal()}
										</Badge>
									) : (
										<Badge variant="outline">{m.account_type_team()}</Badge>
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
