import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { useWebPush } from "@/hooks/use-web-push";

export function WebPushSettings({ slug }: { slug: string }) {
	const { status, ready, pending, error, enable, disable } = useWebPush(slug);

	const needsIosInstall = status.ios && !status.standalone;
	const denied = status.supported && status.permission === "denied";

	return (
		<Card className="max-w-105">
			<CardHeader>
				<CardTitle>Browser notifications</CardTitle>
				<CardDescription>
					Get a system notification when a beep is due on this device.
				</CardDescription>
			</CardHeader>
			<CardContent className="flex flex-col gap-4">
				{!ready ? (
					<p className="text-sm text-muted-foreground">
						Checking this browser…
					</p>
				) : !status.supported ? (
					<p className="text-sm text-muted-foreground">
						This browser does not support web push.
					</p>
				) : needsIosInstall ? (
					<p className="text-sm text-muted-foreground">
						On iPhone and iPad, add Beep to your Home Screen, then open it from
						there to enable notifications.
					</p>
				) : denied ? (
					<p className="text-sm text-muted-foreground">
						Notifications are blocked. Enable them in your browser or system
						settings, then try again.
					</p>
				) : (
					<>
						<p className="text-sm text-muted-foreground">
							{status.subscribed
								? "This device is subscribed for this workspace."
								: "Chrome, Firefox, Edge, and desktop Safari are supported. iOS requires the Home Screen app."}
						</p>
						<Button
							type="button"
							className="w-fit"
							disabled={pending}
							variant={status.subscribed ? "outline" : "default"}
							onClick={() => {
								void (status.subscribed ? disable() : enable());
							}}
						>
							{pending
								? status.subscribed
									? "Turning off…"
									: "Enabling…"
								: status.subscribed
									? "Turn off"
									: "Enable notifications"}
						</Button>
					</>
				)}
				{error ? (
					<p className="text-sm text-destructive" role="alert">
						{error}
					</p>
				) : null}
			</CardContent>
		</Card>
	);
}
