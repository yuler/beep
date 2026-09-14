import { createFileRoute, Link, redirect } from "@tanstack/react-router";
import { Building2, Server, UserRound } from "lucide-react";
import { AuthCard, AuthLayout, AuthPending } from "@/components/layout";
import type { AccountSummary } from "@/lib/auth/account";
import { requireSession } from "@/lib/auth/guards";
import { parseDeviceSearch } from "../device";

export const Route = createFileRoute("/device/runner")({
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
				to: "/$account_slug/device/runner",
				params: { account_slug: me.accounts[0].slug },
				search: search.code ? { code: search.code } : {},
			});
		}
		return { me };
	},
	component: DeviceRunnerAccountPickerPage,
});

function DeviceRunnerAccountPickerPage() {
	const { me } = Route.useRouteContext();
	const search = Route.useSearch();
	const lastSlug = me.last_account_slug;

	return (
		<AuthLayout>
			<AuthCard description="Select which account you want to connect your self-hosted runner to.">
				{search.code ? (
					<div className="mb-4 rounded-lg border border-border bg-muted/40 p-3">
						<div className="flex items-center gap-3">
							<span className="flex size-8 shrink-0 items-center justify-center rounded-md border bg-background">
								<Server className="size-4 text-muted-foreground" />
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
				to="/$account_slug/device/runner"
				params={{ account_slug: account.slug }}
				search={code ? { code } : {}}
				className="flex items-center justify-between rounded-lg border border-border bg-card p-3 text-left transition hover:border-foreground/30 hover:bg-muted/40 focus:outline-none focus:ring-2 focus:ring-ring"
			>
				<div className="flex items-center gap-3 min-w-0">
					<span className="flex size-9 shrink-0 items-center justify-center rounded-md border bg-muted">
						<Icon className="size-4 text-muted-foreground" />
					</span>
					<div className="flex min-w-0 flex-col">
						<span className="truncate text-sm font-medium">{account.name}</span>
						<span className="truncate text-xs text-muted-foreground font-mono">
							{account.slug}
						</span>
					</div>
				</div>
				{lastUsed ? (
					<span className="shrink-0 rounded bg-muted px-2 py-0.5 text-xs text-muted-foreground">
						Last used
					</span>
				) : null}
			</Link>
		</li>
	);
}
