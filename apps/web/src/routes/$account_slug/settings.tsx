import {
	createFileRoute,
	getRouteApi,
	Outlet,
	useNavigate,
	useRouterState,
} from "@tanstack/react-router";

import { DashboardHeader } from "@/components/dashboard/dashboard-header";
import { Segmented } from "@/components/ui/segmented";
import { fetchSettings } from "@/lib/api/settings";
import { withAuthRedirects } from "@/lib/auth/guards";
import { m } from "@/locale/paraglide/messages";

const accountRoute = getRouteApi("/$account_slug");

export const Route = createFileRoute("/$account_slug/settings")({
	loader: withAuthRedirects(({ params }) =>
		fetchSettings(params?.account_slug ?? ""),
	),
	component: SettingsLayout,
});

function SettingsLayout() {
	const { account_slug: slug } = accountRoute.useParams();
	const navigate = useNavigate();
	const pathname = useRouterState({ select: (s) => s.location.pathname });
	const currentTab = pathname.includes("/channels") ? "channels" : "general";
	const currentTabLabel =
		currentTab === "channels"
			? m.settings_channels_title()
			: m.settings_general_title();

	return (
		<>
			<DashboardHeader
				breadcrumbs={[
					{
						label: m.nav_home(),
						to: "/$account_slug",
						params: { account_slug: slug },
					},
					{ label: m.nav_settings() },
					{ label: currentTabLabel, isCurrentPage: true },
				]}
			/>

			<div className="flex flex-col gap-6">
				<div className="w-full max-w-sm">
					<Segmented
						value={currentTab}
						onChange={(val) => {
							void navigate({
								to: `/$account_slug/settings/${val}`,
								params: { account_slug: slug },
							});
						}}
						items={[
							{ value: "general", label: m.settings_general_title() },
							{ value: "channels", label: m.settings_channels_title() },
						]}
					/>
				</div>
				<Outlet />
			</div>
		</>
	);
}
