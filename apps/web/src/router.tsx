import { createRouter } from "@tanstack/react-router";

import { NotFound } from "@/components/not-found";
import { ME_STALE_MS } from "@/lib/api/session";
import { deLocalizeUrl, localizeUrl } from "@/locale/paraglide/runtime";

import { routeTree } from "./routeTree.gen";

export function getRouter() {
	let currentRequestLocale: string | undefined;

	const router = createRouter({
		routeTree,
		rewrite: {
			input: ({ url }) => {
				const segments = url.pathname.split("/").filter(Boolean);
				const first = segments[0]?.toLowerCase();
				if (first === "zh" || first === "zh-cn" || first === "zh_cn") {
					currentRequestLocale = "zh";
					const newUrl = new URL(url.href);
					segments.shift();
					newUrl.pathname = `/${segments.join("/")}`;
					return newUrl;
				}
				currentRequestLocale = undefined;
				return url;
			},
			output: ({ url }) => {
				const isTargetZh =
					currentRequestLocale === "zh" ||
					(typeof window !== "undefined" &&
						window.location.pathname.toLowerCase().startsWith("/zh"));

				if (isTargetZh) {
					const segments = url.pathname.split("/").filter(Boolean);
					const first = segments[0]?.toLowerCase();
					if (first === "zh" || first === "zh-cn" || first === "zh_cn") {
						segments.shift();
					}
					const newUrl = new URL(url.href);
					newUrl.pathname = `/zh${segments.length > 0 ? `/${segments.join("/")}` : ""}`;
					return newUrl;
				}
				return url;
			},
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
