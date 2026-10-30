import { Link } from "@tanstack/react-router";
import { Fragment, type ReactNode } from "react";
import { WebPushSetupBanner } from "@/components/settings/web-push-setup-banner";
import {
	Breadcrumb,
	BreadcrumbItem,
	BreadcrumbLink,
	BreadcrumbList,
	BreadcrumbPage,
	BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";

export type DashboardBreadcrumbItem = {
	label: string;
	to?:
		| "/$account_slug"
		| "/$account_slug/settings"
		| "/$account_slug/settings/general"
		| "/$account_slug/settings/channels"
		| "/$account_slug/settings/notifications"
		| "/$account_slug/beeps"
		| "/$account_slug/beepers"
		| "/$account_slug/runners"
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
		<div className="mb-6 flex flex-col gap-4">
			<div className="flex flex-wrap items-center justify-between gap-4">
				<Breadcrumb className="min-w-0">
					<BreadcrumbList className="type-meta text-gray">
						{breadcrumbs.map((item) => (
							<Fragment key={`${item.label}-${item.to ?? "current"}`}>
								{item !== breadcrumbs[0] ? (
									<BreadcrumbSeparator className="text-line" />
								) : null}
								<BreadcrumbItem>
									{item.isCurrentPage ? (
										<BreadcrumbPage className="type-app-title font-medium text-ink">
											{item.label}
										</BreadcrumbPage>
									) : item.to && item.params ? (
										<BreadcrumbLink
											render={<Link to={item.to} params={item.params} />}
											className="hover:text-ink transition-colors"
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

				{actions ? (
					<div className="flex shrink-0 items-center gap-2">{actions}</div>
				) : null}
			</div>

			<WebPushSetupBanner />
		</div>
	);
}
