// beep-web-push v3 — keep this URL unhashed; bump the comment to roll clients.

self.addEventListener("install", (event) => {
	event.waitUntil(self.skipWaiting());
});

self.addEventListener("activate", (event) => {
	event.waitUntil(self.clients.claim());
});

self.addEventListener("push", (event) => {
	event.waitUntil(showPushNotification(event));
});

async function showPushNotification(event) {
	let title = "Beep";
	let options = { body: "You have a new notification." };

	try {
		const payload = event.data ? await event.data.json() : {};
		title = payload.title || title;
		options = payload.options || options;
	} catch {
		// Still show a notification so Chrome does not drop a focused-tab push.
	}

	await self.registration.showNotification(title, options);
	const badge = options.data?.badge;
	if (typeof self.navigator.setAppBadge === "function") {
		await self.navigator.setAppBadge(typeof badge === "number" ? badge : 0);
	}
}

self.addEventListener("notificationclick", (event) => {
	event.notification.close();
	const url = event.notification.data?.url;
	event.waitUntil(openOrFocus(url));
});

async function openOrFocus(url) {
	const origin = self.location.origin;
	const target = url ? new URL(url, origin).href : origin;
	const windows = await self.clients.matchAll({
		type: "window",
		includeUncontrolled: true,
	});

	for (const client of windows) {
		if (!client.url.startsWith(origin) || !("focus" in client)) {
			continue;
		}
		await client.focus();
		if ("navigate" in client) {
			await client.navigate(target);
		}
		return;
	}

	await self.clients.openWindow(target);
}
