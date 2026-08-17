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

export type NotificationPlatform =
	| "ios"
	| "macos"
	| "windows"
	| "linux"
	| "other";

type DetectedPlatform = NotificationPlatform | "android";

export const IOS_HOME_SCREEN_HINT =
	"On iPhone and iPad, add Beep to your Home Screen, then open it from there to enable notifications.";

export function notificationPlatform(): NotificationPlatform {
	if (typeof navigator === "undefined") return "other";
	const detected = platformFromUserAgent(navigator.userAgent, {
		maxTouchPoints: navigator.maxTouchPoints,
	});
	return detected === "android" ? "other" : detected;
}

export function notificationBrowserName() {
	if (typeof navigator === "undefined") return "this browser";
	return browserNameFromUserAgent(navigator.userAgent);
}

export function describePushDevice(userAgent: string | null) {
	if (!userAgent) return "Unknown device";
	if (/^curl\//i.test(userAgent)) return "curl";

	const browser = browserNameFromUserAgent(userAgent);
	const os = DEVICE_OS_LABEL[platformFromUserAgent(userAgent)];
	return os ? `${browser} on ${os}` : browser;
}

const DEVICE_OS_LABEL: Record<DetectedPlatform, string | null> = {
	ios: "iOS",
	macos: "Mac",
	windows: "Windows",
	linux: "Linux",
	android: "Android",
	other: null,
};

function platformFromUserAgent(
	ua: string,
	options?: { maxTouchPoints?: number },
): DetectedPlatform {
	if (/iphone|ipad|ipod/i.test(ua)) return "ios";
	if (/Mac OS X|Macintosh/i.test(ua)) {
		// iPadOS "Request Desktop Website" uses a Macintosh UA.
		if ((options?.maxTouchPoints ?? 0) > 1) return "ios";
		return "macos";
	}
	if (/Android/i.test(ua)) return "android";
	if (/Windows/i.test(ua)) return "windows";
	if (/Linux|X11/i.test(ua)) return "linux";
	return "other";
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

async function subscriptionIdForEndpoint(
	slug: string,
	endpoint: string | null,
) {
	if (!endpoint) return null;
	const { push_subscriptions } = await fetchPushSubscriptions(slug);
	return (
		push_subscriptions.find((record) => record.endpoint === endpoint)?.id ??
		null
	);
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
