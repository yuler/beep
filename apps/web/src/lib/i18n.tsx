import { createContext, useContext } from "react";
import {
	buildLocalizedPath,
	DEFAULT_LOCALE,
	extractLocaleFromPath,
	isSupportedLocale,
	normalizeLocale,
	resolveLocaleFromUrl,
	SUPPORTED_LOCALES,
} from "@/locale/middleware";
import {
	baseLocale,
	isLocale,
	type Locale,
	locales,
} from "@/locale/paraglide/runtime";

export interface I18nContextValue {
	locale: Locale;
	getLocalizedPath: (pathname: string, targetLocale?: Locale) => string;
}

export const I18nContext = createContext<I18nContextValue>({
	locale: DEFAULT_LOCALE,
	getLocalizedPath: (p) => p,
});

export function useI18n(): I18nContextValue {
	return useContext(I18nContext);
}

export {
	baseLocale,
	buildLocalizedPath,
	DEFAULT_LOCALE,
	extractLocaleFromPath,
	isLocale,
	isSupportedLocale,
	type Locale,
	locales,
	normalizeLocale,
	resolveLocaleFromUrl,
	SUPPORTED_LOCALES,
};
