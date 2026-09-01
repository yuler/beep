import {
	createRootRoute,
	HeadContent,
	Outlet,
	Scripts,
} from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools";
import { type ReactNode, useEffect } from "react";
import { NotFound } from "@/components/not-found";
import { VersionUpdateDialog } from "@/components/version-update-dialog";
import { fetchMeOrNull } from "@/lib/api/session";
import { logBuildInfo } from "@/lib/build-info";
import { getLocale, localeConfig, locales, m } from "@/lib/locale";

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
  var link = document.getElementById("favicon-theme");
  if (link) link.href = dark ? "/icon-dark.svg" : "/icon-light.svg";
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
	head: () => {
		const currentLocale = getLocale();
		const ogLocale = localeConfig[currentLocale].hreflang.replace("-", "_");
		const alternateOgLocales = locales
			.filter((l) => l !== currentLocale)
			.map((l) => localeConfig[l].hreflang.replace("-", "_"));

		return {
			meta: [
				{ charSet: "utf-8" },
				{ name: "viewport", content: "width=device-width, initial-scale=1" },
				{
					title: m.term_beep(),
				},
				{ property: "og:locale", content: ogLocale },
				...alternateOgLocales.map((loc) => ({
					property: "og:locale:alternate",
					content: loc,
				})),
			],
			links: [
				{ rel: "stylesheet", href: appCss },
				{ rel: "icon", href: "/favicon.ico", sizes: "any" },
				{
					id: "favicon-theme",
					rel: "icon",
					href: "/icon-light.svg",
					type: "image/svg+xml",
				},
				{
					rel: "icon",
					href: "/favicon-32.png",
					type: "image/png",
					sizes: "32x32",
				},
				{
					rel: "apple-touch-icon",
					href: "/apple-touch-icon.png",
					sizes: "180x180",
				},
			],
			scripts: [{ children: themeBootScript }],
		};
	},
	notFoundComponent: NotFound,
	shellComponent: RootDocument,
	component: RootComponent,
});

function RootComponent() {
	useEffect(() => {
		logBuildInfo();
	}, []);

	return (
		<>
			<Outlet />
			<VersionUpdateDialog />
		</>
	);
}

function RootDocument({ children }: Readonly<{ children: ReactNode }>) {
	return (
		// Theme boot script / browser extensions may mutate <html>/<body> attrs before hydrate.
		<html lang={localeConfig[getLocale()].hreflang} suppressHydrationWarning>
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
