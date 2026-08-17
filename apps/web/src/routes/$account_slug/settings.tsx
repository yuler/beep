import { createFileRoute, getRouteApi } from "@tanstack/react-router";

import { DashboardHeader } from "@/components/dashboard/dashboard-header";
import { WebPushSettings } from "@/components/settings/web-push-settings";
import { withAuthRedirects } from "@/lib/auth/guards";

const accountRoute = getRouteApi("/$account_slug");

export const Route = createFileRoute("/$account_slug/settings")({
	loader: withAuthRedirects(({ params }) =>
		Promise.resolve({ slug: params?.account_slug ?? "" }),
	),
	component: SettingsPage,
});

function SettingsPage() {
	const { account_slug: slug } = accountRoute.useParams();

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

				<WebPushSettings slug={slug} />
			</div>
		</>
	);
}
