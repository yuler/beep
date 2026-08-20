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
						label: "Home",
						to: "/$account_slug",
						params: { account_slug: slug },
					},
					{ label: "Beeps", isCurrentPage: true },
				]}
			/>

			<div className="flex flex-1 flex-col gap-6 p-4 md:p-6">
				<div>
					<h1 className="font-heading text-2xl font-semibold tracking-tight">
						Beeps
					</h1>
					<p className="mt-1 text-sm text-muted-foreground">
						Reminders scheduled for this workspace.
					</p>
				</div>

				<BeepQuickCreate slug={slug} onCreated={handleCreated} />
				<BeepStats beeps={beeps} />
				<BeepList beeps={beeps} slug={slug} variant="full" />
			</div>
		</>
	);
}
