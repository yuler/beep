// Base URL of the Rails core API / auth server.
// Inlined from process.env.VITE_* (mise / Docker). Empty → Mode B same-origin /api.
export const CORE_URL = import.meta.env.VITE_CORE_URL;

/** Origin that external clients (cron, curl) should call. Mode A: core host; Mode B: this site. */
export function publicApiOrigin(): string {
	if (CORE_URL) return CORE_URL.replace(/\/$/, "");
	if (typeof window !== "undefined") return window.location.origin;
	return "";
}

/** Absolute URL to a Rails HTML page. Empty in Mode B — core is not public. */
export function coreAppUrl(path: string): string {
	if (!CORE_URL) return "";
	const normalized = path.startsWith("/") ? path : `/${path}`;
	return `${CORE_URL}${normalized}`;
}
