import { createServerFn } from "@tanstack/react-start";
import type { NotificationChannel } from "@/lib/notification-channels";
import { coreFetch } from "@/server/core";

export type AccountSettings = {
	id: string;
	name: string;
	slug: string;
	personal: boolean;
	notification_channels: NotificationChannel[];
};

export const fetchSettings = createServerFn({ method: "GET", strict: false })
	.validator((s: { slug: string }) => s)
	.handler(({ data }) =>
		coreFetch<AccountSettings>(`/api/v1/${data.slug}/settings`, {
			method: "GET",
		}),
	);

export const updateSettings = createServerFn({ method: "POST", strict: false })
	.validator(
		(s: { slug: string; notification_channels: NotificationChannel[] }) => s,
	)
	.handler(({ data }) =>
		coreFetch<AccountSettings>(`/api/v1/${data.slug}/settings`, {
			method: "PATCH",
			body: { notification_channels: data.notification_channels },
		}),
	);
