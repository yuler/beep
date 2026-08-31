import { getDictionary, translate } from "@beep/locales";
import {
	createRootRoute,
	HeadContent,
	Outlet,
	Scripts,
	useLocation,
} from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools";
import { type ReactNode, useEffect, useMemo } from "react";
import { NotFound } from "@/components/not-found";
import { VersionUpdateDialog } from "@/components/version-update-dialog";
import { fetchMeOrNull } from "@/lib/api/session";
import { logBuildInfo } from "@/lib/build-info";
import {
	DEFAULT_LOCALE,
	I18nContext,
	isSupportedLocale,
	type Locale,
	type TranslationKey,
} from "@/lib/i18n";

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

function getLocaleFromPath(pathname: string): Locale {
	const segments = pathname.split("/").filter(Boolean);
	const first = segments[0];
	if (isSupportedLocale(first)) {
		return first;
	}
	return DEFAULT_LOCALE;
}

function buildSetLocaleUrl(
	pathname: string,
	searchParams: Record<string, unknown>,
	newLocale: Locale,
): string {
	const segments = pathname.split("/").filter(Boolean);
	const first = segments[0];
	if (isSupportedLocale(first)) {
		segments.shift();
	}
	const cleanPath = `/${segments.join("/")}`;
	const prefix = newLocale === DEFAULT_LOCALE ? "" : `/${newLocale}`;
	const path =
		`${prefix}${cleanPath === "/" && prefix !== "" ? "" : cleanPath}` || "/";
	const searchEntries = Object.entries(searchParams).filter(
		([_, v]) => v !== undefined && v !== null,
	);
	if (searchEntries.length === 0) {
		return path;
	}
	const searchStr = new URLSearchParams(
		searchEntries.map(([k, v]) => [k, String(v)]),
	).toString();
	return searchStr ? `${path}?${searchStr}` : path;
}

function RootComponent() {
	const location = useLocation();

	useEffect(() => {
		logBuildInfo();
	}, []);

	const currentLocale = useMemo(() => {
		return getLocaleFromPath(location.pathname);
	}, [location.pathname]);

	const dict = useMemo(() => {
		return getDictionary(currentLocale);
	}, [currentLocale]);

	const i18nValue = useMemo(() => {
		return {
			locale: currentLocale,
			dict,
			t: (key: TranslationKey, params?: Record<string, string | number>) =>
				translate(dict, key, params),
			setLocaleUrl: (newLocale: Locale) =>
				buildSetLocaleUrl(
					location.pathname,
					location.search as Record<string, unknown>,
					newLocale,
				),
		};
	}, [currentLocale, dict, location.pathname, location.search]);

	return (
		<I18nContext.Provider value={i18nValue}>
			<Outlet />
			<VersionUpdateDialog />
		</I18nContext.Provider>
	);
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
