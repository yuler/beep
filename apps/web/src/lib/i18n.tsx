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

export interface I18nContextValue {
	locale: Locale;
	dict: TranslationSchema;
	t: (key: TranslationKey, params?: Record<string, string | number>) => string;
	setLocaleUrl: (newLocale: Locale) => string;
}

export const I18nContext = createContext<I18nContextValue>({
	locale: DEFAULT_LOCALE,
	dict: getDictionary(DEFAULT_LOCALE),
	t: (key, params) => translate(getDictionary(DEFAULT_LOCALE), key, params),
	setLocaleUrl: () => "/",
});

export function useI18n(): I18nContextValue {
	return useContext(I18nContext);
}

export function useTranslation() {
	const { t, locale, dict } = useI18n();
	return { t, locale, dict };
}

export {
	DEFAULT_LOCALE,
	isSupportedLocale,
	type Locale,
	SUPPORTED_LOCALES,
	type TranslationKey,
};
