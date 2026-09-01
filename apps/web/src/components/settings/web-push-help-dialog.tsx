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
import { browserLabel, pushPlatformLabel } from "@/lib/i18n-labels";
import { iosHomeScreenHint, type NotificationPlatform } from "@/lib/web-push";
import { m } from "@/locale/paraglide/messages";

function osStepBody(platform: NotificationPlatform, browserName: string) {
	const browser = browserLabel(browserName);

	switch (platform) {
		case "macos":
			return m.push_help_macos_body({ browser });
		case "windows":
			return m.push_help_windows_body({ browser });
		case "linux":
			return m.push_help_linux_body({ browser });
		case "ios":
			return iosHomeScreenHint();
		default:
			return m.push_help_other_body({ browser });
	}
}

export function WebPushHelpDialog({
	browserName,
	platform,
}: {
	browserName: string;
	platform: NotificationPlatform;
}) {
	const browser = browserLabel(browserName);
	const platformLabel = pushPlatformLabel(platform);

	return (
		<Dialog>
			<DialogTrigger
				render={
					<Button type="button" variant="outline" size="sm" className="w-fit" />
				}
			>
				<CircleHelp data-icon="inline-start" />
				{m.common_tips()}
			</DialogTrigger>
			<DialogContent className="sm:max-w-md">
				<DialogHeader>
					<DialogTitle>{m.push_help_title()}</DialogTitle>
					<DialogDescription>
						{m.push_help_description({ browser })}
					</DialogDescription>
				</DialogHeader>
				<ol className="flex list-decimal flex-col gap-4 pl-4 text-sm">
					<li>
						<span className="font-medium">{m.push_help_step_browser()}</span>{" "}
						<span className="text-muted-foreground">
							{m.push_help_browser_step()}
						</span>
					</li>
					<li>
						<span className="font-medium">
							{m.push_help_step_os({ platform: platformLabel })}
						</span>{" "}
						<span className="text-muted-foreground">
							{osStepBody(platform, browserName)}
						</span>
					</li>
					<li>
						<span className="font-medium">{m.push_help_step_test()}</span>{" "}
						<span className="text-muted-foreground">
							{m.push_help_step_test_body()}
						</span>
					</li>
				</ol>
				<DialogFooter showCloseButton />
			</DialogContent>
		</Dialog>
	);
}
