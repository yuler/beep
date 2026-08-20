import { createServerFn } from "@tanstack/react-start";

import { coreFetch } from "@/server/core";

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

export const fetchWebPushConfig = createServerFn({
	method: "GET",
	strict: false,
}).handler(async () =>
	coreFetch<WebPushConfig>("/api/v1/web_push", { method: "GET" }),
);

export const fetchPushSubscriptions = createServerFn({
	method: "GET",
	strict: false,
})
	.validator((s: { slug: string }) => s)
	.handler(({ data }) =>
		coreFetch<PushSubscriptionsResponse>(
			`/api/v1/${data.slug}/push_subscriptions`,
			{ method: "GET" },
		),
	);

export const createPushSubscription = createServerFn({
	method: "POST",
	strict: false,
})
	.validator(
		(s: {
			slug: string;
			endpoint: string;
			p256dh_key: string;
			auth_key: string;
		}) => s,
	)
	.handler(({ data }) =>
		coreFetch<PushSubscriptionRecord>(
			`/api/v1/${data.slug}/push_subscriptions`,
			{
				method: "POST",
				body: {
					endpoint: data.endpoint,
					p256dh_key: data.p256dh_key,
					auth_key: data.auth_key,
				},
			},
		),
	);

export const destroyPushSubscription = createServerFn({
	method: "POST",
	strict: false,
})
	.validator((s: { slug: string; id: string }) => s)
	.handler(({ data }) =>
		coreFetch<void>(`/api/v1/${data.slug}/push_subscriptions/${data.id}`, {
			method: "DELETE",
		}),
	);

export const testPushSubscription = createServerFn({
	method: "POST",
	strict: false,
})
	.validator((s: { slug: string; id: string }) => s)
	.handler(({ data }) =>
		coreFetch<void>(`/api/v1/${data.slug}/push_subscriptions/${data.id}/test`, {
			method: "POST",
		}),
	);
