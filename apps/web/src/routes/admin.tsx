import { createFileRoute, Outlet, redirect } from "@tanstack/react-router";

import { DashboardShell } from "@/components/dashboard/dashboard-shell";
import { AuthPending } from "@/components/layout";
import { resolveShellAccount } from "@/lib/auth/account";
import { requireSession, requireStaff } from "@/lib/auth/guards";

export const Route = createFileRoute("/admin")({
	ssr: false,
	pendingComponent: AuthPending,
	beforeLoad: async ({ context, location }) => {
		const me = await requireSession({ context, location });
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
