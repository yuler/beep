import {
	DEFAULT_LOCALE,
	getDictionary,
	isSupportedLocale,
	type Locale,
	SUPPORTED_LOCALES,
	type TranslationKey,
	type TranslationSchema,
	translate,
} from "@beep/locales";
import { createContext, useContext } from "react";

const LOCALE_COOKIE_KEY = "locale";
const LOCALE_STORAGE_KEY = "beep_locale";

export interface I18nContextValue {
	locale: Locale;
	dict: TranslationSchema;
	t: (key: TranslationKey, params?: Record<string, string | number>) => string;
	setLocale: (newLocale: Locale) => void;
	getLocalizedPath: (pathname: string, targetLocale?: Locale) => string;
}

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
	return isSupportedLocale(value) ? value : null;
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
 * Builds localized URL based on given or current locale
 */
export function buildLocalizedUrl(
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

export function getStoredLocale(): Locale {
	if (typeof window === "undefined") {
		return DEFAULT_LOCALE;
	}

	// 1. URL path priority
	const { locale: pathLocale } = extractLocaleFromPath(
		window.location.pathname,
	);
	if (pathLocale) {
		return pathLocale;
	}

	// 2. LocalStorage (with backward compat)
	const stored = localStorage.getItem(LOCALE_STORAGE_KEY);
	const normalizedStored = normalizeLocale(stored);
	if (normalizedStored) {
		return normalizedStored;
	}

	// 3. Cookie (with backward compat)
	const match = document.cookie.match(
		new RegExp(`(?:^|; )${LOCALE_COOKIE_KEY}=([^;]*)`),
	);
	if (match) {
		const cookieVal = decodeURIComponent(match[1]);
		const normalizedCookie = normalizeLocale(cookieVal);
		if (normalizedCookie) {
			return normalizedCookie;
		}
	}

	// 4. Browser Navigator Language
	const navLang = navigator.language;
	if (navLang.startsWith("zh")) {
		return "zh";
	}

	return DEFAULT_LOCALE;
}

export function saveLocalePreference(locale: Locale) {
	if (typeof window === "undefined") return;

	localStorage.setItem(LOCALE_STORAGE_KEY, locale);
	// Set 1-year cookie for SSR and Core API requests
	// biome-ignore lint/suspicious/noDocumentCookie: Cookie for SSR and Core
	document.cookie = `${LOCALE_COOKIE_KEY}=${locale}; path=/; max-age=31536000; SameSite=Lax`;
	document.documentElement.lang = locale;
}

export const I18nContext = createContext<I18nContextValue>({
	locale: DEFAULT_LOCALE,
	dict: getDictionary(DEFAULT_LOCALE),
	t: (key, params) => translate(getDictionary(DEFAULT_LOCALE), key, params),
	setLocale: () => {},
	getLocalizedPath: (p) => p,
});

export function useI18n(): I18nContextValue {
	return useContext(I18nContext);
}

export function useTranslation() {
	const { t, locale, dict, setLocale, getLocalizedPath } = useI18n();
	return { t, locale, dict, setLocale, getLocalizedPath };
}

export {
	DEFAULT_LOCALE,
	isSupportedLocale,
	type Locale,
	SUPPORTED_LOCALES,
	type TranslationKey,
};
