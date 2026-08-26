import { apiFetch } from "@/lib/api/client";
import type { NotificationChannel } from "@/lib/notification-channels";
import { browserTimezone } from "@/lib/timezone";

export type TimezoneSource = "detected" | "manual";

export type AccountSettings = {
	id: string;
	name: string;
	slug: string;
	personal: boolean;
	notification_channels: NotificationChannel[];
	timezone: string | null;
	timezone_source: TimezoneSource | null;
};

export function fetchSettings(slug: string) {
	return apiFetch<AccountSettings>(`/api/v1/${slug}/settings`, {
		method: "GET",
	});
}

export function updateSettings(
	slug: string,
	body: {
		notification_channels?: NotificationChannel[];
		timezone?: string;
		timezone_source?: TimezoneSource;
	},
) {
	return apiFetch<AccountSettings>(`/api/v1/${slug}/settings`, {
		method: "PATCH",
		body,
	});
}

/** Fills User.timezone from the browser when this membership is still empty. */
export async function detectAccountTimezone(slug: string) {
	const settings = await fetchSettings(slug);
	if (settings.timezone) {
		return settings;
	}

	return updateSettings(slug, {
		timezone: browserTimezone(),
		timezone_source: "detected",
	});
}
