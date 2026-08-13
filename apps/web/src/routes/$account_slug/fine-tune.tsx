import { createFileRoute, getRouteApi } from "@tanstack/react-router";

import { FineTuneWorkspace } from "@/components/beautiful-ui/fine-tune-card";
import { DashboardHeader } from "@/components/dashboard/dashboard-header";
import { withAuthRedirects } from "@/lib/auth/guards";

const accountRoute = getRouteApi("/$account_slug");

export const Route = createFileRoute("/$account_slug/fine-tune")({
	loader: withAuthRedirects(() => Promise.resolve({})),
	component: FineTunePage,
});

function FineTunePage() {
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
					{ label: "Fine-tune", isCurrentPage: true },
				]}
			/>

			<div className="bui flex flex-1 flex-col gap-4 p-4 md:p-6">
				<div>
					<h1 className="font-heading text-2xl font-semibold tracking-tight text-ink">
						Fine-tune card
					</h1>
					<p className="mt-1 text-sm text-ink-2">
						Inspector controls from Beautiful UI #18 — scrub numbers, pick
						layout, and choose a type.
					</p>
				</div>

				<FineTuneWorkspace />
			</div>
		</>
	);
}
