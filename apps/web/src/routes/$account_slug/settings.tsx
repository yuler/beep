import {
	createFileRoute,
	Outlet,
	useNavigate,
	useRouterState,
} from "@tanstack/react-router";

import { Segmented } from "@/components/ui/segmented";
import { fetchSettings } from "@/lib/api/settings";
import { withAuthRedirects } from "@/lib/auth/guards";
import { m } from "@/locale/paraglide/messages";

export const Route = createFileRoute("/$account_slug/settings")({
	loader: withAuthRedirects(({ params }) =>
		fetchSettings(params?.account_slug ?? ""),
	),
	component: SettingsLayout,
});

function SettingsLayout() {
	const { account_slug: slug } = Route.useParams();
	const navigate = useNavigate();
	const pathname = useRouterState({ select: (s) => s.location.pathname });
	const currentTab = pathname.includes("/channels") ? "channels" : "general";

	return (
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
	);
}
