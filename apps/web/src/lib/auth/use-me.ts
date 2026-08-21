import { getRouteApi } from "@tanstack/react-router";

import type { MeResponse } from "@/lib/api/session";

const rootRoute = getRouteApi("__root__");

/**
 * Current user from root route context.
 */
export function useMe(): { me: MeResponse | null } {
	const { me } = rootRoute.useRouteContext();
	return { me };
}
