import {
	createPushSubscription,
	destroyPushSubscription,
	fetchWebPushConfig,
} from "@/lib/api/push";

const SERVICE_WORKER_URL = "/service-worker.js";
const STORAGE_PREFIX = "beep.pushSubscription.";

type StoredSubscription = {
	id: string;
	endpoint: string;
};

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

export async function isSubscribedForAccount(slug: string) {
	if (!readStoredSubscription(slug)) return false;
	return hasBrowserPushSubscription();
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
		const record = await createPushSubscription(slug, payload);
		storeSubscription(slug, {
			id: record.id,
			endpoint: record.endpoint,
		});
		return record;
	} catch (error) {
		await subscription.unsubscribe();
		clearStoredSubscription(slug);
		throw error;
	}
}

export async function disableWebPush(slug: string) {
	const stored = readStoredSubscription(slug);
	const registration = await navigator.serviceWorker.getRegistration("/");
	const subscription = await registration?.pushManager.getSubscription();

	if (subscription) {
		await subscription.unsubscribe();
	}

	if (stored) {
		try {
			await destroyPushSubscription(slug, stored.id);
		} catch {
			// Browser is already unsubscribed; ignore a missing server row.
		}
		clearStoredSubscription(slug);
	}
}

async function ensureServiceWorker() {
	const existing = await navigator.serviceWorker.getRegistration("/");
	if (existing) return existing;
	return navigator.serviceWorker.register(SERVICE_WORKER_URL, { scope: "/" });
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
	const applicationServerKey = urlBase64ToUint8Array(vapidPublicKey);
	try {
		return await registration.pushManager.subscribe({
			userVisibleOnly: true,
			applicationServerKey,
		});
	} catch {
		const existing = await registration.pushManager.getSubscription();
		await existing?.unsubscribe();
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

function storageKey(slug: string) {
	return `${STORAGE_PREFIX}${slug}`;
}

function readStoredSubscription(slug: string): StoredSubscription | null {
	try {
		const raw = localStorage.getItem(storageKey(slug));
		if (!raw) return null;
		const parsed = JSON.parse(raw) as StoredSubscription;
		if (!parsed.id || !parsed.endpoint) return null;
		return parsed;
	} catch {
		return null;
	}
}

function storeSubscription(slug: string, value: StoredSubscription) {
	localStorage.setItem(storageKey(slug), JSON.stringify(value));
}

function clearStoredSubscription(slug: string) {
	localStorage.removeItem(storageKey(slug));
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
