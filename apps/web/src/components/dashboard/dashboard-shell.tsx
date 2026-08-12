import type { CSSProperties, ReactNode } from "react";

import { DashboardSidebar } from "@/components/dashboard/dashboard-sidebar";
import { SidebarInset, SidebarProvider } from "@/components/ui/sidebar";
import { TooltipProvider } from "@/components/ui/tooltip";
import type { MeResponse } from "@/lib/api/session";

export function DashboardShell({
	user,
	accounts,
	slug,
	children,
}: {
	user: MeResponse["identity"] | null;
	accounts: MeResponse["accounts"];
	slug: string;
	children: ReactNode;
}) {
	return (
		<TooltipProvider>
			<SidebarProvider
				className="min-h-svh"
				style={
					{
						"--sidebar-width": "calc(var(--spacing) * 64)",
						"--header-height": "calc(var(--spacing) * 12)",
					} as CSSProperties
				}
			>
				<DashboardSidebar user={user} accounts={accounts} slug={slug} />
				<SidebarInset>{children}</SidebarInset>
			</SidebarProvider>
		</TooltipProvider>
	);
}
