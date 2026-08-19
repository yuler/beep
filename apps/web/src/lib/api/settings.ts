import { apiFetch } from "@/lib/api/client";
import type { NotificationChannel } from "@/lib/notification-channels";

export type AccountSettings = {
	id: string;
	name: string;
	slug: string;
	personal: boolean;
	notification_channels: NotificationChannel[];
};

export function fetchSettings(slug: string) {
	return apiFetch<AccountSettings>(`/api/v1/${slug}/settings`, {
		method: "GET",
	});
}

export function updateSettings(
	slug: string,
	body: { notification_channels: NotificationChannel[] },
) {
	return apiFetch<AccountSettings>(`/api/v1/${slug}/settings`, {
		method: "PATCH",
		body,
	});
}
