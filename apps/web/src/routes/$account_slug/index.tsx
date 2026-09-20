import {
	createFileRoute,
	getRouteApi,
	useNavigate,
	useRouter,
} from "@tanstack/react-router";

import { DashboardActivityFeed } from "@/components/dashboard/dashboard-activity-feed";
import { DashboardHeader } from "@/components/dashboard/dashboard-header";
import { DashboardKpiCards } from "@/components/dashboard/dashboard-kpi-cards";
import { DashboardQuickDialog } from "@/components/dashboard/dashboard-quick-dialog";
import { DashboardTrendChart } from "@/components/dashboard/dashboard-trend-chart";
import { fetchDashboard } from "@/lib/api/dashboard";
import { fetchSettings } from "@/lib/api/settings";
import { withAuthRedirects } from "@/lib/auth/guards";
import { m } from "@/locale/paraglide/messages";

type DashboardSearch = {
	range?: "24h" | "7d";
};

const accountRoute = getRouteApi("/$account_slug");

export const Route = createFileRoute("/$account_slug/")({
	validateSearch: (search: Record<string, unknown>): DashboardSearch => {
		const range = search.range === "24h" ? "24h" : "7d";
		return { range };
	},
	loaderDeps: ({ search }) => ({
		range: search.range ?? "7d",
	}),
	loader: withAuthRedirects(async ({ params, deps, abortController }) => {
		const slug = params?.account_slug ?? "";
		const [dashboardRes, settingsRes] = await Promise.all([
			fetchDashboard(slug, {
				range: deps.range,
				signal: abortController?.signal,
			}),
			fetchSettings(slug).catch(() => null),
		]);
		return {
			dashboard: dashboardRes,
			settings: settingsRes,
		};
	}),
	component: AccountHomePage,
});

function AccountHomePage() {
	const { me } = accountRoute.useRouteContext();
	const { account_slug: slug } = accountRoute.useParams();
	const { dashboard, settings } = Route.useLoaderData();
	const navigate = useNavigate({ from: Route.fullPath });
	const router = useRouter();
	const account = me.accounts.find((item) => item.slug === slug);

	async function handleCreated() {
		await router.invalidate();
	}

	function handleRangeChange(range: "24h" | "7d") {
		void navigate({
			search: (prev) => ({ ...prev, range }),
			replace: true,
		});
	}

	return (
		<>
			<DashboardHeader
				breadcrumbs={[{ label: m.nav_home(), isCurrentPage: true }]}
			/>

			<div className="flex flex-1 flex-col gap-6 p-4 md:p-6">
				<div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
					<div>
						<h1 className="font-heading text-2xl font-semibold tracking-tight">
							{account?.name ?? m.account_fallback_name()}
						</h1>
						<p className="mt-1 text-sm text-muted-foreground">
							Workspace overview, live system stability, and recent events.
						</p>
					</div>

					<div className="flex items-center gap-2">
						<DashboardQuickDialog
							slug={slug}
							defaultChannels={settings?.notification_channels}
							onCreated={handleCreated}
						/>
					</div>
				</div>

				<DashboardKpiCards stats={dashboard.stats} />

				<DashboardTrendChart
					chart={dashboard.chart}
					onRangeChange={handleRangeChange}
				/>

				<DashboardActivityFeed activities={dashboard.activities} slug={slug} />
			</div>
		</>
	);
}
