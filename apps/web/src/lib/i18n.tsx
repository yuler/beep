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
 * Extracts a supported locale from pathname if present (e.g. /zh-CN/my-account -> "zh-CN")
 */
export function extractLocaleFromPath(pathname: string): {
	locale: Locale | null;
	cleanPath: string;
} {
	const segments = pathname.split("/").filter(Boolean);
	const first = segments[0];

	if (first && isSupportedLocale(first)) {
		const remaining = `/${segments.slice(1).join("/")}`;
		return {
			locale: first,
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

	// 2. LocalStorage
	const stored = localStorage.getItem(LOCALE_STORAGE_KEY);
	if (isSupportedLocale(stored)) {
		return stored;
	}

	// 3. Cookie
	const match = document.cookie.match(
		new RegExp(`(?:^|; )${LOCALE_COOKIE_KEY}=([^;]*)`),
	);
	if (match && isSupportedLocale(decodeURIComponent(match[1]))) {
		return decodeURIComponent(match[1]) as Locale;
	}

	// 4. Browser Navigator Language
	const navLang = navigator.language;
	if (navLang.startsWith("zh")) {
		return "zh-CN";
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
