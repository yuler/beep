import { CircleHelp } from "lucide-react";

import { Button } from "@/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
	DialogTrigger,
} from "@/components/ui/dialog";
import { useTranslation } from "@/lib/i18n";
import { browserLabel, pushPlatformLabel } from "@/lib/i18n-labels";
import {
	IOS_HOME_SCREEN_HINT,
	type NotificationPlatform,
} from "@/lib/web-push";

function osStepBody(
	t: ReturnType<typeof useTranslation>["t"],
	platform: NotificationPlatform,
	browserName: string,
) {
	const browser = browserLabel(t, browserName);

	switch (platform) {
		case "macos":
			return t("push.help_macos_body", { browser });
		case "windows":
			return t("push.help_windows_body", { browser });
		case "linux":
			return t("push.help_linux_body", { browser });
		case "ios":
			return t(IOS_HOME_SCREEN_HINT);
		default:
			return t("push.help_other_body", { browser });
	}
}

export function WebPushHelpDialog({
	browserName,
	platform,
}: {
	browserName: string;
	platform: NotificationPlatform;
}) {
	const { t } = useTranslation();
	const browser = browserLabel(t, browserName);
	const platformLabel = pushPlatformLabel(t, platform);

	return (
		<Dialog>
			<DialogTrigger
				render={
					<Button type="button" variant="outline" size="sm" className="w-fit" />
				}
			>
				<CircleHelp data-icon="inline-start" />
				{t("common.tips")}
			</DialogTrigger>
			<DialogContent className="sm:max-w-md">
				<DialogHeader>
					<DialogTitle>{t("push.help_title")}</DialogTitle>
					<DialogDescription>
						{t("push.help_description", { browser })}
					</DialogDescription>
				</DialogHeader>
				<ol className="flex list-decimal flex-col gap-4 pl-4 text-sm">
					<li>
						<span className="font-medium">{t("push.help_step_browser")}</span>{" "}
						<span className="text-muted-foreground">
							{t("push.help_browser_step")}
						</span>
					</li>
					<li>
						<span className="font-medium">
							{t("push.help_step_os", { platform: platformLabel })}
						</span>{" "}
						<span className="text-muted-foreground">
							{osStepBody(t, platform, browserName)}
						</span>
					</li>
					<li>
						<span className="font-medium">{t("push.help_step_test")}</span>{" "}
						<span className="text-muted-foreground">
							{t("push.help_step_test_body")}
						</span>
					</li>
				</ol>
				<DialogFooter showCloseButton />
			</DialogContent>
		</Dialog>
	);
}
