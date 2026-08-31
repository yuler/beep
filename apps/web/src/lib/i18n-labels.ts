import type { Beep } from "@/lib/api/beeps";
import { I18nError } from "@/lib/notification-channels";
import type { TranslationKey, TranslationSchema } from "@/lib/i18n";
import type { NotificationChannel } from "@/lib/notification-channels";

type Translate = (
	key: TranslationKey,
	params?: Record<string, string | number>,
) => string;

export function isTranslationKey(
	dict: TranslationSchema,
	key: string,
): key is TranslationKey {
	return key in dict;
}

export function translateError(
	dict: TranslationSchema,
	t: Translate,
	err: unknown,
): string {
	if (err instanceof I18nError) {
		return t(err.key, err.params);
	}
	if (err instanceof Error && isTranslationKey(dict, err.message)) {
		return t(err.message);
	}
	if (err instanceof Error && err.message) {
		return err.message;
	}
	return t("errors.something_went_wrong");
}

export function beepStatusLabel(t: Translate, status: Beep["status"]): string {
	return t(`status.beep.${status}` as TranslationKey);
}

export function beepRunStatusLabel(t: Translate, status: string): string {
	const key = `status.run.${status}`;
	if (key === "status.run.pending" || key === "status.run.running" || key === "status.run.succeeded" || key === "status.run.failed" || key === "status.run.skipped" || key === "status.run.expired") {
		return t(key as TranslationKey);
	}
	return status;
}

export function healthStatusLabel(t: Translate, status: string): string {
	const key = `status.health.${status}` as TranslationKey;
	return t(key);
}

export function jobStatusLabel(t: Translate, status: string): string {
	const key = `status.job.${status}` as TranslationKey;
	return t(key);
}

export function channelLabel(
	t: Translate,
	channel: NotificationChannel,
): string {
	return channel === "email"
		? t("push.channel_email")
		: t("push.channel_web_push");
}

const BROWSER_KEYS: Record<string, TranslationKey> = {
	"Google Chrome": "push.browser_chrome",
	"Microsoft Edge": "push.browser_edge",
	Firefox: "push.browser_firefox",
	Opera: "push.browser_opera",
	Safari: "push.browser_safari",
	"this browser": "push.browser_this",
};

export function browserLabel(t: Translate, browserName: string): string {
	const key = BROWSER_KEYS[browserName];
	return key ? t(key) : browserName;
}

const OS_KEYS: Record<string, TranslationKey> = {
	iOS: "push.os_ios",
	Mac: "push.os_macos",
	Windows: "push.os_windows",
	Linux: "push.os_linux",
	Android: "push.os_android",
};

export function osLabel(t: Translate, os: string): string {
	const key = OS_KEYS[os];
	return key ? t(key) : os;
}

export function pushPlatformLabel(
	t: Translate,
	platform: "ios" | "macos" | "windows" | "linux" | "other",
): string {
	switch (platform) {
		case "ios":
			return t("push.help_os_ios");
		case "macos":
			return t("push.help_os_macos");
		case "windows":
			return t("push.help_os_windows");
		case "linux":
			return t("push.help_os_linux");
		default:
			return t("push.help_os_other");
	}
}
