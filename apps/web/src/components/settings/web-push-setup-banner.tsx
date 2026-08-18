import { Link, useMatchRoute, useRouterState } from "@tanstack/react-router";
import { Bell, X } from "lucide-react";
import { useEffect, useState } from "react";

import { Button, buttonVariants } from "@/components/ui/button";
import {
	Item,
	ItemActions,
	ItemContent,
	ItemDescription,
	ItemMedia,
	ItemTitle,
} from "@/components/ui/item";
import { useWebPush } from "@/hooks/use-web-push";
import { IOS_HOME_SCREEN_HINT } from "@/lib/web-push";

function dismissKey(slug: string) {
	return `beep:web-push-banner:${slug}`;
}

function accountSlugFromMatches(matches: ReadonlyArray<{ context: unknown }>) {
	for (let i = matches.length - 1; i >= 0; i--) {
		const account = (
			matches[i]?.context as { account?: { slug?: string } } | undefined
		)?.account;
		if (account?.slug) return account.slug;
	}
	return undefined;
}

export function WebPushSetupBanner() {
	const slug = useRouterState({
		select: (state) => accountSlugFromMatches(state.matches),
	});
	const matchSettings = useMatchRoute();
	const onSettings = Boolean(
		matchSettings({ to: "/$account_slug/settings", fuzzy: false }),
	);

	if (!slug || onSettings) return null;
	return <WebPushSetupBannerInner slug={slug} />;
}

function WebPushSetupBannerInner({ slug }: { slug: string }) {
	const { status, ready, pending, enable, error } = useWebPush(slug);
	const [dismissed, setDismissed] = useState(true);

	useEffect(() => {
		setDismissed(localStorage.getItem(dismissKey(slug)) === "1");
	}, [slug]);

	function dismiss() {
		localStorage.setItem(dismissKey(slug), "1");
		setDismissed(true);
	}

	if (!ready || dismissed || !status.supported || status.subscribed) {
		return null;
	}

	const needsIosInstall = status.platform === "ios" && !status.standalone;
	const denied = status.permission === "denied";

	return (
		<div className="border-b bg-muted/40 px-4 py-2.5 lg:px-6">
			<Item size="sm" className="px-0 py-0">
				<ItemMedia variant="icon">
					<Bell />
				</ItemMedia>
				<ItemContent>
					<ItemTitle>Browser notifications are off</ItemTitle>
					<ItemDescription>
						{needsIosInstall
							? IOS_HOME_SCREEN_HINT
							: denied
								? "Notifications are blocked. Allow them in the browser or system settings, then enable this device."
								: "Enable this browser so you get a system ping when a beep is due."}
					</ItemDescription>
				</ItemContent>
				<ItemActions>
					{needsIosInstall || denied ? (
						<Link
							to="/$account_slug/settings"
							params={{ account_slug: slug }}
							className={buttonVariants({ variant: "outline", size: "sm" })}
						>
							Settings
						</Link>
					) : (
						<Button
							type="button"
							size="sm"
							disabled={pending}
							onClick={() => {
								void enable();
							}}
						>
							{pending ? "Enabling…" : "Enable"}
						</Button>
					)}
					<Button
						type="button"
						variant="ghost"
						size="icon-sm"
						aria-label="Dismiss notification tip"
						onClick={dismiss}
					>
						<X />
					</Button>
				</ItemActions>
				{error ? (
					<p className="text-sm text-destructive" role="alert">
						{error}
					</p>
				) : null}
			</Item>
		</div>
	);
}
