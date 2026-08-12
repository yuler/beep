import { createFileRoute, Outlet } from "@tanstack/react-router";

import { AuthLayout } from "@/components/layout";
import { requireGuest } from "@/lib/auth/guards";
import { parseSignSearch } from "@/lib/auth/return-to";

export const Route = createFileRoute("/sign")({
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
