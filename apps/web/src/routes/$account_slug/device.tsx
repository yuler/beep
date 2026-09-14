import { createFileRoute, redirect } from "@tanstack/react-router";
import { parseDeviceSearch } from "../device";

export const Route = createFileRoute("/$account_slug/device")({
	ssr: false,
	validateSearch: parseDeviceSearch,
	beforeLoad: async ({ params, search }) => {
		throw redirect({
			to: "/$account_slug/device/channel",
			params: { account_slug: params.account_slug },
			search: search.code ? { code: search.code } : {},
		});
	},
});
