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
import { useTranslation } from "@/lib/i18n";
import { browserLabel, osLabel } from "@/lib/i18n-labels";
import { describePushDevice, IOS_HOME_SCREEN_HINT } from "@/lib/web-push";

function localizedPushDevice(
	t: ReturnType<typeof useTranslation>["t"],
	userAgent: string | null,
) {
	const raw = describePushDevice(userAgent);
	if (!userAgent || raw === "Unknown device" || raw === "curl") return raw;

	const onIndex = raw.lastIndexOf(" on ");
	if (onIndex === -1) return browserLabel(t, raw);

	return t("push.device_description", {
		browser: browserLabel(t, raw.slice(0, onIndex)),
		os: osLabel(t, raw.slice(onIndex + 4)),
	});
}

export function WebPushSettings({ slug }: { slug: string }) {
	const { t } = useTranslation();
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
	const localizedBrowser = browserLabel(t, status.browserName);

	async function handleEnable() {
		const result = await probe();
		if (result.kind !== "unreachable") await enable();
	}

	return (
		<Card className="max-w-lg">
			<CardHeader>
				<CardTitle>{t("push.settings_title")}</CardTitle>
				<CardDescription>{t("push.settings_description")}</CardDescription>
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
						{t("push.checking_browser")}
					</p>
				) : !status.supported ? (
					<p className="text-sm text-muted-foreground">
						{t("push.unsupported")}
					</p>
				) : needsIosInstall ? (
					<p className="text-sm text-muted-foreground">
						{t(IOS_HOME_SCREEN_HINT)}
					</p>
				) : denied ? (
					<p className="text-sm text-muted-foreground">
						{t("push.permission_denied_retry")}
					</p>
				) : (
					<>
						<p className="text-sm text-muted-foreground">
							{status.subscribed
								? t("push.subscribed")
								: t("push.enable_each_device")}
						</p>
						{probing ? (
							<p className="text-sm text-muted-foreground">
								{t("push.checking_network")}
							</p>
						) : blocked ? (
							<p className="text-sm text-destructive" role="alert">
								{t("push.network_blocked", { browser: localizedBrowser })}
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
										? t("push.turning_off")
										: t("push.enabling")
									: probing
										? t("push.checking_browser")
										: status.subscribed
											? t("push.turn_off_browser")
											: t("push.enable_browser")}
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
									{testing ? t("beeps.sending") : t("common.test")}
								</Button>
							) : null}
						</div>
						{testSent ? (
							<output className="text-sm text-muted-foreground">
								{t("push.test_sent")}
							</output>
						) : null}
					</>
				)}

				{ready ? (
					<div className="flex flex-col gap-2">
						<p className="text-sm font-medium">{t("push.devices")}</p>
						{subscriptions.length === 0 ? (
							<p className="text-sm text-muted-foreground">
								{t("push.devices_empty")}
							</p>
						) : (
							<ul className="flex flex-col gap-2">
								{subscriptions.map((record) => {
									const current = record.endpoint === currentEndpoint;
									const deviceLabel = localizedPushDevice(
										t,
										record.user_agent,
									);
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
															{t("push.this_browser")}
														</Badge>
													) : null}
												</div>
												<p className="mt-0.5 text-xs text-muted-foreground tabular-nums">
													{t("push.device_added", {
														date: new Date(
															record.created_at,
														).toLocaleString(),
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
