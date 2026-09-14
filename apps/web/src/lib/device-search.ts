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
