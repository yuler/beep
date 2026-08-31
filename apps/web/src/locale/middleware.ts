import {
	DEFAULT_LOCALE,
	isSupportedLocale,
	type Locale,
	SUPPORTED_LOCALES,
} from "@beep/locales";

/**
 * Normalizes locale strings (e.g. backward compat for "zh-CN" -> "zh")
 */
export function normalizeLocale(
	value: string | null | undefined,
): Locale | null {
	if (!value) return null;
	const lower = value.toLowerCase();
	if (lower === "zh-cn" || lower === "zh_cn" || lower === "zh") {
		return "zh";
	}
	return isSupportedLocale(value) ? (value as Locale) : null;
}

/**
 * Extracts a supported locale from pathname if present (e.g. /zh/my-account -> "zh")
 * Also backwards-compatible with /zh-CN/my-account.
 */
export function extractLocaleFromPath(pathname: string): {
	locale: Locale | null;
	cleanPath: string;
} {
	const segments = pathname.split("/").filter(Boolean);
	const first = segments[0];

	const normalized = normalizeLocale(first);
	if (normalized) {
		const remaining = `/${segments.slice(1).join("/")}`;
		return {
			locale: normalized,
			cleanPath: remaining === "/" ? "/" : remaining,
		};
	}

	return {
		locale: null,
		cleanPath: pathname.startsWith("/") ? pathname : `/${pathname}`,
	};
}

/**
 * Low-level URL de-localization function for TanStack Router `rewrite.input`
 * Maps localized URLs e.g. `/zh/about` -> `/about`, `/zh` -> `/`
 */
export function deLocalizeUrl(url: URL): URL {
	const newUrl = new URL(url.href);
	const segments = newUrl.pathname.split("/").filter(Boolean);
	const first = segments[0]?.toLowerCase();

	if (first === "zh" || first === "zh-cn" || first === "zh_cn") {
		segments.shift();
		newUrl.pathname = "/" + segments.join("/");
	}
	return newUrl;
}

/**
 * Low-level URL localization function for TanStack Router `rewrite.output`
 * Ensures /zh prefix is preserved consistently in SSR and client
 */
export function localizeUrl(url: URL, targetLocale?: Locale): URL {
	const newUrl = new URL(url.href);

	// Bypass non-app routes / static assets
	if (
		newUrl.pathname.startsWith("/api") ||
		newUrl.pathname.startsWith("/up") ||
		newUrl.pathname === "/version.json" ||
		newUrl.pathname === "/service-worker.js"
	) {
		return newUrl;
	}

	const segments = newUrl.pathname.split("/").filter(Boolean);
	const first = segments[0]?.toLowerCase();

	const isTargetZh =
		targetLocale === "zh" ||
		first === "zh" ||
		first === "zh-cn" ||
		first === "zh_cn" ||
		(typeof window !== "undefined" &&
			window.location.pathname.toLowerCase().startsWith("/zh"));

	if (first === "zh" || first === "zh-cn" || first === "zh_cn") {
		segments.shift();
	}

	if (isTargetZh) {
		newUrl.pathname =
			"/zh" + (segments.length > 0 ? "/" + segments.join("/") : "");
	} else {
		newUrl.pathname = "/" + segments.join("/");
	}

	return newUrl;
}

/**
 * Determines current active locale purely from URL pathname (no cookies)
 */
export function resolveLocaleFromUrl(pathname: string): Locale {
	const { locale: pathLocale } = extractLocaleFromPath(pathname);
	if (pathLocale) {
		return pathLocale;
	}

	return DEFAULT_LOCALE;
}

/**
 * Builds localized URL path based on target locale
 */
export function buildLocalizedPath(
	pathname: string,
	targetLocale: Locale,
	search?: string | Record<string, unknown>,
): string {
	const { cleanPath } = extractLocaleFromPath(pathname);
	const prefix = targetLocale === DEFAULT_LOCALE ? "" : `/${targetLocale}`;
	const formattedPath =
		`${prefix}${cleanPath === "/" && prefix !== "" ? "" : cleanPath}` || "/";

	if (!search) return formattedPath;

	if (typeof search === "string") {
		const cleanSearch = search.startsWith("?") ? search : `?${search}`;
		return `${formattedPath}${cleanSearch}`;
	}

	const searchEntries = Object.entries(search).filter(
		([_, v]) => v !== undefined && v !== null,
	);
	if (searchEntries.length === 0) {
		return formattedPath;
	}

	const searchStr = new URLSearchParams(
		searchEntries.map(([k, v]) => [k, String(v)]),
	).toString();

	return searchStr ? `${formattedPath}?${searchStr}` : formattedPath;
}

export { DEFAULT_LOCALE, isSupportedLocale, type Locale, SUPPORTED_LOCALES };
