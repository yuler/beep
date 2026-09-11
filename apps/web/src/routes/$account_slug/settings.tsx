import { createFileRoute, Outlet } from "@tanstack/react-router";

import { fetchSettings } from "@/lib/api/settings";
import { withAuthRedirects } from "@/lib/auth/guards";

export const Route = createFileRoute("/$account_slug/settings")({
	loader: withAuthRedirects(({ params }) =>
		fetchSettings(params?.account_slug ?? ""),
	),
	component: SettingsLayout,
});

function SettingsLayout() {
	return <Outlet />;
}
