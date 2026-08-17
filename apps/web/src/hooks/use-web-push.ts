import { useCallback, useEffect, useState } from "react";

import { ApiError } from "@/lib/api/client";
import type { PushSubscriptionRecord } from "@/lib/api/push";
import {
	disableWebPush,
	enableWebPush,
	getBrowserPushEndpoint,
	getPushPermission,
	isIosDevice,
	isMacOS,
	isStandaloneDisplay,
	isSubscribedForAccount,
	isWebPushSupported,
	listPushSubscriptions,
	notificationBrowserName,
	notificationPlatform,
	removePushSubscription,
	sendTestPush,
} from "@/lib/web-push";

export type WebPushStatus = {
	supported: boolean;
	permission: NotificationPermission;
	subscribed: boolean;
	ios: boolean;
	macos: boolean;
	standalone: boolean;
	browserName: string;
	platform: ReturnType<typeof notificationPlatform>;
};

const INITIAL_STATUS: WebPushStatus = {
	supported: false,
	permission: "default",
	subscribed: false,
	ios: false,
	macos: false,
	standalone: false,
	browserName: "this browser",
	platform: "other",
};

function errorMessage(err: unknown) {
	if (err instanceof ApiError || err instanceof Error) return err.message;
	return "Something went wrong.";
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

	const refresh = useCallback(async () => {
		const endpoint = await getBrowserPushEndpoint();
		setCurrentEndpoint(endpoint);

		let records: PushSubscriptionRecord[] = [];
		try {
			records = await listPushSubscriptions(slug);
			setSubscriptions(records);
		} catch (err) {
			setError(errorMessage(err));
			setSubscriptions([]);
		}

		if (!isWebPushSupported()) {
			setStatus({
				supported: false,
				permission: "denied",
				subscribed: false,
				ios: isIosDevice(),
				macos: isMacOS(),
				standalone: isStandaloneDisplay(),
				browserName: notificationBrowserName(),
				platform: notificationPlatform(),
			});
			return;
		}

		const permission = await getPushPermission();
		setStatus({
			supported: true,
			permission,
			subscribed: isSubscribedForAccount(endpoint, records),
			ios: isIosDevice(),
			macos: isMacOS(),
			standalone: isStandaloneDisplay(),
			browserName: notificationBrowserName(),
			platform: notificationPlatform(),
		});
	}, [slug]);

	useEffect(() => {
		let cancelled = false;
		void refresh().finally(() => {
			if (!cancelled) setReady(true);
		});
		return () => {
			cancelled = true;
		};
	}, [refresh]);

	const enable = useCallback(async () => {
		setPending(true);
		setError(null);
		setTestSent(false);
		try {
			await enableWebPush(slug);
			await refresh();
		} catch (err) {
			setError(errorMessage(err));
			await refresh();
		} finally {
			setPending(false);
		}
	}, [refresh, slug]);

	const disable = useCallback(async () => {
		setPending(true);
		setError(null);
		setTestSent(false);
		try {
			await disableWebPush(slug);
			await refresh();
		} catch (err) {
			setError(errorMessage(err));
			await refresh();
		} finally {
			setPending(false);
		}
	}, [refresh, slug]);

	const sendTest = useCallback(async () => {
		setTesting(true);
		setError(null);
		setTestSent(false);
		try {
			await sendTestPush(slug);
			setTestSent(true);
			await refresh();
		} catch (err) {
			setError(errorMessage(err));
			await refresh();
		} finally {
			setTesting(false);
		}
	}, [refresh, slug]);

	const remove = useCallback(
		async (record: PushSubscriptionRecord) => {
			setRemovingId(record.id);
			setError(null);
			try {
				await removePushSubscription(slug, record);
				await refresh();
			} catch (err) {
				setError(errorMessage(err));
				await refresh();
			} finally {
				setRemovingId(null);
			}
		},
		[refresh, slug],
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
