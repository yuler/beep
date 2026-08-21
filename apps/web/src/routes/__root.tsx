import {
	createRootRoute,
	HeadContent,
	Outlet,
	Scripts,
} from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools";
import { type ReactNode, useEffect } from "react";
import { NotFound } from "@/components/not-found";
import { fetchMeOrNull } from "@/lib/api/session";
import { logBuildInfo } from "@/lib/build-info";

import appCss from "../styles.css?url";

const themeBootScript = `(() => {
  var stored = localStorage.getItem("theme");
  var dark =
    stored === "dark" ||
    (!stored && window.matchMedia("(prefers-color-scheme: dark)").matches);
  if (dark) {
    document.documentElement.classList.add("dark");
    document.documentElement.dataset.theme = "dark";
  } else {
    document.documentElement.dataset.theme = "light";
  }
})();`;

export const Route = createRootRoute({
	// Re-run on each navigation; fetchMe dedupes within ME_STALE_MS.
	staleTime: 0,
	beforeLoad: async () => {
		// SSR forwards the document request's cookie header server-side (see
		// serverCookieHeader in lib/api/client). Mode A cookies are
		// parent-domain (SESSION_COOKIE_DOMAIN=.${APP_HOST}), so a logged-in
		// browser's `session_id` reaches web.* and hydrates with `me` resolved.
		const me = await fetchMeOrNull();
		return { me };
	},
	head: () => ({
		meta: [
			{ charSet: "utf-8" },
			{ name: "viewport", content: "width=device-width, initial-scale=1" },
			{ title: "beep" },
		],
		links: [{ rel: "stylesheet", href: appCss }],
		scripts: [{ children: themeBootScript }],
	}),
	notFoundComponent: NotFound,
	shellComponent: RootDocument,
	component: RootComponent,
});

function RootComponent() {
	useEffect(() => {
		logBuildInfo();
	}, []);

	return <Outlet />;
}

function RootDocument({ children }: Readonly<{ children: ReactNode }>) {
	return (
		// Theme boot script / browser extensions may mutate <html>/<body> attrs before hydrate.
		<html lang="en" suppressHydrationWarning>
			<head>
				<HeadContent />
			</head>
			<body suppressHydrationWarning>
				{children}
				{import.meta.env.DEV ? (
					<TanStackRouterDevtools position="bottom-right" />
				) : null}
				<Scripts />
			</body>
		</html>
	);
}
