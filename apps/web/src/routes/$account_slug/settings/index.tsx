import { createFileRoute, redirect } from "@tanstack/react-router";

export const Route = createFileRoute("/$account_slug/settings/")({
	beforeLoad: ({ params }) => {
		throw redirect({
			to: "/$account_slug/settings/general",
			params: { account_slug: params.account_slug },
		});
	},
});
