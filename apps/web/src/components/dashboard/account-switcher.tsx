import { useNavigate, useRouterState } from "@tanstack/react-router";
import { Building2, Check, ChevronDown, UserRound, Users } from "lucide-react";
import { useEffect, useState } from "react";
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
import { rememberLastAccount } from "@/lib/api/session";
import type { AccountSummary } from "@/lib/auth/account";
import { cn } from "@/lib/utils";
import { m } from "@/locale/paraglide/messages";

export function AccountSwitcher({
	accounts,
	slug,
	className,
}: {
	accounts: AccountSummary[];
	slug: string;
	className?: string;
}) {
	const navigate = useNavigate();
	const pathname = useRouterState({ select: (s) => s.location.pathname });
	const [active, setActive] = useState<AccountSummary | null>(
		() => accounts.find((account) => account.slug === slug) ?? null,
	);

	useEffect(() => {
		setActive(accounts.find((account) => account.slug === slug) ?? null);
	}, [accounts, slug]);

	if (!active) {
		return null;
	}

	function selectAccount(account: AccountSummary) {
		setActive(account);
		void rememberLastAccount(account.slug).catch(() => {
			// Picker hint is best-effort.
		});

		if (pathname.startsWith("/admin") || pathname.startsWith("/dev")) {
			return;
		}

		void navigate({
			to: "/$account_slug",
			params: { account_slug: account.slug },
		});
	}

	const ActiveIcon = active.personal ? UserRound : Building2;

	return (
		<DropdownMenu>
			<DropdownMenuTrigger
				render={
					<Button
						variant="ghost"
						size="sm"
						className={cn(
							"h-8 gap-1.5 px-2 text-xs font-normal text-ink hover:border-ink min-h-0 sm:min-h-0",
							className,
						)}
					/>
				}
			>
				<ActiveIcon className="size-3.5 text-gray shrink-0" />
				<span className="max-w-[120px] sm:max-w-[180px] truncate">
					{active.name}
				</span>
				<ChevronDown className="size-3 text-gray shrink-0" />
			</DropdownMenuTrigger>

			<DropdownMenuContent className="w-56" align="start" sideOffset={4}>
				<DropdownMenuGroup>
					<DropdownMenuLabel>{m.account_accounts()}</DropdownMenuLabel>
					{accounts.map((account) => {
						const Icon = account.personal ? UserRound : Building2;
						const selected = account.id === active.id;
						return (
							<DropdownMenuItem
								key={account.id}
								className="gap-2.5 py-1.5"
								onClick={() => selectAccount(account)}
							>
								<Icon className="size-3.5 text-gray shrink-0" />
								<span className="flex-1 truncate">{account.name}</span>
								<Check
									className={cn(
										"size-3.5 text-ink",
										selected ? "opacity-100" : "opacity-0",
									)}
								/>
							</DropdownMenuItem>
						);
					})}
				</DropdownMenuGroup>
				<DropdownMenuSeparator />
				<DropdownMenuGroup>
					<DropdownMenuItem
						className="gap-2.5 py-1.5"
						onClick={() => {
							void navigate({ to: "/accounts" });
						}}
					>
						<Users className="size-3.5 text-gray shrink-0" />
						<span className="flex-1 truncate">
							{m.account_manage_accounts()}
						</span>
					</DropdownMenuItem>
				</DropdownMenuGroup>
			</DropdownMenuContent>
		</DropdownMenu>
	);
}
