// beep-web-push v1 — keep this URL unhashed; bump the comment to roll clients.

self.addEventListener("install", (event) => {
	event.waitUntil(self.skipWaiting());
});

self.addEventListener("activate", (event) => {
	event.waitUntil(self.clients.claim());
});

self.addEventListener("push", (event) => {
	const payload = event.data ? event.data.json() : {};
	const title = payload.title || "Beep";
	const options = payload.options || {};

	event.waitUntil(
		(async () => {
			await self.registration.showNotification(title, options);
			const badge = options.data?.badge;
			if (typeof badge === "number" && navigator.setAppBadge) {
				await navigator.setAppBadge(badge);
			}
		})(),
	);
});

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
