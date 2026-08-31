import { createRouter } from "@tanstack/react-router";

import { NotFound } from "@/components/not-found";
import { ME_STALE_MS } from "@/lib/api/session";
import { deLocalizeUrl, localizeUrl } from "@/locale/middleware";

import { routeTree } from "./routeTree.gen";

export function getRouter() {
	const router = createRouter({
		routeTree,
		rewrite: {
			input: ({ url }) => deLocalizeUrl(url),
			output: ({ url }) => localizeUrl(url),
		},
		defaultPreload: "intent",
		// Reuse preloaded beforeLoad/loader data briefly so hover→click is not a second fetch.
		defaultPreloadStaleTime: ME_STALE_MS,
		defaultStaleTime: ME_STALE_MS,
		defaultNotFoundComponent: NotFound,
		scrollRestoration: true,
	});

	return router;
}

declare module "@tanstack/react-router" {
	interface Register {
		router: ReturnType<typeof getRouter>;
	}
}
