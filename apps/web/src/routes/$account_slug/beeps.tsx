import {
	createFileRoute,
	getRouteApi,
	useRouter,
} from "@tanstack/react-router";

import { BeepList } from "@/components/beeps/beep-list";
import { BeepQuickCreate } from "@/components/beeps/beep-quick-create";
import { BeepStats } from "@/components/beeps/beep-stats";
import { DashboardHeader } from "@/components/dashboard/dashboard-header";
import { fetchBeeps } from "@/lib/api/beeps";
import { withAuthRedirects } from "@/lib/auth/guards";
import { m } from "@/locale/paraglide/messages";

const accountRoute = getRouteApi("/$account_slug");

export const Route = createFileRoute("/$account_slug/beeps")({
	loader: withAuthRedirects(({ params }) =>
		fetchBeeps(params?.account_slug ?? ""),
	),
	component: BeepsPage,
});

function BeepsPage() {
	const { account_slug: slug } = accountRoute.useParams();
	const router = useRouter();
	const { beeps } = Route.useLoaderData();

	async function handleCreated() {
		await router.invalidate();
	}

	return (
		<>
			<DashboardHeader
				breadcrumbs={[
					{
						label: m.nav_home(),
						to: "/$account_slug",
						params: { account_slug: slug },
					},
					{ label: m.nav_beeps(), isCurrentPage: true },
				]}
			/>

			<div className="flex flex-1 flex-col gap-6 p-4 md:p-6">
				<div className="flex flex-col gap-1">
					<h1 className="font-heading text-2xl font-bold tracking-tight sm:text-3xl">
						{m.beeps_page_title()}
					</h1>
					<p className="text-sm text-muted-foreground">
						{m.beeps_manage_description()}
					</p>
				</div>

				<BeepQuickCreate slug={slug} onCreated={handleCreated} />
				<BeepStats beeps={beeps} />
				<BeepList beeps={beeps} slug={slug} variant="full" />
			</div>
		</>
	);
}
