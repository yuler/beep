import { apiFetch } from "@/lib/api/client";

export type AccountSettings = {
	id: string;
	name: string;
	slug: string;
	personal: boolean;
	email_channel_enabled: boolean;
};

export function fetchSettings(slug: string) {
	return apiFetch<AccountSettings>(`/api/v1/${slug}/settings`, {
		method: "GET",
	});
}

export function updateSettings(
	slug: string,
	body: { email_channel_enabled: boolean },
) {
	return apiFetch<AccountSettings>(`/api/v1/${slug}/settings`, {
		method: "PATCH",
		body,
	});
}
