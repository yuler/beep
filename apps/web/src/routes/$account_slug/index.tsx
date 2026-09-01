import {
	createFileRoute,
	getRouteApi,
	Link,
	useRouter,
} from "@tanstack/react-router";

import { BeepList } from "@/components/beeps/beep-list";
import { BeepQuickCreate } from "@/components/beeps/beep-quick-create";
import { BeepStats } from "@/components/beeps/beep-stats";
import { DashboardHeader } from "@/components/dashboard/dashboard-header";
import { fetchBeeps } from "@/lib/api/beeps";
import { withAuthRedirects } from "@/lib/auth/guards";
import { upcomingBeeps } from "@/lib/beep-stats";
import * as m from "@/locale/paraglide/messages";

const accountRoute = getRouteApi("/$account_slug");

export const Route = createFileRoute("/$account_slug/")({
	loader: withAuthRedirects(({ params }) =>
		fetchBeeps(params?.account_slug ?? ""),
	),
	component: AccountHomePage,
});

function AccountHomePage() {
	const { me } = accountRoute.useRouteContext();
	const { account_slug: slug } = accountRoute.useParams();
	const { beeps } = Route.useLoaderData();
	const router = useRouter();
	const account = me.accounts.find((item) => item.slug === slug);
	const upcoming = upcomingBeeps(beeps);

	async function handleCreated() {
		await router.invalidate();
	}

	return (
		<>
			<DashboardHeader
				breadcrumbs={[{ label: m.nav_home(), isCurrentPage: true }]}
			/>

			<div className="flex flex-1 flex-col gap-6 p-4 md:p-6">
				<div>
					<h1 className="font-heading text-2xl font-semibold tracking-tight">
						{account?.name ?? m.account_fallback_name()}
					</h1>
					<p className="mt-1 text-sm text-muted-foreground">
						{m.beeps_create_description()}
					</p>
				</div>

				<BeepQuickCreate slug={slug} onCreated={handleCreated} />
				<BeepStats beeps={beeps} />

				<div className="flex flex-col gap-3">
					<div className="flex items-center justify-between gap-3">
						<h2 className="font-heading text-lg font-semibold tracking-tight">
							{m.beeps_upcoming()}
						</h2>
						<Link
							to="/$account_slug/beeps"
							params={{ account_slug: slug }}
							className="text-xs font-medium text-muted-foreground hover:text-foreground"
						>
							{m.beeps_view_all()}
						</Link>
					</div>
					<BeepList beeps={upcoming} slug={slug} variant="compact" />
				</div>
			</div>
		</>
	);
}
