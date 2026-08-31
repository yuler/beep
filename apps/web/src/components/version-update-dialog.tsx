import { RefreshCw } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import { useVersionPoll } from "@/hooks/use-version-poll";
import { buildInfo } from "@/lib/build-info";

const shortHash = (hash: string) =>
	hash === "unknown" ? hash : hash.slice(0, 7);

export function VersionUpdateDialog() {
	const { updateAvailable, deployed, confirmRefresh, declineRefresh } =
		useVersionPoll();

	return (
		<Dialog
			open={updateAvailable}
			onOpenChange={(open) => {
				if (!open) declineRefresh();
			}}
		>
			<DialogContent showCloseButton={false}>
				<DialogHeader>
					<DialogTitle>A new version is available</DialogTitle>
					<DialogDescription>
						beep was updated while this tab was open. Refresh to load the latest
						version
						{deployed ? (
							<>
								{" "}
								(
								<span className="tabular-nums">
									v{deployed.version} · {shortHash(deployed.gitHash)}
								</span>
								, currently v{buildInfo.version} ·{" "}
								{shortHash(buildInfo.gitHash)}).
							</>
						) : (
							"."
						)}
					</DialogDescription>
				</DialogHeader>
				<DialogFooter>
					<Button type="button" variant="outline" onClick={declineRefresh}>
						Not now
					</Button>
					<Button type="button" onClick={confirmRefresh}>
						<RefreshCw data-icon="inline-start" />
						Refresh
					</Button>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	);
}
