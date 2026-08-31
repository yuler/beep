import en from "./en.json";
import zhCN from "./zh-CN.json";

export const SUPPORTED_LOCALES = ["en", "zh-CN"] as const;
export type Locale = (typeof SUPPORTED_LOCALES)[number];
export const DEFAULT_LOCALE: Locale = "en";

export const dictionaries = {
	en,
	"zh-CN": zhCN,
} as const satisfies Record<Locale, typeof en>;

export type TranslationSchema = typeof en;
export type TranslationKey = keyof TranslationSchema;

export function getDictionary(locale: string | undefined): TranslationSchema {
	if (locale && locale in dictionaries) {
		return dictionaries[locale as Locale];
	}
	return dictionaries[DEFAULT_LOCALE];
}

export function isSupportedLocale(locale: string | undefined): locale is Locale {
	return Boolean(locale && SUPPORTED_LOCALES.includes(locale as Locale));
}

/**
 * Resolve flat key translation with fallback and parameter interpolation
 * Example: translate(en, "common.save") => "Save"
 * Example: translate(en, "greeting.user", { name: "John" }) => "Hello John"
 */
export function translate(
	dict: TranslationSchema,
	key: TranslationKey | string,
	params?: Record<string, string | number>,
): string {
	const raw = (dict as Record<string, string>)[key];
	if (!raw) {
		return key;
	}

	if (params) {
		return raw.replace(/\{(\w+)\}/g, (_, p) =>
			p in params ? String(params[p]) : `{${p}}`,
		);
	}

	return raw;
}
