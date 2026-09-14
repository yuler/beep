import { createFileRoute, Link } from "@tanstack/react-router";
import { CheckCircle2, Laptop, XCircle } from "lucide-react";
import { type FormEvent, useCallback, useState } from "react";
import { DashboardHeader } from "@/components/dashboard/dashboard-header";
import { Button, buttonVariants } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
	approveDeviceAuth,
	type DeviceAuthInfo,
	denyDeviceAuth,
	verifyDeviceCode,
} from "@/lib/api/channels";
import { ApiError } from "@/lib/api/client";
import { parseDeviceSearch } from "@/lib/device-search";
import { translateError } from "@/lib/i18n-labels";
import { m } from "@/locale/paraglide/messages";

export const Route = createFileRoute("/$account_slug/device/channel")({
	ssr: false,
	validateSearch: parseDeviceSearch,
	component: AccountChannelDeviceAuthPage,
});

function AccountChannelDeviceAuthPage() {
	const { account } = Route.useRouteContext();
	const search = Route.useSearch();
	const navigate = Route.useNavigate();
	const [code, setCode] = useState(search.code?.toUpperCase() ?? "");
	const [channelName, setChannelName] = useState("");
	const [authInfo, setAuthInfo] = useState<DeviceAuthInfo | null>(null);
	const [approvedName, setApprovedName] = useState<string | null>(null);
	const [verifying, setVerifying] = useState(false);
	const [submitting, setSubmitting] = useState(false);
	const [status, setStatus] = useState<
		"idle" | "verified" | "approved" | "denied"
	>("idle");
	const [error, setError] = useState<string | null>(null);

	// A 401 here means the session died after the page loaded (logout in
	// another tab, revoked session). Send the user to sign back in and bring
	// them back with the code pre-filled, instead of dead-ending on
	// "Unauthorized".
	const signInAgain = useCallback(
		(userCode: string) => {
			const cleaned = userCode.trim().toUpperCase();
			const returnTo = cleaned
				? `/${account.slug}/device/channel?code=${encodeURIComponent(cleaned)}`
				: `/${account.slug}/device/channel`;
			void navigate({ to: "/sign", search: { return_to: returnTo } });
		},
		[account.slug, navigate],
	);

	// NOTE: `search.code` only pre-fills the input below — it is never
	// verified automatically. Auto-verifying a linkable code lets an attacker
	// send `.../device/channel?code=ATTACKER-CODE` and get one click closer
	// to binding their device to the victim's account (RFC 8628 §8.4).
	const handleVerifyCode = useCallback(
		async (userCode: string) => {
			const cleaned = userCode.trim().toUpperCase();
			if (!cleaned) return;

			setVerifying(true);
			setError(null);
			try {
				const info = await verifyDeviceCode(account.slug, cleaned);
				setAuthInfo(info);
				setChannelName(info.channel_name || "CLI Device");
				setStatus("verified");
				void navigate({ search: {}, replace: true });
			} catch (err) {
				if (err instanceof ApiError && err.status === 401) {
					signInAgain(cleaned);
					return;
				}
				setError(err instanceof ApiError ? err.message : translateError(err));
				setStatus("idle");
			} finally {
				setVerifying(false);
			}
		},
		[account.slug, navigate, signInAgain],
	);

	async function handleApprove(e: FormEvent) {
		e.preventDefault();
		if (!authInfo) return;

		setSubmitting(true);
		setError(null);
		try {
			const res = await approveDeviceAuth(account.slug, {
				user_code: authInfo.user_code,
				channel_name: channelName.trim() || undefined,
			});
			setApprovedName(res.channel.name);
			setStatus("approved");
		} catch (err) {
			if (err instanceof ApiError && err.status === 401) {
				signInAgain(authInfo.user_code);
				return;
			}
			setError(err instanceof ApiError ? err.message : translateError(err));
		} finally {
			setSubmitting(false);
		}
	}

	async function handleDeny() {
		if (!authInfo) return;

		setSubmitting(true);
		setError(null);
		try {
			await denyDeviceAuth(account.slug, authInfo.user_code);
			setStatus("denied");
		} catch (err) {
			if (err instanceof ApiError && err.status === 401) {
				signInAgain(authInfo.user_code);
				return;
			}
			setError(err instanceof ApiError ? err.message : translateError(err));
		} finally {
			setSubmitting(false);
		}
	}

	return (
		<>
			<DashboardHeader
				breadcrumbs={[
					{
						label: m.nav_home(),
						to: "/$account_slug",
						params: { account_slug: account.slug },
					},
					{
						label: m.settings_channels_title(),
						to: "/$account_slug/settings/channels",
						params: { account_slug: account.slug },
					},
					{ label: "Authorize Device", isCurrentPage: true },
				]}
			/>

			<div className="flex flex-1 flex-col items-center justify-center p-4 md:p-8">
				<div className="w-full max-w-md">
					<div className="mb-4 flex items-center justify-between rounded-lg border border-border bg-muted/30 px-3 py-2 text-xs">
						<div className="flex items-center gap-2 text-muted-foreground">
							<span>Authorizing for:</span>
							<span className="font-semibold text-foreground">
								{account.name}
							</span>
							<span className="font-mono text-muted-foreground">
								({account.slug})
							</span>
						</div>
					</div>

					{status === "approved" ? (
						<Card className="text-center">
							<CardHeader className="flex flex-col items-center pb-2">
								<div className="mb-3 flex size-12 items-center justify-center rounded-full bg-primary/10 text-primary">
									<CheckCircle2 className="size-6" />
								</div>
								<CardTitle className="text-xl">Channel Connected!</CardTitle>
								<CardDescription>
									Device{" "}
									<span className="font-semibold text-foreground">
										{approvedName}
									</span>{" "}
									has been authorized.
								</CardDescription>
							</CardHeader>
							<CardContent className="space-y-4 pt-2">
								<p className="text-xs text-muted-foreground">
									You can close this tab and return to your terminal.
									Notifications routed to this channel will now be delivered to
									your device.
								</p>
								<div className="flex justify-center gap-2">
									<Link
										to="/$account_slug/settings/channels"
										params={{ account_slug: account.slug }}
										className={buttonVariants({
											variant: "outline",
											size: "sm",
										})}
									>
										Manage Channels
									</Link>
								</div>
							</CardContent>
						</Card>
					) : status === "denied" ? (
						<Card className="text-center">
							<CardHeader className="flex flex-col items-center pb-2">
								<div className="mb-3 flex size-12 items-center justify-center rounded-full bg-destructive/10 text-destructive">
									<XCircle className="size-6" />
								</div>
								<CardTitle className="text-xl">Request Denied</CardTitle>
								<CardDescription>
									The device authorization request has been rejected.
								</CardDescription>
							</CardHeader>
							<CardContent className="space-y-4 pt-2">
								<p className="text-xs text-muted-foreground">
									You can close this tab. To connect later, run{" "}
									<code className="rounded bg-muted px-1.5 py-0.5 font-mono text-xs">
										beep channel connect
									</code>{" "}
									in your terminal again.
								</p>
							</CardContent>
						</Card>
					) : (
						<Card>
							<CardHeader className="text-center">
								<div className="mx-auto mb-2 flex size-12 items-center justify-center rounded-xl bg-primary/10 text-primary">
									<Laptop className="size-6" />
								</div>
								<CardTitle className="text-xl">Connect CLI Channel</CardTitle>
								<CardDescription>
									Authorize your local terminal to receive notification alerts.
								</CardDescription>
							</CardHeader>

							<CardContent>
								{status === "idle" && (
									<form
										onSubmit={(e) => {
											e.preventDefault();
											void handleVerifyCode(code);
										}}
										className="space-y-4"
									>
										<div className="space-y-2">
											<Label htmlFor="user-code">One-Time Code</Label>
											<Input
												id="user-code"
												type="text"
												placeholder="ABCD-EFGH"
												value={code}
												onChange={(e) => setCode(e.target.value.toUpperCase())}
												className="font-mono text-center tracking-widest uppercase text-base"
												maxLength={9}
												required
												autoFocus
											/>
											<p className="text-xs text-muted-foreground">
												Enter the 8-character code displayed in your terminal.
												Never enter a code sent to you by someone else.
											</p>
										</div>

										{error && (
											<div className="rounded-md bg-destructive/10 p-3 text-xs text-destructive">
												{error}
											</div>
										)}

										<Button
											type="submit"
											className="w-full"
											disabled={verifying || !code.trim()}
										>
											{verifying ? "Verifying..." : "Continue"}
										</Button>
									</form>
								)}

								{status === "verified" && authInfo && (
									<form onSubmit={handleApprove} className="space-y-4">
										<div className="rounded-lg border bg-muted/40 p-3 space-y-1 text-xs">
											<div className="flex justify-between text-muted-foreground">
												<span>Code</span>
												<span className="font-mono font-semibold text-foreground">
													{authInfo.user_code}
												</span>
											</div>
											<div className="flex justify-between text-muted-foreground">
												<span>Status</span>
												<span className="capitalize text-foreground font-medium">
													{authInfo.status}
												</span>
											</div>
										</div>

										<div className="space-y-2">
											<Label htmlFor="channel-name">Channel Name</Label>
											<Input
												id="channel-name"
												type="text"
												value={channelName}
												onChange={(e) => setChannelName(e.target.value)}
												placeholder="e.g. Work MacBook Pro"
												maxLength={80}
												required
											/>
											<p className="text-xs text-muted-foreground">
												A friendly label so you know which device this is.
											</p>
										</div>

										{error && (
											<div className="rounded-md bg-destructive/10 p-3 text-xs text-destructive">
												{error}
											</div>
										)}

										<div className="flex gap-2 pt-2">
											<Button
												type="button"
												variant="outline"
												className="flex-1"
												onClick={handleDeny}
												disabled={submitting}
											>
												Deny
											</Button>
											<Button
												type="submit"
												className="flex-1"
												disabled={submitting || !channelName.trim()}
											>
												{submitting ? "Approving..." : "Approve Device"}
											</Button>
										</div>
									</form>
								)}
							</CardContent>
						</Card>
					)}
				</div>
			</div>
		</>
	);
}
