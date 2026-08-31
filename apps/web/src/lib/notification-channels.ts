export const NOTIFICATION_CHANNELS = ["email", "web_push"] as const;

export type NotificationChannel = (typeof NOTIFICATION_CHANNELS)[number];

export const CHANNEL_LABELS: Record<NotificationChannel, string> = {
	email: "Email",
	web_push: "Browser notifications",
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
