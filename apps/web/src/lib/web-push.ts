import {
	createPushSubscription,
	destroyPushSubscription,
	fetchPushSubscriptions,
	fetchWebPushConfig,
	testPushSubscription,
} from "@/lib/api/push";
import { I18nError } from "@/lib/notification-channels";
import { m } from "@/locale/paraglide/messages";

const SERVICE_WORKER_URL = "/service-worker.js";
const PROBE_TIMEOUT_MS = 8_000;
const SUBSCRIBE_TIMEOUT_MS = 15_000;

function pushServiceHostForBrowser(browserName: string) {
	if (browserName === "Firefox") return "updates.push.services.mozilla.com";
	if (browserName === "Safari") return "web.push.apple.com";
	if (/Chrome|Edge|Opera/.test(browserName)) return "fcm.googleapis.com";
	return null;
}

async function probePushServiceReachable(host: string) {
	const controller = new AbortController();
	const timer = setTimeout(() => controller.abort(), PROBE_TIMEOUT_MS);
	try {
		await fetch(`https://${host}/`, {
			method: "GET",
			mode: "no-cors",
			signal: controller.signal,
		});
		return true;
	} catch {
		return false;
	} finally {
		clearTimeout(timer);
	}
}

function withTimeout<T>(promise: Promise<T>, ms: number, message: string) {
	let timer: ReturnType<typeof setTimeout> | undefined;
	const timeout = new Promise<never>((_resolve, reject) => {
		timer = setTimeout(() => reject(new Error(message)), ms);
	});
	return Promise.race([promise, timeout]).finally(() => clearTimeout(timer));
}

export function isWebPushSupported() {
	return (
		typeof window !== "undefined" &&
		"serviceWorker" in navigator &&
		"PushManager" in window &&
		"Notification" in window
	);
}

export type PushReachability =
	| { kind: "unsupported" }
	| { kind: "unknown" }
	| { kind: "reachable" }
	| { kind: "unreachable"; host: string };

export async function probePushServiceReachability(): Promise<PushReachability> {
	if (!isWebPushSupported()) return { kind: "unsupported" };
	const host = pushServiceHostForBrowser(notificationBrowserName());
	if (!host) return { kind: "unknown" };
	const reachable = await probePushServiceReachable(host);
	return reachable ? { kind: "reachable" } : { kind: "unreachable", host };
}

export type NotificationPlatform =
	| "ios"
	| "macos"
	| "windows"
	| "linux"
	| "other";

type DetectedPlatform = NotificationPlatform | "android";

export const iosHomeScreenHint = () => m.push_ios_home_screen_hint();

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
		throw new I18nError(m.push_unsupported());
	}

	const reachability = await probePushServiceReachability();
	if (reachability.kind === "unreachable") {
		throw new I18nError(
			m.push_network_unreachable({ host: reachability.host }),
		);
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
		throw new I18nError(m.push_device_not_subscribed());
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
		throw new I18nError(m.push_permission_denied());
	}

	const permission = await Notification.requestPermission();
	if (permission !== "granted") {
		throw new I18nError(m.push_permission_not_granted());
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
		return await timedSubscribe(registration, applicationServerKey);
	} catch {
		const leftover = await registration.pushManager.getSubscription();
		await leftover?.unsubscribe();
		return timedSubscribe(registration, applicationServerKey);
	}
}

async function timedSubscribe(
	registration: ServiceWorkerRegistration,
	applicationServerKey: Uint8Array<ArrayBuffer>,
) {
	const message = "push.subscribe_timeout";
	try {
		return await withTimeout(
			registration.pushManager.subscribe({
				userVisibleOnly: true,
				applicationServerKey,
			}),
			SUBSCRIBE_TIMEOUT_MS,
			message,
		);
	} catch (error) {
		if (error instanceof Error && error.message === message) {
			// The browser may still finish subscribing after we gave up; undo it
			// so the browser and server stay in sync.
			void registration.pushManager
				.getSubscription()
				.then((sub) => sub?.unsubscribe())
				.catch(() => undefined);
		}
		throw error;
	}
}

function pushSubscriptionPayload(subscription: PushSubscription) {
	const json = subscription.toJSON();
	const endpoint = json.endpoint;
	const p256dh = json.keys?.p256dh;
	const auth = json.keys?.auth;

	if (!endpoint || !p256dh || !auth) {
		throw new I18nError(m.push_subscription_missing_keys());
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
