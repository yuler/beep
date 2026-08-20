import type { NotificationChannel } from "@/lib/notification-channels";
import {
	fetchSettings as serverFetchSettings,
	updateSettings as serverUpdateSettings,
} from "@/server/settings";

export type { AccountSettings } from "@/server/settings";

export function fetchSettings(slug: string) {
	return serverFetchSettings({ data: { slug } });
}

export function updateSettings(
	slug: string,
	body: { notification_channels: NotificationChannel[] },
) {
	return serverUpdateSettings({
		data: { slug, notification_channels: body.notification_channels },
	});
}
