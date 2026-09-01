import {
	useMatchRoute,
	useNavigate,
	useRouterState,
} from "@tanstack/react-router";
import { Bell, X } from "lucide-react";
import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import {
	Item,
	ItemActions,
	ItemContent,
	ItemDescription,
	ItemMedia,
	ItemTitle,
} from "@/components/ui/item";
import { useWebPush } from "@/hooks/use-web-push";
import { iosHomeScreenHint } from "@/lib/web-push";
import * as m from "@/locale/paraglide/messages";

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
	const navigate = useNavigate();

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

	function openSettings() {
		navigate({ to: "/$account_slug/settings", params: { account_slug: slug } });
	}

	function handleClick() {
		if (!needsIosInstall && !denied && !pending) void enable();
		openSettings();
	}

	return (
		<div className="border-b bg-muted/40 px-4 py-2.5 lg:px-6">
			<Item size="sm" className="px-0 py-0">
				<button
					type="button"
					className="flex min-w-0 flex-1 cursor-pointer items-center gap-2.5 rounded-lg text-left outline-none focus-visible:ring-3 focus-visible:ring-ring/50"
					onClick={handleClick}
				>
					<ItemMedia variant="icon">
						<Bell />
					</ItemMedia>
					<ItemContent>
						<ItemTitle>{m.push_banner_title()}</ItemTitle>
						<ItemDescription>
							{needsIosInstall
								? iosHomeScreenHint()
								: denied
									? m.push_banner_blocked()
									: m.push_banner_click_enable()}
						</ItemDescription>
					</ItemContent>
				</button>
				<ItemActions>
					<Button
						type="button"
						variant="ghost"
						size="icon-sm"
						aria-label={m.common_not_now()}
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
