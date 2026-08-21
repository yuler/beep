import { getRouteApi } from "@tanstack/react-router";

import type { MeResponse } from "@/lib/api/session";

const rootRoute = getRouteApi("__root__");

/**
 * Current user from root route context.
 * No client-side resume: SSR (and every navigation) resolves `me` from the
 * session cookie, which reaches web.* in both modes (parent-domain in Mode A,
 * same-origin host cookie in Mode B).
 */
export function useMe(): { me: MeResponse | null } {
	const { me } = rootRoute.useRouteContext();
	return { me };
}
