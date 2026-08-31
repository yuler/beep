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
}

export function getStoredLocale(): Locale {
	if (typeof window === "undefined") {
		return DEFAULT_LOCALE;
	}

	// 1. LocalStorage
	const stored = localStorage.getItem(LOCALE_STORAGE_KEY);
	if (isSupportedLocale(stored)) {
		return stored;
	}

	// 2. Cookie
	const match = document.cookie.match(
		new RegExp(`(?:^|; )${LOCALE_COOKIE_KEY}=([^;]*)`),
	);
	if (match && isSupportedLocale(decodeURIComponent(match[1]))) {
		return decodeURIComponent(match[1]) as Locale;
	}

	// 3. Browser Navigator Language
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
});

export function useI18n(): I18nContextValue {
	return useContext(I18nContext);
}

export function useTranslation() {
	const { t, locale, dict, setLocale } = useI18n();
	return { t, locale, dict, setLocale };
}

export {
	DEFAULT_LOCALE,
	isSupportedLocale,
	type Locale,
	SUPPORTED_LOCALES,
	type TranslationKey,
};
