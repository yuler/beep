import { RefreshCw } from "lucide-react";

import { Button } from "@/components/ui/button";
import { useVersionPoll } from "@/hooks/use-version-poll";
import { buildInfo } from "@/lib/build-info";
import { useTranslation } from "@/lib/i18n";

const shortHash = (hash: string) =>
	hash === "unknown" ? hash : hash.slice(0, 7);

export function VersionUpdateDialog() {
	const { t } = useTranslation();
	const { updateAvailable, deployed, confirmRefresh, declineRefresh } =
		useVersionPoll();

	if (!updateAvailable || !deployed) return null;

	return (
		<div className="pointer-events-none fixed inset-x-0 bottom-0 z-50 p-4">
			<div className="pointer-events-auto mx-auto flex max-w-lg flex-col gap-3 rounded-xl border bg-popover p-4 text-sm text-popover-foreground shadow-lg ring-1 ring-foreground/10">
				<div className="flex flex-col gap-1">
					<p className="font-heading text-base leading-none font-medium">
						{t("version.new_available")}
					</p>
					<p className="text-muted-foreground tabular-nums">
						{t("version.refresh_prompt", {
							newVersion: deployed.version,
							newHash: shortHash(deployed.gitHash),
							currentVersion: buildInfo.version,
							currentHash: shortHash(buildInfo.gitHash),
						})}
					</p>
				</div>
				<div className="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
					<Button type="button" variant="outline" onClick={declineRefresh}>
						{t("common.not_now")}
					</Button>
					<Button type="button" onClick={confirmRefresh}>
						<RefreshCw data-icon="inline-start" />
						{t("common.refresh")}
					</Button>
				</div>
			</div>
		</div>
	);
}
