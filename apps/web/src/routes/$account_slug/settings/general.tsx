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

			<div className="flex flex-1 flex-col gap-6">
				<div>
					<h1 className="type-title text-ink font-medium">
						{m.settings_general_title()}
					</h1>
					<p className="type-meta text-gray mt-1">
						{m.settings_general_description()}
					</p>
				</div>

				<TimezoneSettings
					slug={slug}
					accountTimezone={settings.timezone}
					accountTimezoneName={settings.timezone_name}
					onChanged={() => router.invalidate()}
				/>
			</div>
		</>
	);
}
