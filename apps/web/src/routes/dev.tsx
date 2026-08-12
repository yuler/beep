import {
	createFileRoute,
	notFound,
	Outlet,
	redirect,
} from "@tanstack/react-router";

import { DashboardShell } from "@/components/dashboard/dashboard-shell";
import { resolveShellAccount } from "@/lib/auth/account";
import { requireSession } from "@/lib/auth/guards";

export const Route = createFileRoute("/dev")({
	beforeLoad: ({ context, location }) => {
		if (!import.meta.env.DEV) {
			throw notFound();
		}

		const me = requireSession({ context, location });
		const account = resolveShellAccount(me);
		if (!account) {
			throw redirect({ to: "/sign" });
		}
		return { me, account };
	},
	component: DevLayout,
});

function DevLayout() {
	const { me, account } = Route.useRouteContext();

	return (
		<DashboardShell
			user={me.identity}
			accounts={me.accounts}
			slug={account.slug}
		>
			<Outlet />
		</DashboardShell>
	);
}
