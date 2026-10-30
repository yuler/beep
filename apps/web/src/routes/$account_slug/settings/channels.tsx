import {
	createFileRoute,
	getRouteApi,
	useRouter,
} from "@tanstack/react-router";

import { DashboardHeader } from "@/components/dashboard/dashboard-header";
import { ChannelManagementSettings } from "@/components/settings/channel-management-settings";
import { NotificationChannelSettings } from "@/components/settings/notification-channel-settings";
import { WebPushSettings } from "@/components/settings/web-push-settings";
import { m } from "@/locale/paraglide/messages";

const accountRoute = getRouteApi("/$account_slug");
const settingsRoute = getRouteApi("/$account_slug/settings");

export const Route = createFileRoute("/$account_slug/settings/channels")({
	component: ChannelSettingsPage,
});

function ChannelSettingsPage() {
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
					{ label: m.settings_channels_title(), isCurrentPage: true },
				]}
			/>

			<div className="flex flex-1 flex-col gap-6">
				<div>
					<h1 className="type-title text-ink font-medium">
						{m.settings_channels_title()}
					</h1>
					<p className="type-meta text-gray mt-1">
						{m.settings_channels_description()}
					</p>
				</div>

				<NotificationChannelSettings
					slug={slug}
					channels={settings.notification_channels}
					onChanged={() => router.invalidate()}
				/>
				<ChannelManagementSettings slug={slug} />
				<WebPushSettings slug={slug} />
			</div>
		</>
	);
}
