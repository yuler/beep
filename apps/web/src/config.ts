// Base URL of the Rails core API / auth server.
// Set VITE_CORE_URL in the root .env (e.g. http://core…:${CORE_PORT} or "" for Mode B).
// Empty string is Mode B (same-origin /api via Nitro proxy). Undefined → local Mode A default.
const appHost = import.meta.env.APP_HOST || "beep.localhost";
export const CORE_URL =
	import.meta.env.VITE_CORE_URL !== undefined
		? import.meta.env.VITE_CORE_URL
		: `http://core.${appHost}:3001`;

/** Core HTML page URL — Mode A absolute origin, Mode B same-origin path. */
export function coreAppUrl(path: string): string {
	const normalized = path.startsWith("/") ? path : `/${path}`;
	return CORE_URL ? `${CORE_URL}${normalized}` : normalized;
}
