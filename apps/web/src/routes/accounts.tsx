import { createFileRoute, Link, redirect } from "@tanstack/react-router";
import { Building2, Check, UserRound } from "lucide-react";
import { AuthCard, AuthLayout, AuthPending } from "@/components/layout";
import { buttonVariants } from "@/components/ui/button";
import type { AccountSummary } from "@/lib/auth/account";
import { requireSession } from "@/lib/auth/guards";
import { cn } from "@/lib/utils";

export const Route = createFileRoute("/accounts")({
	pendingComponent: AuthPending,
	beforeLoad: async ({ context, location }) => {
		const me = await requireSession({ context, location });
		if (me.accounts.length === 0) {
			throw redirect({ to: "/sign" });
		}
		if (me.accounts.length === 1) {
			throw redirect({
				to: "/$account_slug",
				params: { account_slug: me.accounts[0].slug },
			});
		}
		return { me };
	},
	component: AccountsPage,
});

function AccountsPage() {
	const { me } = Route.useRouteContext();
	const lastSlug = me.last_account_slug;
	const description = lastSlug
		? `Last used: ${lastSlug} — pick an account to continue.`
		: "Pick an account to continue.";

	return (
		<AuthLayout>
			<AuthCard description={description}>
				<ul className="flex flex-col gap-2">
					{me.accounts.map((account) => (
						<AccountChoice
							key={account.id}
							account={account}
							lastUsed={account.slug === lastSlug}
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
}: {
	account: AccountSummary;
	lastUsed: boolean;
}) {
	const Icon = account.personal ? UserRound : Building2;

	return (
		<li>
			<Link
				to="/$account_slug"
				params={{ account_slug: account.slug }}
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
						{account.personal ? "Personal" : "Team"} · /{account.slug}
						{lastUsed ? " · last used" : ""}
					</span>
				</span>
				{lastUsed ? <Check className="size-4 shrink-0 text-primary" /> : null}
			</Link>
		</li>
	);
}
