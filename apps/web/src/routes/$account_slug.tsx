import { createFileRoute, notFound, Outlet } from "@tanstack/react-router";
import { useEffect } from "react";

import { DashboardShell } from "@/components/dashboard/dashboard-shell";
import { rememberLastAccount } from "@/lib/api/session";
import { requireSession } from "@/lib/auth/guards";
import { isAccountSlug } from "@/lib/auth/slugs";

export const Route = createFileRoute("/$account_slug")({
	beforeLoad: ({ context, location, params }) => {
		if (!isAccountSlug(params.account_slug)) {
			throw notFound();
		}

		const me = requireSession({ context, location });
		const account = me.accounts.find(
			(item) => item.slug === params.account_slug,
		);
		if (!account) {
			throw notFound();
		}
		return { me, account };
	},
	component: AccountLayout,
});

function AccountLayout() {
	const { account_slug } = Route.useParams();
	const { me, account } = Route.useRouteContext();

	useEffect(() => {
		void rememberLastAccount(account.slug).catch(() => {
			// Picker hint is best-effort; ignore network failures.
		});
	}, [account.slug]);

	return (
		<DashboardShell
			user={me.identity}
			accounts={me.accounts}
			slug={account_slug}
		>
			<Outlet />
		</DashboardShell>
	);
}
