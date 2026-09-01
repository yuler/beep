import { Link, useRouterState } from "@tanstack/react-router";
import {
	Activity,
	Bell,
	BriefcaseBusiness,
	KeyRound,
	LayoutDashboard,
	Mail,
	Settings,
	User,
} from "lucide-react";
import { AccountSwitcher } from "@/components/dashboard/account-switcher";
import { SidebarUserMenu } from "@/components/dashboard/sidebar-user-menu";
import { LogoMark } from "@/components/logo-mark";
import {
	Sidebar,
	SidebarContent,
	SidebarFooter,
	SidebarGroup,
	SidebarGroupContent,
	SidebarGroupLabel,
	SidebarHeader,
	SidebarMenu,
	SidebarMenuButton,
	SidebarMenuItem,
	useSidebar,
} from "@/components/ui/sidebar";
import type { MeResponse } from "@/lib/api/session";
import { m } from "@/locale/paraglide/messages";

type DashboardSidebarProps = React.ComponentProps<typeof Sidebar> & {
	user: MeResponse["identity"] | null;
	accounts: MeResponse["accounts"];
	slug: string;
};

export function DashboardSidebar({
	user,
	accounts,
	slug,
	...props
}: DashboardSidebarProps) {
	const pathname = useRouterState({ select: (s) => s.location.pathname });
	const { isMobile, setOpenMobile } = useSidebar();

	const closeMobileSidebar = () => {
		if (isMobile) setOpenMobile(false);
	};

	const isMy = pathname.startsWith("/my");
	const homePath = `/${slug}`;
	const beepsPath = `/${slug}/beeps`;
	const settingsPath = `/${slug}/settings`;
	const lettersPath = "/dev/letters";
	const jobsPath = "/admin/jobs";
	const statsPath = "/admin/stats";
	const mySettingsPath = "/my/settings";
	const accessTokensPath = "/my/access_tokens";

	return (
		<Sidebar collapsible="icon" variant="inset" {...props}>
			<SidebarHeader>
				{isMy ? (
					<SidebarMenu>
						<SidebarMenuItem>
							<SidebarMenuButton
								size="lg"
								render={
									<Link
										to="/$account_slug"
										params={{ account_slug: slug }}
										onClick={closeMobileSidebar}
									/>
								}
								tooltip={m.nav_back_to_workspace()}
								className="group-data-[collapsible=icon]:size-8! group-data-[collapsible=icon]:p-0!"
							>
								<span className="flex aspect-square size-8 shrink-0 items-center justify-center rounded-lg bg-foreground text-background">
									<LogoMark className="size-4" />
								</span>
								<div className="grid flex-1 text-left text-sm leading-tight">
									<span className="truncate font-semibold">
										{m.term_beep_capitalized()}
									</span>
									<span className="truncate text-xs text-muted-foreground">
										{m.nav_personal_settings()}
									</span>
								</div>
							</SidebarMenuButton>
						</SidebarMenuItem>
					</SidebarMenu>
				) : (
					<AccountSwitcher accounts={accounts} slug={slug} />
				)}
			</SidebarHeader>

			<SidebarContent>
				{isMy ? (
					<SidebarGroup>
						<SidebarGroupLabel>{m.nav_personal_settings()}</SidebarGroupLabel>
						<SidebarGroupContent>
							<SidebarMenu>
								<SidebarMenuItem>
									<SidebarMenuButton
										isActive={
											pathname === mySettingsPath ||
											pathname.startsWith(`${mySettingsPath}/`)
										}
										tooltip={m.nav_profile_preferences()}
										render={
											<Link to="/my/settings" onClick={closeMobileSidebar} />
										}
									>
										<User />
										<span>{m.nav_settings()}</span>
									</SidebarMenuButton>
								</SidebarMenuItem>
								<SidebarMenuItem>
									<SidebarMenuButton
										isActive={
											pathname === accessTokensPath ||
											pathname.startsWith(`${accessTokensPath}/`)
										}
										tooltip={m.nav_api_access_tokens()}
										render={
											<Link
												to="/my/access_tokens"
												onClick={closeMobileSidebar}
											/>
										}
									>
										<KeyRound />
										<span>{m.nav_api_tokens()}</span>
									</SidebarMenuButton>
								</SidebarMenuItem>
							</SidebarMenu>
						</SidebarGroupContent>
					</SidebarGroup>
				) : (
					<>
						<SidebarGroup>
							<SidebarGroupContent>
								<SidebarMenu>
									<SidebarMenuItem>
										<SidebarMenuButton
											isActive={
												pathname === homePath || pathname === `${homePath}/`
											}
											tooltip={m.nav_home()}
											render={
												<Link
													to="/$account_slug"
													params={{ account_slug: slug }}
													onClick={closeMobileSidebar}
												/>
											}
										>
											<LayoutDashboard />
											<span>{m.nav_home()}</span>
										</SidebarMenuButton>
									</SidebarMenuItem>
								</SidebarMenu>
							</SidebarGroupContent>
						</SidebarGroup>

						<SidebarGroup>
							<SidebarGroupLabel>{m.nav_workspace()}</SidebarGroupLabel>
							<SidebarGroupContent>
								<SidebarMenu>
									<SidebarMenuItem>
										<SidebarMenuButton
											isActive={
												pathname === beepsPath ||
												pathname.startsWith(`${beepsPath}/`)
											}
											tooltip={m.nav_beeps()}
											render={
												<Link
													to="/$account_slug/beeps"
													params={{ account_slug: slug }}
													onClick={closeMobileSidebar}
												/>
											}
										>
											<Bell />
											<span>{m.nav_beeps()}</span>
										</SidebarMenuButton>
									</SidebarMenuItem>
									<SidebarMenuItem>
										<SidebarMenuButton
											isActive={
												pathname === `/${slug}/beepers` ||
												pathname.startsWith(`/${slug}/beepers/`)
											}
											tooltip={m.nav_beepers()}
											render={
												<Link
													to="/$account_slug/beepers"
													params={{ account_slug: slug }}
													onClick={closeMobileSidebar}
												/>
											}
										>
											<Activity />
											<span>{m.nav_beepers()}</span>
										</SidebarMenuButton>
									</SidebarMenuItem>
									<SidebarMenuItem>
										<SidebarMenuButton
											isActive={
												pathname === settingsPath ||
												pathname.startsWith(`${settingsPath}/`)
											}
											tooltip={m.nav_settings()}
											render={
												<Link
													to="/$account_slug/settings"
													params={{ account_slug: slug }}
													onClick={closeMobileSidebar}
												/>
											}
										>
											<Settings />
											<span>{m.nav_settings()}</span>
										</SidebarMenuButton>
									</SidebarMenuItem>
								</SidebarMenu>
							</SidebarGroupContent>
						</SidebarGroup>

						{import.meta.env.DEV ? (
							<SidebarGroup>
								<SidebarGroupLabel>{m.nav_dev()}</SidebarGroupLabel>
								<SidebarGroupContent>
									<SidebarMenu>
										<SidebarMenuItem>
											<SidebarMenuButton
												isActive={
													pathname === lettersPath ||
													pathname.startsWith(`${lettersPath}/`)
												}
												tooltip={m.nav_letters()}
												render={
													<Link
														to="/dev/letters"
														params={{ account_slug: slug }}
														onClick={closeMobileSidebar}
													/>
												}
											>
												<Mail />
												<span>{m.nav_letters()}</span>
											</SidebarMenuButton>
										</SidebarMenuItem>
									</SidebarMenu>
								</SidebarGroupContent>
							</SidebarGroup>
						) : null}

						{user?.staff ? (
							<SidebarGroup>
								<SidebarGroupLabel>{m.nav_admin()}</SidebarGroupLabel>
								<SidebarGroupContent>
									<SidebarMenu>
										<SidebarMenuItem>
											<SidebarMenuButton
												isActive={
													pathname === jobsPath ||
													pathname.startsWith(`${jobsPath}/`)
												}
												tooltip={m.nav_jobs()}
												render={
													<Link
														to="/admin/jobs"
														params={{ account_slug: slug }}
														onClick={closeMobileSidebar}
													/>
												}
											>
												<BriefcaseBusiness />
												<span>{m.nav_jobs()}</span>
											</SidebarMenuButton>
										</SidebarMenuItem>
										<SidebarMenuItem>
											<SidebarMenuButton
												isActive={
													pathname === statsPath ||
													pathname.startsWith(`${statsPath}/`)
												}
												tooltip={m.nav_stats()}
												render={
													<Link
														to="/admin/stats"
														params={{ account_slug: slug }}
														onClick={closeMobileSidebar}
													/>
												}
											>
												<Activity />
												<span>{m.nav_stats()}</span>
											</SidebarMenuButton>
										</SidebarMenuItem>
									</SidebarMenu>
								</SidebarGroupContent>
							</SidebarGroup>
						) : null}
					</>
				)}
			</SidebarContent>

			<SidebarFooter>
				{user ? <SidebarUserMenu user={user} /> : null}
			</SidebarFooter>
		</Sidebar>
	);
}
