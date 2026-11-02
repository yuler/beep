import {
	createFileRoute,
	getRouteApi,
	useRouter,
} from "@tanstack/react-router";

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
				timezone={settings.timezone}
				timezoneSource={settings.timezone_source}
				onChanged={() => router.invalidate()}
			/>
		</div>
	);
}
