import type { TranslationKey } from "@/lib/i18n";

export class I18nError extends Error {
	constructor(
		public readonly key: TranslationKey,
		public readonly params?: Record<string, string | number>,
	) {
		super(key);
		this.name = "I18nError";
	}
}

export const NOTIFICATION_CHANNELS = ["email", "web_push"] as const;

export type NotificationChannel = (typeof NOTIFICATION_CHANNELS)[number];

export const CHANNEL_LABEL_KEYS: Record<NotificationChannel, TranslationKey> = {
	email: "push.channel_email",
	web_push: "push.channel_web_push",
};

export function toggleChannel(
	channels: NotificationChannel[],
	channel: NotificationChannel,
	enabled: boolean,
): NotificationChannel[] {
	if (enabled) {
		if (channels.includes(channel)) return channels;
		return [...channels, channel];
	}
	return channels.filter((item) => item !== channel);
}
