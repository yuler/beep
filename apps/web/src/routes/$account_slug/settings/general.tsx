import {
	createFileRoute,
	getRouteApi,
	useRouter,
} from "@tanstack/react-router";

import { DashboardHeader } from "@/components/dashboard/dashboard-header";
import { TimezoneSettings } from "@/components/settings/timezone-settings";
import { m } from "@/locale/paraglide/messages";

const accountRoute = getRouteApi("/$account_slug");
const settingsRoute = getRouteApi("/$account_slug/settings");

export const Route = createFileRoute("/$account_slug/settings/general")({
	component: GeneralSettingsPage,
});

function GeneralSettingsPage() {
	const { account_slug: slug } = accountRoute.useParams();
	const settings = settingsRoute.useLoaderData();
	const router = useRouter();

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
					{ label: m.settings_general_title(), isCurrentPage: true },
				]}
			/>

			<div className="flex flex-1 flex-col gap-6 p-4 md:p-6">
				<div>
					<h1 className="font-heading text-2xl font-semibold tracking-tight">
						{m.settings_general_title()}
					</h1>
					<p className="mt-1 text-sm text-muted-foreground">
						{m.settings_general_description()}
					</p>
				</div>

				<TimezoneSettings
					slug={slug}
					timezone={settings.timezone}
					timezoneSource={settings.timezone_source}
					onChanged={() => router.invalidate()}
				/>
			</div>
		</>
	);
}
