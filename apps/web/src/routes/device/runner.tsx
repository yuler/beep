import { createFileRoute, redirect } from "@tanstack/react-router";
import { resolveShellAccount } from "@/lib/auth/account";
import { requireSession } from "@/lib/auth/guards";
import { parseDeviceSearch } from "@/lib/device-search";

export const Route = createFileRoute("/device/runner")({
	ssr: false,
	validateSearch: parseDeviceSearch,
	beforeLoad: async ({ context, location }) => {
		const me = await requireSession({ context, location });
		const account = resolveShellAccount(me);
		if (!account) {
			throw redirect({ to: "/sign" });
		}
		throw redirect({
			to: "/$account_slug/device/runner",
			params: { account_slug: account.slug },
			search: location.search,
		});
	},
});
