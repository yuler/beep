import type { ReactNode } from "react";

import { SiteFooter } from "@/components/layout/site-footer";
import { SiteHeader } from "@/components/layout/site-header";

export function SiteLayout({ children }: { children: ReactNode }) {
	return (
		<div className="flex min-h-svh flex-col bg-background text-foreground">
			<SiteHeader />
			{children}
			<SiteFooter />
		</div>
	);
}
