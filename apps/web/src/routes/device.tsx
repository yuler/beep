import { createFileRoute, Link, redirect } from "@tanstack/react-router";
import { Building2, Check, Laptop, UserRound } from "lucide-react";
import { AuthCard, AuthLayout, AuthPending } from "@/components/layout";
import { buttonVariants } from "@/components/ui/button";
import type { AccountSummary } from "@/lib/auth/account";
import { requireSession } from "@/lib/auth/guards";
import { cn } from "@/lib/utils";
import { m } from "@/locale/paraglide/messages";

export type DeviceSearch = {
	code?: string;
};

export function parseDeviceSearch(
	search: Record<string, unknown>,
): DeviceSearch {
	const code =
		typeof search.code === "string" && search.code.trim() !== ""
			? search.code.trim()
			: undefined;
	return code ? { code } : {};
}

export const Route = createFileRoute("/device")({
	ssr: false,
	pendingComponent: AuthPending,
	validateSearch: parseDeviceSearch,
	beforeLoad: async ({ context, location, search }) => {
		const me = await requireSession({ context, location });
		if (me.accounts.length === 0) {
			throw redirect({ to: "/sign" });
		}
		if (me.accounts.length === 1) {
			throw redirect({
				to: "/$account_slug/device",
				params: { account_slug: me.accounts[0].slug },
				search: search.code ? { code: search.code } : {},
			});
		}
		return { me };
	},
	component: DeviceAccountPickerPage,
});

function DeviceAccountPickerPage() {
	const { me } = Route.useRouteContext();
	const search = Route.useSearch();
	const lastSlug = me.last_account_slug;

	return (
		<AuthLayout>
			<AuthCard description="Select which account you want to connect your CLI channel to.">
				{search.code ? (
					<div className="mb-4 rounded-lg border border-border bg-muted/40 p-3">
						<div className="flex items-center gap-3">
							<span className="flex size-8 shrink-0 items-center justify-center rounded-md border bg-background">
								<Laptop className="size-4 text-muted-foreground" />
							</span>
							<div className="flex min-w-0 flex-col">
								<span className="text-xs text-muted-foreground">
									Verification Code
								</span>
								<span className="font-mono text-sm font-bold tracking-wider">
									{search.code}
								</span>
							</div>
						</div>
					</div>
				) : null}

				<ul className="flex flex-col gap-2">
					{me.accounts.map((account) => (
						<AccountChoice
							key={account.id}
							account={account}
							lastUsed={account.slug === lastSlug}
							code={search.code}
						/>
					))}
				</ul>
			</AuthCard>
		</AuthLayout>
	);
}

function AccountChoice({
	account,
	lastUsed,
	code,
}: {
	account: AccountSummary;
	lastUsed: boolean;
	code?: string;
}) {
	const Icon = account.personal ? UserRound : Building2;

	return (
		<li>
			<Link
				to="/$account_slug/device"
				params={{ account_slug: account.slug }}
				search={code ? { code } : {}}
				className={cn(
					buttonVariants({ variant: "outline" }),
					"h-auto w-full justify-start gap-3 px-3 py-3",
					lastUsed && "border-primary/40 bg-primary/5",
				)}
			>
				<span className="flex size-8 shrink-0 items-center justify-center rounded-lg border border-border">
					<Icon className="size-4" />
				</span>
				<span className="flex min-w-0 flex-1 flex-col items-start text-left">
					<span className="truncate font-medium">{account.name}</span>
					<span className="truncate text-xs font-normal text-muted-foreground">
						{account.personal
							? m.account_type_personal()
							: m.account_type_team()}{" "}
						· /{account.slug}
						{lastUsed ? m.auth_last_used_suffix() : ""}
					</span>
				</span>
				{lastUsed ? <Check className="size-4 shrink-0 text-primary" /> : null}
			</Link>
		</li>
	);
}
