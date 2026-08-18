// Base URL of the Rails core API / auth server.
// Only VITE_* is inlined into the browser bundle. Set VITE_CORE_URL in `.env`
// (copy from `.env.example`). Empty string is Mode B (same-origin /api).
const viteCoreUrl = import.meta.env.VITE_CORE_URL;
if (viteCoreUrl === undefined) {
	throw new Error("VITE_CORE_URL is required. Copy .env.example to .env.");
}
export const CORE_URL = viteCoreUrl;

/** Core HTML page URL — Mode A absolute origin, Mode B same-origin path. */
export function coreAppUrl(path: string): string {
	const normalized = path.startsWith("/") ? path : `/${path}`;
	return CORE_URL ? `${CORE_URL}${normalized}` : normalized;
}
