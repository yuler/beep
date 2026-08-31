import {
	DEFAULT_LOCALE,
	getDictionary,
	type Locale,
	SUPPORTED_LOCALES,
	type TranslationKey,
	type TranslationSchema,
	translate,
} from "@beep/locales";
import { createContext, useContext } from "react";
import {
	buildLocalizedPath,
	extractLocaleFromPath,
	isSupportedLocale,
	normalizeLocale,
	resolveLocaleFromUrl,
} from "@/locale/middleware";

export interface I18nContextValue {
	locale: Locale;
	dict: TranslationSchema;
	t: (key: TranslationKey, params?: Record<string, string | number>) => string;
	getLocalizedPath: (pathname: string, targetLocale?: Locale) => string;
}

export const I18nContext = createContext<I18nContextValue>({
	locale: DEFAULT_LOCALE,
	dict: getDictionary(DEFAULT_LOCALE),
	t: (key, params) => translate(getDictionary(DEFAULT_LOCALE), key, params),
	getLocalizedPath: (p) => p,
});

export function useI18n(): I18nContextValue {
	return useContext(I18nContext);
}

export function useTranslation() {
	const { t, locale, dict, getLocalizedPath } = useI18n();
	return { t, locale, dict, getLocalizedPath };
}

export {
	DEFAULT_LOCALE,
	isSupportedLocale,
	type Locale,
	SUPPORTED_LOCALES,
	type TranslationKey,
	buildLocalizedPath,
	extractLocaleFromPath,
	normalizeLocale,
	resolveLocaleFromUrl,
};
