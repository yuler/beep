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
import { useTranslation } from "@/lib/i18n";

const adminRoute = getRouteApi("/admin");

export const Route = createFileRoute("/admin/stats")({
	loader: withAuthRedirects(fetchAdminStats),
	component: StatsPage,
});

function StatsPage() {
	const { t } = useTranslation();
	const { account } = adminRoute.useRouteContext();
	const data = Route.useLoaderData();

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
					{ label: t("admin.stats"), isCurrentPage: true },
				]}
			/>

			<div className="flex flex-1 flex-col gap-6 p-4 md:p-6">
				<div>
					<h1 className="font-heading text-2xl font-semibold tracking-tight">
						{t("admin.stats")}
					</h1>
					<p className="mt-1 text-sm text-muted-foreground">
						{t("admin.growth_description")}
					</p>
				</div>

				<section className="grid gap-4 sm:grid-cols-3">
					<StatCard label={t("admin.accounts_total")} value={data.accounts.total} />
					<StatCard
						label={t("admin.accounts_7d")}
						value={data.accounts.last_7_days}
					/>
					<StatCard
						label={t("admin.accounts_24h")}
						value={data.accounts.last_24_hours}
					/>
				</section>

				<section className="grid gap-4 sm:grid-cols-3">
					<StatCard
						label={t("admin.identities_total")}
						value={data.identities.total}
					/>
					<StatCard
						label={t("admin.identities_7d")}
						value={data.identities.last_7_days}
					/>
					<StatCard
						label={t("admin.identities_24h")}
						value={data.identities.last_24_hours}
					/>
				</section>

				<Card>
					<CardHeader>
						<CardTitle>{t("admin.recent_accounts_title")}</CardTitle>
						<CardDescription>
							{t("admin.recent_accounts_description")}
						</CardDescription>
					</CardHeader>
					<CardContent className="flex flex-col gap-3">
						{data.recent_accounts.length === 0 ? (
							<p className="text-sm text-muted-foreground">
								{t("admin.no_accounts")}
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
											{t("account.type_personal")}
										</Badge>
									) : (
										<Badge variant="outline">{t("account.type_team")}</Badge>
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
