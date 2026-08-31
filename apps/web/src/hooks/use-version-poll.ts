import { useEffect, useRef, useState } from "react";

import { fetchDeployedVersion, isNewerDeploy } from "@/lib/api/version";
import { type BuildInfo, buildInfo } from "@/lib/build-info";

/** How often to check for a new deploy. */
export const VERSION_POLL_MS = 5 * 60 * 1000;

/** After the user declines, wait this long before prompting again. */
export const VERSION_DECLINE_MS = 30 * 60 * 1000;

const DISMISS_KEY = "beep:version-update-dismissed";

function dismissedRecently(): boolean {
	try {
		const raw = localStorage.getItem(DISMISS_KEY);
		if (!raw) return false;
		const dismissedAt = Number(raw);
		if (!Number.isFinite(dismissedAt)) return false;
		return Date.now() - dismissedAt < VERSION_DECLINE_MS;
	} catch {
		return false;
	}
}

export function recordVersionUpdateDismissed(): void {
	try {
		localStorage.setItem(DISMISS_KEY, String(Date.now()));
	} catch {
		// ignore quota / private mode
	}
}

export function useVersionPoll(): {
	updateAvailable: boolean;
	deployed: BuildInfo | null;
	confirmRefresh: () => void;
	declineRefresh: () => void;
} {
	const [updateAvailable, setUpdateAvailable] = useState(false);
	const [deployed, setDeployed] = useState<BuildInfo | null>(null);
	const openRef = useRef(false);

	useEffect(() => {
		openRef.current = updateAvailable;
	}, [updateAvailable]);

	useEffect(() => {
		if (import.meta.env.DEV) return;

		let cancelled = false;

		async function check() {
			try {
				const latest = await fetchDeployedVersion();
				if (cancelled) return;
				if (!isNewerDeploy(latest, buildInfo)) {
					setUpdateAvailable(false);
					setDeployed(null);
					return;
				}
				if (openRef.current || dismissedRecently()) return;
				setDeployed(latest);
				setUpdateAvailable(true);
			} catch {
				// Network blip — retry on the next interval.
			}
		}

		void check();
		const id = setInterval(() => void check(), VERSION_POLL_MS);
		return () => {
			cancelled = true;
			clearInterval(id);
		};
	}, []);

	function confirmRefresh() {
		window.location.reload();
	}

	function declineRefresh() {
		recordVersionUpdateDismissed();
		setUpdateAvailable(false);
	}

	return { updateAvailable, deployed, confirmRefresh, declineRefresh };
}
