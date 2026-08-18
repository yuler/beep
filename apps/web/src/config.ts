// Base URL of the Rails core API / auth server.
// Inlined from process.env.VITE_* (mise / Docker). Empty → Mode B same-origin /api.
export const CORE_URL = import.meta.env.VITE_CORE_URL;

/** Absolute URL to a Rails HTML page. Empty in Mode B — core is not public. */
export function coreAppUrl(path: string): string {
	if (!CORE_URL) return "";
	const normalized = path.startsWith("/") ? path : `/${path}`;
	return `${CORE_URL}${normalized}`;
}
