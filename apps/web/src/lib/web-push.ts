import {
	createPushSubscription,
	destroyPushSubscription,
	fetchPushSubscriptions,
	fetchWebPushConfig,
	testPushSubscription,
} from "@/lib/api/push";

const SERVICE_WORKER_URL = "/service-worker.js";

export function isWebPushSupported() {
	return (
		typeof window !== "undefined" &&
		"serviceWorker" in navigator &&
		"PushManager" in window &&
		"Notification" in window
	);
}

export function isIosDevice() {
	if (typeof navigator === "undefined") return false;
	return /iphone|ipad|ipod/i.test(navigator.userAgent);
}

export function isMacOS() {
	if (typeof navigator === "undefined") return false;
	if (isIosDevice()) return false;
	if (!/Mac OS X|Macintosh/i.test(navigator.userAgent)) return false;
	// iPadOS "Request Desktop Website" uses a Macintosh UA.
	return navigator.maxTouchPoints <= 1;
}

export type NotificationPlatform =
	| "ios"
	| "macos"
	| "windows"
	| "linux"
	| "other";

export function notificationPlatform(): NotificationPlatform {
	if (typeof navigator === "undefined") return "other";
	if (isIosDevice()) return "ios";
	if (isMacOS()) return "macos";
	const ua = navigator.userAgent;
	if (/Android/i.test(ua)) return "other";
	if (/Windows/i.test(ua)) return "windows";
	if (/Linux|X11/i.test(ua)) return "linux";
	return "other";
}

export function notificationBrowserName() {
	if (typeof navigator === "undefined") return "this browser";
	return browserNameFromUserAgent(navigator.userAgent);
}

export function describePushDevice(userAgent: string | null) {
	if (!userAgent) return "Unknown device";
	if (/^curl\//i.test(userAgent)) return "curl";

	const browser = browserNameFromUserAgent(userAgent);
	if (/iPhone|iPad|iPod/i.test(userAgent)) return `${browser} on iOS`;
	if (/Android/i.test(userAgent)) return `${browser} on Android`;
	if (/Mac OS X|Macintosh/i.test(userAgent)) return `${browser} on Mac`;
	if (/Windows/i.test(userAgent)) return `${browser} on Windows`;
	if (/Linux|X11/i.test(userAgent)) return `${browser} on Linux`;
	return browser;
}

function browserNameFromUserAgent(ua: string) {
	if (/Edg\//.test(ua)) return "Microsoft Edge";
	if (/OPR\/|Opera/.test(ua)) return "Opera";
	if (/Firefox\//.test(ua)) return "Firefox";
	if (/Safari/.test(ua) && !/Chrome|Chromium|Edg\//.test(ua)) return "Safari";
	if (/Chrome|Chromium/.test(ua)) return "Google Chrome";
	return "this browser";
}

export function isStandaloneDisplay() {
	if (typeof window === "undefined") return false;
	const media = window.matchMedia("(display-mode: standalone)").matches;
	const safariStandalone =
		"standalone" in navigator &&
		Boolean((navigator as Navigator & { standalone?: boolean }).standalone);
	return media || safariStandalone;
}

export async function getPushPermission(): Promise<NotificationPermission> {
	if (!isWebPushSupported()) return "denied";
	return Notification.permission;
}

export async function hasBrowserPushSubscription() {
	if (!isWebPushSupported()) return false;
	const registration = await navigator.serviceWorker.getRegistration("/");
	if (!registration) return false;
	const subscription = await registration.pushManager.getSubscription();
	return subscription !== null;
}

export function isSubscribedForAccount(
	endpoint: string | null,
	records: { endpoint: string }[],
) {
	return (
		endpoint !== null && records.some((record) => record.endpoint === endpoint)
	);
}

export async function enableWebPush(slug: string) {
	if (!isWebPushSupported()) {
		throw new Error("This browser does not support web push.");
	}

	const { vapid_public_key: vapidPublicKey } = await fetchWebPushConfig();
	const registration = await ensureServiceWorker();
	await ensureNotificationPermission();

	const subscription = await subscribePushManager(registration, vapidPublicKey);
	const payload = pushSubscriptionPayload(subscription);

	try {
		return await createPushSubscription(slug, payload);
	} catch (error) {
		await subscription.unsubscribe();
		throw error;
	}
}

export async function disableWebPush(slug: string) {
	const endpoint = await getBrowserPushEndpoint();
	const registration = await navigator.serviceWorker.getRegistration("/");
	const subscription = await registration?.pushManager.getSubscription();

	if (subscription) {
		await subscription.unsubscribe();
	}

	const id = await subscriptionIdForEndpoint(slug, endpoint);
	if (!id) return;

	try {
		await destroyPushSubscription(slug, id);
	} catch {
		// Browser is already unsubscribed; ignore a missing server row.
	}
}

export async function getBrowserPushEndpoint() {
	if (!isWebPushSupported()) return null;
	const registration = await navigator.serviceWorker.getRegistration("/");
	const subscription = await registration?.pushManager.getSubscription();
	return subscription?.endpoint ?? null;
}

export async function listPushSubscriptions(slug: string) {
	const { push_subscriptions } = await fetchPushSubscriptions(slug);
	return push_subscriptions;
}

export async function removePushSubscription(
	slug: string,
	record: { id: string; endpoint: string },
) {
	const currentEndpoint = await getBrowserPushEndpoint();
	const isCurrent = currentEndpoint === record.endpoint;

	await destroyPushSubscription(slug, record.id);

	if (isCurrent) {
		const registration = await navigator.serviceWorker.getRegistration("/");
		const subscription = await registration?.pushManager.getSubscription();
		if (subscription) await subscription.unsubscribe();
	}
}

export async function sendTestPush(slug: string) {
	const endpoint = await getBrowserPushEndpoint();
	const id = await subscriptionIdForEndpoint(slug, endpoint);
	if (!id) {
		throw new Error("This device is not subscribed.");
	}

	await testPushSubscription(slug, id);
}

async function ensureServiceWorker() {
	await navigator.serviceWorker.register(SERVICE_WORKER_URL, { scope: "/" });
	return navigator.serviceWorker.ready;
}

async function ensureNotificationPermission() {
	if (Notification.permission === "granted") return;
	if (Notification.permission === "denied") {
		throw new Error(
			"Notifications are blocked. Enable them in your browser or system settings.",
		);
	}

	const permission = await Notification.requestPermission();
	if (permission !== "granted") {
		throw new Error("Notification permission was not granted.");
	}
}

async function subscribePushManager(
	registration: ServiceWorkerRegistration,
	vapidPublicKey: string,
) {
	const existing = await registration.pushManager.getSubscription();
	if (existing) return existing;

	const applicationServerKey = urlBase64ToUint8Array(vapidPublicKey);
	try {
		return await registration.pushManager.subscribe({
			userVisibleOnly: true,
			applicationServerKey,
		});
	} catch {
		const leftover = await registration.pushManager.getSubscription();
		await leftover?.unsubscribe();
		return registration.pushManager.subscribe({
			userVisibleOnly: true,
			applicationServerKey,
		});
	}
}

function pushSubscriptionPayload(subscription: PushSubscription) {
	const json = subscription.toJSON();
	const endpoint = json.endpoint;
	const p256dh = json.keys?.p256dh;
	const auth = json.keys?.auth;

	if (!endpoint || !p256dh || !auth) {
		throw new Error("Push subscription is missing endpoint or keys.");
	}

	return { endpoint, p256dh_key: p256dh, auth_key: auth };
}

async function subscriptionIdForEndpoint(slug: string, endpoint: string | null) {
	if (!endpoint) return null;
	const records = await listPushSubscriptions(slug);
	return records.find((record) => record.endpoint === endpoint)?.id ?? null;
}

function urlBase64ToUint8Array(base64String: string) {
	const padding = "=".repeat((4 - (base64String.length % 4)) % 4);
	const base64 = (base64String + padding).replace(/-/g, "+").replace(/_/g, "/");
	const raw = window.atob(base64);
	const output = new Uint8Array(raw.length);
	for (let i = 0; i < raw.length; i += 1) {
		output[i] = raw.charCodeAt(i);
	}
	return output;
}
