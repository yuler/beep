import { createFileRoute, redirect } from "@tanstack/react-router";

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
	validateSearch: parseDeviceSearch,
	beforeLoad: async ({ search }) => {
		throw redirect({
			to: "/device/channel",
			search: search.code ? { code: search.code } : {},
		});
	},
});
