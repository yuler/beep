import { getRouteApi, useRouter } from "@tanstack/react-router";
import { useEffect, useState } from "react";

import { fetchMeOrNull, type MeResponse } from "@/lib/api/session";

const rootRoute = getRouteApi("__root__");

/**
 * Current user from root route context. Mode B resolves the identity server-side
 * (cookie on the web origin) and re-resolves on client navigation.
 */
export function useMe(): { me: MeResponse | null; isLoading: boolean } {
	const router = useRouter();
	const { me: contextMe } = rootRoute.useRouteContext();
	const [clientMe, setClientMe] = useState<MeResponse | null | undefined>(
		undefined,
	);

	useEffect(() => {
		if (contextMe) {
			setClientMe(undefined);
			return;
		}

		let cancelled = false;
		void fetchMeOrNull().then((fresh) => {
			if (cancelled) return;
			setClientMe(fresh);
			if (fresh) void router.invalidate();
		});

		return () => {
			cancelled = true;
		};
	}, [contextMe, router]);

	if (contextMe) return { me: contextMe, isLoading: false };
	if (clientMe !== undefined) return { me: clientMe, isLoading: false };
	return { me: null, isLoading: true };
}
