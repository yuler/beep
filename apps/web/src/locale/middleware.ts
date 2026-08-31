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
 * Determines current active locale purely from URL pathname and browser settings (no cookies)
 */
export function resolveLocaleFromUrl(pathname: string): Locale {
	const { locale: pathLocale } = extractLocaleFromPath(pathname);
	if (pathLocale) {
		return pathLocale;
	}

	if (typeof window !== "undefined") {
		const navLang = navigator.language;
		if (navLang.startsWith("zh")) {
			return "zh";
		}
	}

	return DEFAULT_LOCALE;
}

/**
 * Builds localized URL based on target locale
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
