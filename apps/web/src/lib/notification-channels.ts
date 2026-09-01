export class I18nError extends Error {
	constructor(message: string) {
		super(message);
		this.name = "I18nError";
	}
}

export const NOTIFICATION_CHANNELS = ["email", "web_push"] as const;

export type NotificationChannel = (typeof NOTIFICATION_CHANNELS)[number];

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
