import {
	createFileRoute,
	notFound,
	Outlet,
	redirect,
} from "@tanstack/react-router";

import { DashboardShell } from "@/components/dashboard/dashboard-shell";
import { AuthPending } from "@/components/layout";
import { resolveShellAccount } from "@/lib/auth/account";
import { requireSession } from "@/lib/auth/guards";

export const Route = createFileRoute("/dev")({
	pendingComponent: AuthPending,
	beforeLoad: async ({ context, location }) => {
		if (!import.meta.env.DEV) {
			throw notFound();
		}

		const me = await requireSession({ context, location });
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
