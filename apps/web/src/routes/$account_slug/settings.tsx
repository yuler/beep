import {
	createFileRoute,
	getRouteApi,
	useRouter,
} from "@tanstack/react-router";

import { DashboardHeader } from "@/components/dashboard/dashboard-header";
import { EmailChannelSettings } from "@/components/settings/email-channel-settings";
import { WebPushSettings } from "@/components/settings/web-push-settings";
import { fetchSettings } from "@/lib/api/settings";
import { withAuthRedirects } from "@/lib/auth/guards";

const accountRoute = getRouteApi("/$account_slug");

export const Route = createFileRoute("/$account_slug/settings")({
	loader: withAuthRedirects(({ params }) =>
		fetchSettings(params?.account_slug ?? ""),
	),
	component: SettingsPage,
});

function SettingsPage() {
	const { account_slug: slug } = accountRoute.useParams();
	const settings = Route.useLoaderData();
	const router = useRouter();

	return (
		<>
			<DashboardHeader
				breadcrumbs={[
					{
						label: "Home",
						to: "/$account_slug",
						params: { account_slug: slug },
					},
					{ label: "Settings", isCurrentPage: true },
				]}
			/>

			<div className="flex flex-1 flex-col gap-6 p-4 md:p-6">
				<div>
					<h1 className="font-heading text-2xl font-semibold tracking-tight">
						Settings
					</h1>
					<p className="mt-1 text-sm text-muted-foreground">
						Notifications and device preferences for this workspace.
					</p>
				</div>

				{settings.personal ? (
					<EmailChannelSettings
						slug={slug}
						enabled={settings.email_channel_enabled}
						onChanged={() => router.invalidate()}
					/>
				) : null}
				<WebPushSettings slug={slug} />
			</div>
		</>
	);
}
