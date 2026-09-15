import { createFileRoute, redirect } from "@tanstack/react-router";

export const Route = createFileRoute("/$account_slug/settings/notifications")({
	beforeLoad: ({ params }) => {
		throw redirect({
			to: "/$account_slug/settings/channels",
			params: { account_slug: params.account_slug },
		});
	},
});
