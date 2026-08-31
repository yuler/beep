import { getDictionary, translate } from "@beep/locales";
import {
	createRootRoute,
	HeadContent,
	Outlet,
	Scripts,
	useLocation,
} from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools";
import { type ReactNode, useEffect, useMemo, useState } from "react";
import { NotFound } from "@/components/not-found";
import { VersionUpdateDialog } from "@/components/version-update-dialog";
import { fetchMeOrNull } from "@/lib/api/session";
import { logBuildInfo } from "@/lib/build-info";
import {
	buildLocalizedUrl,
	DEFAULT_LOCALE,
	extractLocaleFromPath,
	getStoredLocale,
	I18nContext,
	type Locale,
	saveLocalePreference,
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

function RootComponent() {
	const location = useLocation();
	const pathLocale = extractLocaleFromPath(location.pathname).locale;

	const [locale, setLocaleState] = useState<Locale>(
		() => pathLocale ?? getStoredLocale(),
	);

	useEffect(() => {
		logBuildInfo();
	}, []);

	useEffect(() => {
		const current = pathLocale ?? getStoredLocale();
		setLocaleState(current);
		document.documentElement.lang = current;
	}, [pathLocale]);

	const dict = useMemo(() => {
		return getDictionary(locale);
	}, [locale]);

	const i18nValue = useMemo(() => {
		return {
			locale,
			dict,
			t: (key: TranslationKey, params?: Record<string, string | number>) =>
				translate(dict, key, params),
			setLocale: (newLocale: Locale) => {
				setLocaleState(newLocale);
				saveLocalePreference(newLocale);
			},
			getLocalizedPath: (pathname: string, targetLocale?: Locale) =>
				buildLocalizedUrl(pathname, targetLocale ?? locale),
		};
	}, [locale, dict]);

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
		<html lang={DEFAULT_LOCALE} suppressHydrationWarning>
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
