import { createFileRoute, Outlet, redirect } from "@tanstack/react-router";

import { DashboardShell } from "@/components/dashboard/dashboard-shell";
import { resolveShellAccount } from "@/lib/auth/account";
import { requireSession, requireStaff } from "@/lib/auth/guards";

export const Route = createFileRoute("/admin")({
	beforeLoad: ({ context, location }) => {
		const me = requireSession({ context, location });
		requireStaff(me);
		const account = resolveShellAccount(me);
		if (!account) {
			throw redirect({ to: "/sign" });
		}
		return { me, account };
	},
	component: AdminLayout,
});

function AdminLayout() {
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
