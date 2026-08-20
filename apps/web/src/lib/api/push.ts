import {
	createPushSubscription as serverCreatePushSubscription,
	destroyPushSubscription as serverDestroyPushSubscription,
	fetchPushSubscriptions as serverFetchPushSubscriptions,
	fetchWebPushConfig as serverFetchWebPushConfig,
	testPushSubscription as serverTestPushSubscription,
} from "@/server/push";

export type {
	PushSubscriptionRecord,
	PushSubscriptionsResponse,
	WebPushConfig,
} from "@/server/push";

export function fetchWebPushConfig() {
	return serverFetchWebPushConfig({});
}

export function fetchPushSubscriptions(slug: string) {
	return serverFetchPushSubscriptions({ data: { slug } });
}

export function createPushSubscription(
	slug: string,
	body: { endpoint: string; p256dh_key: string; auth_key: string },
) {
	return serverCreatePushSubscription({ data: { slug, ...body } });
}

export function destroyPushSubscription(slug: string, id: string) {
	return serverDestroyPushSubscription({ data: { slug, id } });
}

export function testPushSubscription(slug: string, id: string) {
	return serverTestPushSubscription({ data: { slug, id } });
}
