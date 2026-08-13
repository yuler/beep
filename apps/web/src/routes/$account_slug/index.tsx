import { createFileRoute, getRouteApi } from "@tanstack/react-router";

import { ChatWidget } from "@/components/beautiful-ui/chat-widget";
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

			<div className="flex flex-1 flex-col gap-6 p-4 md:p-6">
				<div className="mx-auto flex w-full max-w-105 flex-col gap-4">
					<div>
						<h1 className="font-heading text-2xl font-semibold tracking-tight">
							{account?.name ?? "Account"}
						</h1>
						{import.meta.env.DEV ? (
							<p className="mt-1 text-sm text-muted-foreground">
								Ask about beeps, schedule reminders, or explore your history.
							</p>
						) : (
							<p className="mt-1 text-sm text-muted-foreground">
								/{slug} · identity and membership for this account.
							</p>
						)}
					</div>
				</div>

				{import.meta.env.DEV ? <ChatWidget /> : null}
			</div>
		</>
	);
}
