import { m } from "@/locale/paraglide/messages";
import {
	baseLocale,
	deLocalizeHref,
	getLocale,
	type Locale,
	locales,
	localizeHref,
} from "@/locale/paraglide/runtime";

export {
	baseLocale,
	deLocalizeHref,
	getLocale,
	locales,
	localizeHref,
	type Locale,
	m,
};

export interface LocaleConfig {
	flag?: string;
	name: string;
	hreflang: string;
}

export const localeConfig = {
	en: {
		flag: "🇺🇸",
		name: "English",
		hreflang: "en",
	},
	zh: {
		flag: "🇨🇳",
		name: "中文",
		hreflang: "zh-CN",
	},
} satisfies Record<Locale, LocaleConfig>;

export function parseMessageJson<T>(value: string, fallback: T): T {
	try {
		return JSON.parse(value) as T;
	} catch {
		return fallback;
	}
}

export function getMessageList(value: string): string[] {
	return parseMessageJson<string[]>(value, []);
}

export function getCanonicalPathname(pathname: string): string {
	return deLocalizeHref(pathname).split("?")[0]?.split("#")[0] ?? pathname;
}
