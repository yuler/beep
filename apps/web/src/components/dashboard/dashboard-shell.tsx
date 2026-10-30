import { Link, useRouterState } from "@tanstack/react-router";
import type { ReactNode } from "react";
import { AccountSwitcher } from "@/components/dashboard/account-switcher";
import { SidebarUserMenu } from "@/components/dashboard/sidebar-user-menu";
import { LocaleSwitcher } from "@/components/layout/locale-switcher";
import { LogoMark } from "@/components/logo-mark";
import { ThemeToggle } from "@/components/theme-toggle";
import { TooltipProvider } from "@/components/ui/tooltip";
import type { MeResponse } from "@/lib/api/session";
import { cn } from "@/lib/utils";
import { m } from "@/locale/paraglide/messages";

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
	const pathname = useRouterState({ select: (s) => s.location.pathname });

	const isAdmin = pathname.startsWith("/admin");
	const isMy = pathname.startsWith("/my");
	const isDev = pathname.startsWith("/dev");

	const tabBaseClass =
		"type-kicker uppercase tracking-[0.18em] py-3 text-xs inline-flex items-center whitespace-nowrap transition-colors border-b-2 -mb-px";
	const tabInactiveClass = "border-transparent text-gray hover:text-ink";
	const tabActiveClass = "border-ink text-ink font-medium";

	return (
		<TooltipProvider>
			<div className="flex min-h-svh flex-col bg-ground text-ink">
				{/* Top AppBar */}
				<header className="sticky top-0 z-50 h-14 md:h-16 border-b border-line bg-ground">
					<div className="mx-auto flex h-full max-w-[1080px] items-center justify-between px-4 md:px-6 lg:px-8">
						{/* Left: LogoMark + Beep name + Account Switcher */}
						<div className="flex items-center gap-2.5 sm:gap-4">
							<Link
								to="/"
								aria-label={m.term_beep()}
								className="inline-flex items-center gap-2 text-ink hover:opacity-80 transition-opacity"
							>
								<LogoMark className="size-5 text-ink" />
								<span className="type-kicker font-medium tracking-[0.18em] uppercase hidden sm:inline">
									{m.term_beep()}
								</span>
							</Link>
							<div className="h-4 w-px bg-line" />
							<AccountSwitcher accounts={accounts} slug={slug} />
						</div>

						{/* Right: LocaleSwitcher + ThemeToggle + UserMenu */}
						<div className="flex items-center gap-1 sm:gap-2">
							<LocaleSwitcher />
							<ThemeToggle />
							{user && <SidebarUserMenu user={user} currentSlug={slug} />}
						</div>
					</div>
				</header>

				{/* Subnav row: sticky top-14 md:top-16 z-40 bg-ground border-b border-line */}
				<nav className="sticky top-14 md:top-16 z-40 bg-ground border-b border-line">
					<div className="mx-auto max-w-[1080px] px-4 md:px-6 lg:px-8">
						<div className="flex items-center gap-6 overflow-x-auto no-scrollbar">
							{isAdmin ? (
								<>
									<Link
										to="/$account_slug"
										params={{ account_slug: slug }}
										className={cn(tabBaseClass, tabInactiveClass)}
									>
										← {m.nav_back_to_workspace()}
									</Link>
									<Link
										to="/admin/jobs"
										className={cn(
											tabBaseClass,
											pathname === "/admin/jobs"
												? tabActiveClass
												: tabInactiveClass,
										)}
									>
										{m.nav_jobs()}
									</Link>
									<Link
										to="/admin/stats"
										className={cn(
											tabBaseClass,
											pathname === "/admin/stats"
												? tabActiveClass
												: tabInactiveClass,
										)}
									>
										{m.nav_stats()}
									</Link>
								</>
							) : isMy ? (
								<>
									<Link
										to="/$account_slug"
										params={{ account_slug: slug }}
										className={cn(tabBaseClass, tabInactiveClass)}
									>
										← {m.nav_back_to_workspace()}
									</Link>
									<Link
										to="/my/settings"
										className={cn(
											tabBaseClass,
											pathname === "/my/settings"
												? tabActiveClass
												: tabInactiveClass,
										)}
									>
										{m.my_profile_settings()}
									</Link>
									<Link
										to="/my/access_tokens"
										className={cn(
											tabBaseClass,
											pathname === "/my/access_tokens"
												? tabActiveClass
												: tabInactiveClass,
										)}
									>
										{m.nav_api_tokens()}
									</Link>
								</>
							) : isDev ? (
								<>
									<Link
										to="/$account_slug"
										params={{ account_slug: slug }}
										className={cn(tabBaseClass, tabInactiveClass)}
									>
										← {m.nav_back_to_workspace()}
									</Link>
									<Link
										to="/dev/letters"
										className={cn(
											tabBaseClass,
											pathname === "/dev/letters"
												? tabActiveClass
												: tabInactiveClass,
										)}
									>
										{m.nav_letters()}
									</Link>
								</>
							) : (
								<>
									<Link
										to="/$account_slug"
										params={{ account_slug: slug }}
										activeOptions={{ exact: true }}
										className={cn(
											tabBaseClass,
											pathname === `/${slug}`
												? tabActiveClass
												: tabInactiveClass,
										)}
									>
										{m.nav_home()}
									</Link>
									<Link
										to="/$account_slug/beeps"
										params={{ account_slug: slug }}
										className={cn(
											tabBaseClass,
											pathname.startsWith(`/${slug}/beeps`)
												? tabActiveClass
												: tabInactiveClass,
										)}
									>
										{m.nav_beeps()}
									</Link>
									<Link
										to="/$account_slug/beepers"
										params={{ account_slug: slug }}
										className={cn(
											tabBaseClass,
											pathname.startsWith(`/${slug}/beepers`)
												? tabActiveClass
												: tabInactiveClass,
										)}
									>
										{m.nav_beepers()}
									</Link>
									<Link
										to="/$account_slug/runners"
										params={{ account_slug: slug }}
										className={cn(
											tabBaseClass,
											pathname.startsWith(`/${slug}/runners`)
												? tabActiveClass
												: tabInactiveClass,
										)}
									>
										{m.nav_runners()}
									</Link>
									<Link
										to="/$account_slug/settings"
										params={{ account_slug: slug }}
										className={cn(
											tabBaseClass,
											pathname.startsWith(`/${slug}/settings`)
												? tabActiveClass
												: tabInactiveClass,
										)}
									>
										{m.nav_settings()}
									</Link>
								</>
							)}
						</div>
					</div>
				</nav>

				{/* Main page content */}
				<main className="flex-1 w-full mx-auto max-w-[1080px] px-4 py-6 md:px-6 lg:px-8">
					{children}
				</main>
			</div>
		</TooltipProvider>
	);
}
