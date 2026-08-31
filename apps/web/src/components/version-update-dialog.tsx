import { RefreshCw } from "lucide-react";

import { Button } from "@/components/ui/button";
import { useVersionPoll } from "@/hooks/use-version-poll";
import { buildInfo } from "@/lib/build-info";

const shortHash = (hash: string) =>
	hash === "unknown" ? hash : hash.slice(0, 7);

export function VersionUpdateDialog() {
	const { updateAvailable, deployed, confirmRefresh, declineRefresh } =
		useVersionPoll();

	if (!updateAvailable || !deployed) return null;

	return (
		<div className="pointer-events-none fixed inset-x-0 bottom-0 z-50 p-4">
			<div
				className="pointer-events-auto mx-auto flex max-w-lg flex-col gap-3 rounded-xl border bg-popover p-4 text-sm text-popover-foreground shadow-lg ring-1 ring-foreground/10"
				role="status"
			>
				<div className="flex flex-col gap-1">
					<p className="font-heading text-base leading-none font-medium">
						A new version is available
					</p>
					<p className="text-muted-foreground">
						beep was updated while this tab was open. Refresh to load the latest
						version{" "}
						<span className="tabular-nums">
							(v{deployed.version} · {shortHash(deployed.gitHash)}, currently v
							{buildInfo.version} · {shortHash(buildInfo.gitHash)}).
						</span>
					</p>
				</div>
				<div className="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
					<Button type="button" variant="outline" onClick={declineRefresh}>
						Not now
					</Button>
					<Button type="button" onClick={confirmRefresh}>
						<RefreshCw data-icon="inline-start" />
						Refresh
					</Button>
				</div>
			</div>
		</div>
	);
}
