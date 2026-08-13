import { createFileRoute, Outlet } from "@tanstack/react-router";

import { AuthLayout, AuthPending } from "@/components/layout";
import { requireGuest } from "@/lib/auth/guards";
import { parseSignSearch } from "@/lib/auth/return-to";

export const Route = createFileRoute("/sign")({
	ssr: false,
	pendingComponent: AuthPending,
	validateSearch: parseSignSearch,
	beforeLoad: requireGuest,
	component: SignLayout,
});

function SignLayout() {
	return (
		<AuthLayout>
			<Outlet />
		</AuthLayout>
	);
}
