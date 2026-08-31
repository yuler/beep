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

// Helper to flatten key paths for autocompletion (e.g. 'common.save' | 'auth.login')
type NestedKeyOf<ObjectType extends object> = {
	[Key in keyof ObjectType & (string | number)]: ObjectType[Key] extends object
		? `${Key}` | `${Key}.${NestedKeyOf<ObjectType[Key]>}`
		: `${Key}`;
}[keyof ObjectType & (string | number)];

export type TranslationKey = NestedKeyOf<TranslationSchema>;

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
 * Resolve nested key translation with fallback and parameter interpolation
 * Example: translate(en, "common.save") => "Save"
 * Example: translate(en, "greeting.user", { name: "John" }) => "Hello John"
 */
export function translate(
	dict: TranslationSchema,
	key: TranslationKey,
	params?: Record<string, string | number>,
): string {
	const keys = key.split(".");
	let result: unknown = dict;

	for (const k of keys) {
		if (result && typeof result === "object" && k in result) {
			result = (result as Record<string, unknown>)[k];
		} else {
			// Fallback to key itself if not found
			return key;
		}
	}

	if (typeof result !== "string") {
		return key;
	}

	if (params) {
		return result.replace(/\{(\w+)\}/g, (_, p) =>
			p in params ? String(params[p]) : `{${p}}`,
		);
	}

	return result;
}
