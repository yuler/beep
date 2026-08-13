import { createFileRoute, getRouteApi } from "@tanstack/react-router";

import { ChatComposer } from "@/components/beautiful-ui/chat-composer";
import { DashboardHeader } from "@/components/dashboard/dashboard-header";
import { withAuthRedirects } from "@/lib/auth/guards";

const accountRoute = getRouteApi("/$account_slug");

export const Route = createFileRoute("/$account_slug/")({
	loader: withAuthRedirects(({ params }) =>
		Promise.resolve({ slug: params?.account_slug ?? "" }),
	),
	component: AccountHomePage,
});

function AccountHomePage() {
	const { me } = accountRoute.useRouteContext();
	const { account_slug: slug } = accountRoute.useParams();
	const account = me.accounts.find((item) => item.slug === slug);

	return (
		<>
			<DashboardHeader breadcrumbs={[{ label: "Home", isCurrentPage: true }]} />

			<div className="bui flex flex-1 flex-col bg-canvas p-4 md:p-6">
				<div className="mx-auto flex w-full max-w-105 flex-1 flex-col gap-4">
					<div>
						<h1 className="font-heading text-2xl font-semibold tracking-tight text-ink">
							{account?.name ?? "Account"}
						</h1>
						<p className="mt-1 text-sm text-ink-2">
							Ask about beeps, schedule reminders, or explore your history.
						</p>
					</div>

					<div className="flex min-h-0 flex-1 flex-col">
						<ChatComposer />
					</div>
				</div>
			</div>
		</>
	);
}
