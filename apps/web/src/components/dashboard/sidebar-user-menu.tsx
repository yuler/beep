import { Link, useNavigate, useRouter } from "@tanstack/react-router";
import { ChevronsUpDown, LogOut, Settings } from "lucide-react";
import { useState } from "react";

import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuGroup,
	DropdownMenuItem,
	DropdownMenuLabel,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
	SidebarMenu,
	SidebarMenuButton,
	SidebarMenuItem,
	useSidebar,
} from "@/components/ui/sidebar";
import { ApiError } from "@/lib/api/client";
import type { MeResponse } from "@/lib/api/session";
import { destroySession } from "@/lib/api/session";

export function SidebarUserMenu({
	user,
	slug,
}: {
	user: MeResponse["identity"];
	slug: string;
}) {
	const navigate = useNavigate();
	const router = useRouter();
	const { isMobile } = useSidebar();
	const [open, setOpen] = useState(false);
	const [signingOut, setSigningOut] = useState(false);
	const [signOutError, setSignOutError] = useState<string | null>(null);

	const initials = user.name.trim().charAt(0).toUpperCase() || "U";

	async function handleSignOut() {
		setSigningOut(true);
		setSignOutError(null);
		try {
			await destroySession();
			setOpen(false);
			await router.invalidate();
			await navigate({ to: "/" });
		} catch (err) {
			setSignOutError(
				err instanceof ApiError
					? err.message
					: "Could not sign out. Please try again.",
			);
		} finally {
			setSigningOut(false);
		}
	}

	return (
		<SidebarMenu>
			<SidebarMenuItem>
				<DropdownMenu
					open={open}
					onOpenChange={(next) => {
						setOpen(next);
						if (!next) setSignOutError(null);
					}}
				>
					<DropdownMenuTrigger
						render={
							<SidebarMenuButton
								size="lg"
								className="data-[popup-open]:bg-sidebar-accent data-[popup-open]:text-sidebar-accent-foreground"
							/>
						}
					>
						<Avatar size="sm" className="rounded-lg">
							<AvatarFallback className="rounded-lg">{initials}</AvatarFallback>
						</Avatar>
						<div className="grid flex-1 text-left text-sm leading-tight">
							<span className="truncate font-semibold">{user.name}</span>
							<span className="truncate text-xs text-muted-foreground">
								{user.email}
							</span>
						</div>
						<ChevronsUpDown className="ml-auto size-4" />
					</DropdownMenuTrigger>

					<DropdownMenuContent
						className="min-w-56 rounded-lg"
						side={isMobile ? "bottom" : "right"}
						align="end"
						sideOffset={4}
					>
						<DropdownMenuGroup>
							<DropdownMenuLabel className="p-0 font-normal">
								<div className="flex items-center gap-2 px-1 py-1.5 text-left text-sm">
									<Avatar size="sm" className="rounded-lg">
										<AvatarFallback className="rounded-lg">
											{initials}
										</AvatarFallback>
									</Avatar>
									<div className="grid flex-1 text-left text-sm leading-tight">
										<span className="truncate font-semibold">{user.name}</span>
										<span className="truncate text-xs text-muted-foreground">
											{user.email}
										</span>
									</div>
								</div>
							</DropdownMenuLabel>
						</DropdownMenuGroup>

						<DropdownMenuSeparator />

						<DropdownMenuGroup>
							<DropdownMenuItem
								render={
									<Link
										to="/$account_slug/settings"
										params={{ account_slug: slug }}
									/>
								}
								onClick={() => setOpen(false)}
							>
								<Settings />
								Settings
							</DropdownMenuItem>
						</DropdownMenuGroup>

						<DropdownMenuSeparator />

						<DropdownMenuGroup>
							<DropdownMenuItem
								variant="destructive"
								disabled={signingOut}
								onClick={() => {
									void handleSignOut();
								}}
							>
								<LogOut />
								{signingOut ? "Signing out…" : "Log out"}
							</DropdownMenuItem>
							{signOutError ? (
								<p
									className="px-2 py-1.5 text-xs text-destructive"
									role="alert"
								>
									{signOutError}
								</p>
							) : null}
						</DropdownMenuGroup>
					</DropdownMenuContent>
				</DropdownMenu>
			</SidebarMenuItem>
		</SidebarMenu>
	);
}
