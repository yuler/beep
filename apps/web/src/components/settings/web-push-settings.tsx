import { Bell, BellOff, BellRing, Trash2 } from "lucide-react";

import { WebPushHelpDialog } from "@/components/settings/web-push-help-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardAction,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { useWebPush } from "@/hooks/use-web-push";
import { describePushDevice } from "@/lib/web-push";

export function WebPushSettings({ slug }: { slug: string }) {
	const {
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
	} = useWebPush(slug);

	const needsIosInstall = status.ios && !status.standalone;
	const denied = status.supported && status.permission === "denied";
	const busy = pending || testing || removingId !== null;

	return (
		<Card className="max-w-lg">
			<CardHeader>
				<CardTitle>Browser notifications</CardTitle>
				<CardDescription>
					Devices that can receive a system notification when a beep is due.
				</CardDescription>
				<CardAction>
					<WebPushHelpDialog
						browserName={status.browserName}
						platform={status.platform}
					/>
				</CardAction>
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
								? "This browser is subscribed for this workspace."
								: "Enable on each device you want to notify."}
						</p>
						<div className="flex flex-wrap gap-2">
							<Button
								type="button"
								className="w-fit"
								disabled={busy}
								variant={status.subscribed ? "outline" : "default"}
								onClick={() => {
									void (status.subscribed ? disable() : enable());
								}}
							>
								{status.subscribed ? (
									<BellOff data-icon="inline-start" />
								) : (
									<Bell data-icon="inline-start" />
								)}
								{pending
									? status.subscribed
										? "Turning off…"
										: "Enabling…"
									: status.subscribed
										? "Turn off this browser"
										: "Enable this browser"}
							</Button>
							{status.subscribed ? (
								<Button
									type="button"
									className="w-fit"
									disabled={busy}
									onClick={() => {
										void sendTest();
									}}
								>
									<BellRing data-icon="inline-start" />
									{testing ? "Sending…" : "Test"}
								</Button>
							) : null}
						</div>
						{testSent ? (
							<output className="text-sm text-muted-foreground">
								Test sent. If nothing appeared, open Tips.
							</output>
						) : null}
					</>
				)}

				{ready ? (
					<div className="flex flex-col gap-2">
						<p className="text-sm font-medium">Devices</p>
						{subscriptions.length === 0 ? (
							<p className="text-sm text-muted-foreground">
								No devices subscribed yet.
							</p>
						) : (
							<ul className="flex flex-col gap-2">
								{subscriptions.map((record) => {
									const current = record.endpoint === currentEndpoint;
									return (
										<li
											key={record.id}
											className="flex items-start justify-between gap-3 rounded-lg border border-border px-3 py-2"
										>
											<div className="min-w-0">
												<div className="flex flex-wrap items-center gap-2">
													<p className="truncate text-sm font-medium">
														{describePushDevice(record.user_agent)}
													</p>
													{current ? (
														<Badge variant="secondary">This browser</Badge>
													) : null}
												</div>
												<p className="mt-0.5 text-xs text-muted-foreground tabular-nums">
													Added {new Date(record.created_at).toLocaleString()}
												</p>
											</div>
											<Button
												type="button"
												variant="ghost"
												size="icon-sm"
												disabled={busy}
												aria-label={`Remove ${describePushDevice(record.user_agent)}`}
												onClick={() => {
													void remove(record);
												}}
											>
												<Trash2 />
											</Button>
										</li>
									);
								})}
							</ul>
						)}
					</div>
				) : null}

				{error ? (
					<p className="text-sm text-destructive" role="alert">
						{error}
					</p>
				) : null}
			</CardContent>
		</Card>
	);
}
