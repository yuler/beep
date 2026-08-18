// Base URL of the Rails core API / auth server.
// Inlined from process.env.VITE_* (mise / Docker). Empty → Mode B same-origin /api.
export const CORE_URL = import.meta.env.VITE_CORE_URL;

/** Core HTML page URL — Mode A absolute origin, Mode B same-origin path. */
export function coreAppUrl(path: string): string {
	const normalized = path.startsWith("/") ? path : `/${path}`;
	return CORE_URL ? `${CORE_URL}${normalized}` : normalized;
}
