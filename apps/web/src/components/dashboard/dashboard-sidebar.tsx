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
import { useTranslation } from "@/lib/i18n";

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
	const { t } = useTranslation();
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
								tooltip={t("nav.back_to_workspace")}
								className="group-data-[collapsible=icon]:size-8! group-data-[collapsible=icon]:p-0!"
							>
								<span className="flex aspect-square size-8 shrink-0 items-center justify-center rounded-lg bg-foreground text-background">
									<LogoMark className="size-4" />
								</span>
								<div className="grid flex-1 text-left text-sm leading-tight">
									<span className="truncate font-semibold">
										{t("term.beep_capitalized")}
									</span>
									<span className="truncate text-xs text-muted-foreground">
										{t("nav.personal_settings")}
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
						<SidebarGroupLabel>{t("nav.personal_settings")}</SidebarGroupLabel>
						<SidebarGroupContent>
							<SidebarMenu>
								<SidebarMenuItem>
									<SidebarMenuButton
										isActive={
											pathname === mySettingsPath ||
											pathname.startsWith(`${mySettingsPath}/`)
										}
										tooltip={t("nav.profile_preferences")}
										render={
											<Link to="/my/settings" onClick={closeMobileSidebar} />
										}
									>
										<User />
										<span>{t("nav.settings")}</span>
									</SidebarMenuButton>
								</SidebarMenuItem>
								<SidebarMenuItem>
									<SidebarMenuButton
										isActive={
											pathname === accessTokensPath ||
											pathname.startsWith(`${accessTokensPath}/`)
										}
										tooltip={t("nav.api_access_tokens")}
										render={
											<Link
												to="/my/access_tokens"
												onClick={closeMobileSidebar}
											/>
										}
									>
										<KeyRound />
										<span>{t("nav.api_tokens")}</span>
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
											tooltip={t("nav.home")}
											render={
												<Link
													to="/$account_slug"
													params={{ account_slug: slug }}
													onClick={closeMobileSidebar}
												/>
											}
										>
											<LayoutDashboard />
											<span>{t("nav.home")}</span>
										</SidebarMenuButton>
									</SidebarMenuItem>
								</SidebarMenu>
							</SidebarGroupContent>
						</SidebarGroup>

						<SidebarGroup>
							<SidebarGroupLabel>{t("nav.workspace")}</SidebarGroupLabel>
							<SidebarGroupContent>
								<SidebarMenu>
									<SidebarMenuItem>
										<SidebarMenuButton
											isActive={
												pathname === beepsPath ||
												pathname.startsWith(`${beepsPath}/`)
											}
											tooltip={t("nav.beeps")}
											render={
												<Link
													to="/$account_slug/beeps"
													params={{ account_slug: slug }}
													onClick={closeMobileSidebar}
												/>
											}
										>
											<Bell />
											<span>{t("nav.beeps")}</span>
										</SidebarMenuButton>
									</SidebarMenuItem>
									<SidebarMenuItem>
										<SidebarMenuButton
											isActive={
												pathname === `/${slug}/beepers` ||
												pathname.startsWith(`/${slug}/beepers/`)
											}
											tooltip={t("nav.beepers")}
											render={
												<Link
													to="/$account_slug/beepers"
													params={{ account_slug: slug }}
													onClick={closeMobileSidebar}
												/>
											}
										>
											<Activity />
											<span>{t("nav.beepers")}</span>
										</SidebarMenuButton>
									</SidebarMenuItem>
									<SidebarMenuItem>
										<SidebarMenuButton
											isActive={
												pathname === settingsPath ||
												pathname.startsWith(`${settingsPath}/`)
											}
											tooltip={t("nav.settings")}
											render={
												<Link
													to="/$account_slug/settings"
													params={{ account_slug: slug }}
													onClick={closeMobileSidebar}
												/>
											}
										>
											<Settings />
											<span>{t("nav.settings")}</span>
										</SidebarMenuButton>
									</SidebarMenuItem>
								</SidebarMenu>
							</SidebarGroupContent>
						</SidebarGroup>

						{import.meta.env.DEV ? (
							<SidebarGroup>
								<SidebarGroupLabel>{t("nav.dev")}</SidebarGroupLabel>
								<SidebarGroupContent>
									<SidebarMenu>
										<SidebarMenuItem>
											<SidebarMenuButton
												isActive={
													pathname === lettersPath ||
													pathname.startsWith(`${lettersPath}/`)
												}
												tooltip={t("nav.letters")}
												render={
													<Link
														to="/dev/letters"
														params={{ account_slug: slug }}
														onClick={closeMobileSidebar}
													/>
												}
											>
												<Mail />
												<span>{t("nav.letters")}</span>
											</SidebarMenuButton>
										</SidebarMenuItem>
									</SidebarMenu>
								</SidebarGroupContent>
							</SidebarGroup>
						) : null}

						{user?.staff ? (
							<SidebarGroup>
								<SidebarGroupLabel>{t("nav.admin")}</SidebarGroupLabel>
								<SidebarGroupContent>
									<SidebarMenu>
										<SidebarMenuItem>
											<SidebarMenuButton
												isActive={
													pathname === jobsPath ||
													pathname.startsWith(`${jobsPath}/`)
												}
												tooltip={t("nav.jobs")}
												render={
													<Link
														to="/admin/jobs"
														params={{ account_slug: slug }}
														onClick={closeMobileSidebar}
													/>
												}
											>
												<BriefcaseBusiness />
												<span>{t("nav.jobs")}</span>
											</SidebarMenuButton>
										</SidebarMenuItem>
										<SidebarMenuItem>
											<SidebarMenuButton
												isActive={
													pathname === statsPath ||
													pathname.startsWith(`${statsPath}/`)
												}
												tooltip={t("nav.stats")}
												render={
													<Link
														to="/admin/stats"
														params={{ account_slug: slug }}
														onClick={closeMobileSidebar}
													/>
												}
											>
												<Activity />
												<span>{t("nav.stats")}</span>
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
