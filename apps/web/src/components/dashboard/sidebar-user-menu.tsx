import { Link, useNavigate, useRouter } from "@tanstack/react-router";
import { KeyRound, LogOut, ShieldAlert, Terminal, User } from "lucide-react";
import { useState } from "react";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Button } from "@/components/ui/button";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuGroup,
	DropdownMenuItem,
	DropdownMenuLabel,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { ApiError } from "@/lib/api/client";
import type { MeResponse } from "@/lib/api/session";
import { destroySession } from "@/lib/api/session";
import { m } from "@/locale/paraglide/messages";

export function SidebarUserMenu({
	user,
}: {
	user: MeResponse["identity"];
	currentSlug?: string;
}) {
	const navigate = useNavigate();
	const router = useRouter();
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
				err instanceof ApiError ? err.message : m.auth_sign_out_failed(),
			);
		} finally {
			setSigningOut(false);
		}
	}

	return (
		<DropdownMenu
			open={open}
			onOpenChange={(next) => {
				setOpen(next);
				if (!next) setSignOutError(null);
			}}
		>
			<DropdownMenuTrigger
				render={
					<Button
						variant="ghost"
						size="icon-sm"
						className="size-8 p-0 min-h-0 sm:min-h-0 border-0"
						aria-label={user.name}
					/>
				}
			>
				<Avatar size="sm">
					<AvatarFallback>{initials}</AvatarFallback>
				</Avatar>
			</DropdownMenuTrigger>

			<DropdownMenuContent className="w-56" align="end" sideOffset={4}>
				<DropdownMenuGroup>
					<DropdownMenuLabel>
						<div className="flex flex-col gap-0.5">
							<span className="truncate font-medium text-ink">{user.name}</span>
							<span className="truncate type-meta text-gray">{user.email}</span>
						</div>
					</DropdownMenuLabel>
				</DropdownMenuGroup>

				<DropdownMenuSeparator />

				<DropdownMenuGroup>
					<DropdownMenuItem
						render={<Link to="/my/settings" />}
						onClick={() => setOpen(false)}
					>
						<User className="size-3.5 text-gray shrink-0" />
						{m.my_profile_settings()}
					</DropdownMenuItem>
					<DropdownMenuItem
						render={<Link to="/my/access_tokens" />}
						onClick={() => setOpen(false)}
					>
						<KeyRound className="size-3.5 text-gray shrink-0" />
						{m.nav_api_tokens()}
					</DropdownMenuItem>
				</DropdownMenuGroup>

				{user.staff ? (
					<>
						<DropdownMenuSeparator />
						<DropdownMenuGroup>
							<DropdownMenuLabel>{m.nav_admin()}</DropdownMenuLabel>
							<DropdownMenuItem
								render={<Link to="/admin/jobs" />}
								onClick={() => setOpen(false)}
							>
								<ShieldAlert className="size-3.5 text-gray shrink-0" />
								{m.nav_jobs()}
							</DropdownMenuItem>
							<DropdownMenuItem
								render={<Link to="/admin/stats" />}
								onClick={() => setOpen(false)}
							>
								<ShieldAlert className="size-3.5 text-gray shrink-0" />
								{m.nav_stats()}
							</DropdownMenuItem>
						</DropdownMenuGroup>
					</>
				) : null}

				{import.meta.env.DEV ? (
					<>
						<DropdownMenuSeparator />
						<DropdownMenuGroup>
							<DropdownMenuLabel>{m.nav_dev()}</DropdownMenuLabel>
							<DropdownMenuItem
								render={<Link to="/dev/letters" />}
								onClick={() => setOpen(false)}
							>
								<Terminal className="size-3.5 text-gray shrink-0" />
								{m.nav_letters()}
							</DropdownMenuItem>
						</DropdownMenuGroup>
					</>
				) : null}

				<DropdownMenuSeparator />

				<DropdownMenuGroup>
					<DropdownMenuItem
						variant="destructive"
						disabled={signingOut}
						onClick={() => {
							void handleSignOut();
						}}
					>
						<LogOut className="size-3.5 shrink-0" />
						{signingOut ? m.auth_signing_out() : m.auth_logout()}
					</DropdownMenuItem>
					{signOutError ? (
						<p
							className="px-2 py-1.5 text-xs text-ink border-l border-ink"
							role="alert"
						>
							{signOutError}
						</p>
					) : null}
				</DropdownMenuGroup>
			</DropdownMenuContent>
		</DropdownMenu>
	);
}

export const ConsoleUserMenu = SidebarUserMenu;
