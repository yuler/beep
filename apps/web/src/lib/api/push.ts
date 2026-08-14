import { apiFetch } from "@/lib/api/client";

export type WebPushConfig = {
	vapid_public_key: string;
};

export type PushSubscriptionRecord = {
	id: string;
	endpoint: string;
	user_agent: string | null;
	created_at: string;
};

export type PushSubscriptionsResponse = {
	push_subscriptions: PushSubscriptionRecord[];
};

export function fetchWebPushConfig() {
	return apiFetch<WebPushConfig>("/api/v1/web_push", { method: "GET" });
}

export function fetchPushSubscriptions(slug: string) {
	return apiFetch<PushSubscriptionsResponse>(
		`/api/v1/${slug}/push_subscriptions`,
		{ method: "GET" },
	);
}

export function createPushSubscription(
	slug: string,
	body: { endpoint: string; p256dh_key: string; auth_key: string },
) {
	return apiFetch<PushSubscriptionRecord>(
		`/api/v1/${slug}/push_subscriptions`,
		{
			method: "POST",
			body,
		},
	);
}

export function destroyPushSubscription(slug: string, id: string) {
	return apiFetch<void>(`/api/v1/${slug}/push_subscriptions/${id}`, {
		method: "DELETE",
	});
}

export function testPushSubscription(slug: string, id: string) {
	return apiFetch<void>(`/api/v1/${slug}/push_subscriptions/${id}/test`, {
		method: "POST",
	});
}
