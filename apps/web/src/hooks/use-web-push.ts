import { useCallback, useEffect, useRef, useState } from "react";

import { ApiError } from "@/lib/api/client";
import {
	fetchPushSubscriptions,
	type PushSubscriptionRecord,
} from "@/lib/api/push";
import {
	disableWebPush,
	enableWebPush,
	getBrowserPushEndpoint,
	isStandaloneDisplay,
	isSubscribedForAccount,
	isWebPushSupported,
	type NotificationPlatform,
	notificationBrowserName,
	notificationPlatform,
	removePushSubscription,
	sendTestPush,
} from "@/lib/web-push";

export type WebPushStatus = {
	supported: boolean;
	permission: NotificationPermission;
	subscribed: boolean;
	standalone: boolean;
	browserName: string;
	platform: NotificationPlatform;
};

const INITIAL_STATUS: WebPushStatus = {
	supported: false,
	permission: "default",
	subscribed: false,
	standalone: false,
	browserName: "this browser",
	platform: "other",
};

type DeviceIdentity = Pick<
	WebPushStatus,
	"standalone" | "browserName" | "platform"
>;

function errorMessage(err: unknown) {
	if (err instanceof ApiError || err instanceof Error) return err.message;
	return "Something went wrong.";
}

function pushPermission() {
	if (!isWebPushSupported()) return "denied";
	return Notification.permission;
}

export function useWebPush(slug: string) {
	const [status, setStatus] = useState<WebPushStatus>(INITIAL_STATUS);
	const [ready, setReady] = useState(false);
	const [pending, setPending] = useState(false);
	const [testing, setTesting] = useState(false);
	const [testSent, setTestSent] = useState(false);
	const [removingId, setRemovingId] = useState<string | null>(null);
	const [subscriptions, setSubscriptions] = useState<PushSubscriptionRecord[]>(
		[],
	);
	const [currentEndpoint, setCurrentEndpoint] = useState<string | null>(null);
	const [error, setError] = useState<string | null>(null);
	const identityRef = useRef<DeviceIdentity | null>(null);

	const deviceIdentity = useCallback((): DeviceIdentity => {
		identityRef.current ??= {
			standalone: isStandaloneDisplay(),
			browserName: notificationBrowserName(),
			platform: notificationPlatform(),
		};
		return identityRef.current;
	}, []);

	const refresh = useCallback(async () => {
		const endpoint = await getBrowserPushEndpoint();
		setCurrentEndpoint(endpoint);

		let records: PushSubscriptionRecord[] = [];
		try {
			const response = await fetchPushSubscriptions(slug);
			records = response.push_subscriptions;
			setSubscriptions(records);
		} catch (err) {
			setError(errorMessage(err));
			setSubscriptions([]);
		}

		setStatus({
			...deviceIdentity(),
			supported: isWebPushSupported(),
			permission: pushPermission(),
			subscribed: isSubscribedForAccount(endpoint, records),
		});
	}, [deviceIdentity, slug]);

	useEffect(() => {
		let cancelled = false;
		void refresh().finally(() => {
			if (!cancelled) setReady(true);
		});
		return () => {
			cancelled = true;
		};
	}, [refresh]);

	const run = useCallback(
		async (
			action: () => Promise<void>,
			busy: {
				pending?: boolean;
				testing?: boolean;
				removingId?: string;
				markTestSent?: boolean;
			},
		) => {
			setError(null);
			if (busy.pending || busy.testing) setTestSent(false);
			if (busy.pending) setPending(true);
			if (busy.testing) setTesting(true);
			if (busy.removingId) setRemovingId(busy.removingId);
			try {
				await action();
				if (busy.markTestSent) setTestSent(true);
			} catch (err) {
				setError(errorMessage(err));
			} finally {
				await refresh();
				if (busy.pending) setPending(false);
				if (busy.testing) setTesting(false);
				if (busy.removingId) setRemovingId(null);
			}
		},
		[refresh],
	);

	const enable = useCallback(() => {
		return run(() => enableWebPush(slug), { pending: true });
	}, [run, slug]);

	const disable = useCallback(() => {
		return run(() => disableWebPush(slug), { pending: true });
	}, [run, slug]);

	const sendTest = useCallback(() => {
		return run(() => sendTestPush(slug), {
			testing: true,
			markTestSent: true,
		});
	}, [run, slug]);

	const remove = useCallback(
		(record: PushSubscriptionRecord) => {
			return run(() => removePushSubscription(slug, record), {
				removingId: record.id,
			});
		},
		[run, slug],
	);

	return {
		status,
		ready,
		pending,
		testing,
		testSent,
		removingId,
		subscriptions,
		currentEndpoint,
		error,
		enable,
		disable,
		sendTest,
		remove,
	};
}
