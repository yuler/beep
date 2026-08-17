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
import {
	IOS_HOME_SCREEN_HINT,
	type NotificationPlatform,
} from "@/lib/web-push";

const OS_STEP: Record<
	NotificationPlatform,
	{ title: string; body: (browserName: string) => string }
> = {
	macos: {
		title: "Mac",
		body: (browserName) =>
			`System Settings → Notifications → ${browserName}. Turn notifications on and set the style to Banners or Alerts.`,
	},
	windows: {
		title: "Windows",
		body: (browserName) =>
			`Settings → System → Notifications. Turn notifications on, then enable ${browserName} in the app list. Banners must be allowed.`,
	},
	linux: {
		title: "Linux",
		body: (browserName) =>
			`GNOME: Settings → Notifications → ${browserName}. KDE: System Settings → Notifications. Allow ${browserName} and turn off Do Not Disturb.`,
	},
	ios: {
		title: "iPhone and iPad",
		body: () => IOS_HOME_SCREEN_HINT,
	},
	other: {
		title: "System",
		body: (browserName) =>
			`Allow notifications for ${browserName} in this device’s system notification settings.`,
	},
};

export function WebPushHelpDialog({
	browserName,
	platform,
}: {
	browserName: string;
	platform: NotificationPlatform;
}) {
	const osStep = OS_STEP[platform];

	return (
		<Dialog>
			<DialogTrigger
				render={
					<Button type="button" variant="outline" size="sm" className="w-fit" />
				}
			>
				<CircleHelp data-icon="inline-start" />
				Tips
			</DialogTrigger>
			<DialogContent className="sm:max-w-md">
				<DialogHeader>
					<DialogTitle>Get notifications on this device</DialogTitle>
					<DialogDescription>
						Allow this site in the browser, then allow {browserName} in system
						settings. Click Test when both are on.
					</DialogDescription>
				</DialogHeader>
				<ol className="flex list-decimal flex-col gap-4 pl-4 text-sm">
					<li>
						<span className="font-medium">Browser.</span>{" "}
						<span className="text-muted-foreground">
							Click Enable notifications and allow the prompt for this site.
						</span>
					</li>
					<li>
						<span className="font-medium">{osStep.title}.</span>{" "}
						<span className="text-muted-foreground">
							{osStep.body(browserName)}
						</span>
					</li>
					<li>
						<span className="font-medium">Test.</span>{" "}
						<span className="text-muted-foreground">
							Click Test. You should see a system banner. A number on the app
							icon only appears after Beep is installed, not on a regular
							browser tab.
						</span>
					</li>
				</ol>
				<DialogFooter showCloseButton />
			</DialogContent>
		</Dialog>
	);
}
