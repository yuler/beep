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
import { browserLabel, osLabel } from "@/lib/i18n-labels";
import { describePushDevice, iosHomeScreenHint } from "@/lib/web-push";
import * as m from "@/locale/paraglide/messages";

function localizedPushDevice(userAgent: string | null) {
	const raw = describePushDevice(userAgent);
	if (!userAgent || raw === "Unknown device" || raw === "curl") return raw;

	const onIndex = raw.lastIndexOf(" on ");
	if (onIndex === -1) return browserLabel(raw);

	return m.push_device_description({
		browser: browserLabel(raw.slice(0, onIndex)),
		os: osLabel(raw.slice(onIndex + 4)),
	});
}

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
		reachability,
		error,
		probe,
		enable,
		disable,
		sendTest,
		remove,
	} = useWebPush(slug);

	const needsIosInstall = status.platform === "ios" && !status.standalone;
	const denied = status.supported && status.permission === "denied";
	const probing = reachability.kind === "probing";
	const blocked = reachability.kind === "unreachable";
	const busy = pending || testing || probing || removingId !== null;
	const localizedBrowser = browserLabel(status.browserName);

	async function handleEnable() {
		const result = await probe();
		if (result.kind !== "unreachable") await enable();
	}

	return (
		<Card className="max-w-lg">
			<CardHeader>
				<CardTitle>{m.push_settings_title()}</CardTitle>
				<CardDescription>{m.push_settings_description()}</CardDescription>
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
						{m.push_checking_browser()}
					</p>
				) : !status.supported ? (
					<p className="text-sm text-muted-foreground">
						{m.push_unsupported()}
					</p>
				) : needsIosInstall ? (
					<p className="text-sm text-muted-foreground">{iosHomeScreenHint()}</p>
				) : denied ? (
					<p className="text-sm text-muted-foreground">
						{m.push_permission_denied_retry()}
					</p>
				) : (
					<>
						<p className="text-sm text-muted-foreground">
							{status.subscribed
								? m.push_subscribed()
								: m.push_enable_each_device()}
						</p>
						{probing ? (
							<p className="text-sm text-muted-foreground">
								{m.push_checking_network()}
							</p>
						) : blocked ? (
							<p className="text-sm text-destructive" role="alert">
								{m.push_network_blocked({ browser: localizedBrowser })}
							</p>
						) : null}
						<div className="flex flex-wrap gap-2">
							<Button
								type="button"
								className="w-fit"
								disabled={busy}
								variant={status.subscribed ? "outline" : "default"}
								onClick={() => {
									void (status.subscribed
										? disable()
										: handleEnable()
												.then(() => undefined)
												.catch(() => undefined));
								}}
							>
								{status.subscribed ? (
									<BellOff data-icon="inline-start" />
								) : (
									<Bell data-icon="inline-start" />
								)}
								{pending
									? status.subscribed
										? m.push_turning_off()
										: m.push_enabling()
									: probing
										? m.push_checking_browser()
										: status.subscribed
											? m.push_turn_off_browser()
											: m.push_enable_browser()}
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
									{testing ? m.beeps_sending() : m.common_test()}
								</Button>
							) : null}
						</div>
						{testSent ? (
							<output className="text-sm text-muted-foreground">
								{m.push_test_sent()}
							</output>
						) : null}
					</>
				)}

				{ready ? (
					<div className="flex flex-col gap-2">
						<p className="text-sm font-medium">{m.push_devices()}</p>
						{subscriptions.length === 0 ? (
							<p className="text-sm text-muted-foreground">
								{m.push_devices_empty()}
							</p>
						) : (
							<ul className="flex flex-col gap-2">
								{subscriptions.map((record) => {
									const current = record.endpoint === currentEndpoint;
									const deviceLabel = localizedPushDevice(t, record.user_agent);
									return (
										<li
											key={record.id}
											className="flex items-start justify-between gap-3 rounded-lg border border-border px-3 py-2"
										>
											<div className="min-w-0">
												<div className="flex flex-wrap items-center gap-2">
													<p className="truncate text-sm font-medium">
														{deviceLabel}
													</p>
													{current ? (
														<Badge variant="secondary">
															{m.push_this_browser()}
														</Badge>
													) : null}
												</div>
												<p className="mt-0.5 text-xs text-muted-foreground tabular-nums">
													{m.push_device_added({
														date: new Date(record.created_at).toLocaleString(),
													})}
												</p>
											</div>
											<Button
												type="button"
												variant="ghost"
												size="icon-sm"
												disabled={busy}
												aria-label={deviceLabel}
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
