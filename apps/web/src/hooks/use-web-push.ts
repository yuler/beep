import { useCallback, useEffect, useState } from "react";

import { ApiError } from "@/lib/api/client";
import {
	disableWebPush,
	enableWebPush,
	getPushPermission,
	isIosDevice,
	isStandaloneDisplay,
	isSubscribedForAccount,
	isWebPushSupported,
} from "@/lib/web-push";

export type WebPushStatus = {
	supported: boolean;
	permission: NotificationPermission;
	subscribed: boolean;
	ios: boolean;
	standalone: boolean;
};

const INITIAL_STATUS: WebPushStatus = {
	supported: false,
	permission: "default",
	subscribed: false,
	ios: false,
	standalone: false,
};

function errorMessage(err: unknown) {
	if (err instanceof ApiError || err instanceof Error) return err.message;
	return "Something went wrong.";
}

export function useWebPush(slug: string) {
	const [status, setStatus] = useState<WebPushStatus>(INITIAL_STATUS);
	const [ready, setReady] = useState(false);
	const [pending, setPending] = useState(false);
	const [error, setError] = useState<string | null>(null);

	const refresh = useCallback(async () => {
		if (!isWebPushSupported()) {
			setStatus({
				supported: false,
				permission: "denied",
				subscribed: false,
				ios: isIosDevice(),
				standalone: isStandaloneDisplay(),
			});
			return;
		}

		const [permission, subscribed] = await Promise.all([
			getPushPermission(),
			isSubscribedForAccount(slug),
		]);
		setStatus({
			supported: true,
			permission,
			subscribed,
			ios: isIosDevice(),
			standalone: isStandaloneDisplay(),
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

	return { status, ready, pending, error, enable, disable };
}
