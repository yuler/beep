import { Link } from "@tanstack/react-router";
import { Fragment, type ReactNode } from "react";
import { LocaleSwitcher } from "@/components/layout/locale-switcher";
import { WebPushSetupBanner } from "@/components/settings/web-push-setup-banner";
import { ThemeToggle } from "@/components/theme-toggle";
import {
	Breadcrumb,
	BreadcrumbItem,
	BreadcrumbLink,
	BreadcrumbList,
	BreadcrumbPage,
	BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { Separator } from "@/components/ui/separator";
import { SidebarTrigger } from "@/components/ui/sidebar";

export type DashboardBreadcrumbItem = {
	label: string;
	to?:
		| "/$account_slug"
		| "/$account_slug/settings"
		| "/$account_slug/beeps"
		| "/$account_slug/beepers"
		| "/admin/jobs"
		| "/admin/stats"
		| "/my/access_tokens"
		| "/my/settings";
	params?: { account_slug: string };
	isCurrentPage?: boolean;
};

export function DashboardHeader({
	breadcrumbs,
	actions,
}: {
	breadcrumbs: DashboardBreadcrumbItem[];
	actions?: ReactNode;
}) {
	return (
		<>
			<header className="flex h-12 shrink-0 items-center gap-2 border-b transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12">
				<div className="flex w-full min-w-0 items-center gap-2 px-4 lg:px-6">
					<SidebarTrigger className="-ml-1 shrink-0" />
					<Separator
						orientation="vertical"
						className="mx-2 data-vertical:h-4 data-vertical:self-auto"
					/>

					<Breadcrumb className="min-w-0 flex-1">
						<BreadcrumbList>
							{breadcrumbs.map((item) => (
								<Fragment key={`${item.label}-${item.to ?? "current"}`}>
									{item !== breadcrumbs[0] ? (
										<BreadcrumbSeparator className="hidden md:block" />
									) : null}
									<BreadcrumbItem
										className={
											item !== breadcrumbs[breadcrumbs.length - 1]
												? "hidden md:block"
												: ""
										}
									>
										{item.isCurrentPage ? (
											<BreadcrumbPage>{item.label}</BreadcrumbPage>
										) : item.to && item.params ? (
											<BreadcrumbLink
												render={<Link to={item.to} params={item.params} />}
											>
												{item.label}
											</BreadcrumbLink>
										) : (
											item.label
										)}
									</BreadcrumbItem>
								</Fragment>
							))}
						</BreadcrumbList>
					</Breadcrumb>

					<div className="ml-auto flex shrink-0 items-center gap-2 pl-4">
						{actions}
						<LocaleSwitcher />
						<ThemeToggle />
					</div>
				</div>
			</header>
			<WebPushSetupBanner />
		</>
	);
}
