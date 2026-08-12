/** Relative same-origin paths only — mirrors Core `SessionsController#safe_return_to`. */
export function safeReturnTo(value: unknown): string | undefined {
	if (typeof value !== "string" || value.trim() === "") return undefined;

	try {
		const uri = new URL(value, "http://example.invalid");
		if (uri.host !== "example.invalid") return undefined;
		if (!value.startsWith("/")) return undefined;
		if (value.startsWith("//")) return undefined;
		return value;
	} catch {
		return undefined;
	}
}

export type SignSearch = {
	return_to?: string;
};

export function parseSignSearch(search: Record<string, unknown>): SignSearch {
	const return_to = safeReturnTo(search.return_to);
	return return_to ? { return_to } : {};
}
