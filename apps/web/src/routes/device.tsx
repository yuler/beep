import { createFileRoute } from "@tanstack/react-router";
import { CheckCircle2, Laptop, XCircle } from "lucide-react";
import { type FormEvent, useCallback, useEffect, useState } from "react";
import { AuthCard, AuthLayout, AuthPending } from "@/components/layout";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
	approveDeviceAuth,
	type DeviceAuthInfo,
	denyDeviceAuth,
	verifyDeviceCode,
} from "@/lib/api/channels";
import { ApiError } from "@/lib/api/client";
import { requireSession } from "@/lib/auth/guards";
import { translateError } from "@/lib/i18n-labels";

export type DeviceSearch = {
	code?: string;
};

export function parseDeviceSearch(
	search: Record<string, unknown>,
): DeviceSearch {
	const code =
		typeof search.code === "string" && search.code.trim() !== ""
			? search.code.trim()
			: undefined;
	return code ? { code } : {};
}

export const Route = createFileRoute("/device")({
	ssr: false,
	pendingComponent: AuthPending,
	validateSearch: parseDeviceSearch,
	beforeLoad: async ({ context, location }) => {
		const me = await requireSession({ context, location });
		return { me };
	},
	component: DeviceAuthPage,
});

function DeviceAuthPage() {
	const search = Route.useSearch();
	const [code, setCode] = useState(search.code?.toUpperCase() ?? "");
	const [channelName, setChannelName] = useState("");
	const [authInfo, setAuthInfo] = useState<DeviceAuthInfo | null>(null);
	const [verifying, setVerifying] = useState(false);
	const [submitting, setSubmitting] = useState(false);
	const [status, setStatus] = useState<
		"idle" | "verified" | "approved" | "denied"
	>("idle");
	const [error, setError] = useState<string | null>(null);

	const handleVerifyCode = useCallback(async (userCode: string) => {
		const cleaned = userCode.trim().toUpperCase();
		if (!cleaned) return;

		setVerifying(true);
		setError(null);
		try {
			const info = await verifyDeviceCode(cleaned);
			setAuthInfo(info);
			setChannelName(info.channel_name || "CLI Device");
			setStatus("verified");
		} catch (err) {
			setError(err instanceof ApiError ? err.message : translateError(err));
			setStatus("idle");
		} finally {
			setVerifying(false);
		}
	}, []);

	useEffect(() => {
		if (search.code && status === "idle") {
			void handleVerifyCode(search.code);
		}
	}, [search.code, status, handleVerifyCode]);

	async function handleApprove(e: FormEvent) {
		e.preventDefault();
		if (!authInfo) return;

		setSubmitting(true);
		setError(null);
		try {
			await approveDeviceAuth({
				user_code: authInfo.user_code,
				channel_name: channelName.trim() || undefined,
			});
			setStatus("approved");
		} catch (err) {
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
			await denyDeviceAuth(authInfo.user_code);
			setStatus("denied");
		} catch (err) {
			setError(err instanceof ApiError ? err.message : translateError(err));
		} finally {
			setSubmitting(false);
		}
	}

	return (
		<AuthLayout>
			<AuthCard description="Authorize a Beep CLI notification channel on your device.">
				{status === "approved" ? (
					<div className="flex flex-col items-center gap-4 py-4 text-center">
						<div className="flex size-12 items-center justify-center rounded-full bg-primary/10 text-primary">
							<CheckCircle2 className="size-8" />
						</div>
						<div className="flex flex-col gap-1">
							<h3 className="font-semibold text-base">Device Connected!</h3>
							<p className="text-muted-foreground text-sm">
								Your CLI channel{" "}
								<span className="font-medium text-foreground">
									{channelName}
								</span>{" "}
								has been connected. You can now close this tab and return to
								your terminal.
							</p>
						</div>
					</div>
				) : status === "denied" ? (
					<div className="flex flex-col items-center gap-4 py-4 text-center">
						<div className="flex size-12 items-center justify-center rounded-full bg-destructive/10 text-destructive">
							<XCircle className="size-8" />
						</div>
						<div className="flex flex-col gap-1">
							<h3 className="font-semibold text-base">Authorization Denied</h3>
							<p className="text-muted-foreground text-sm">
								The connection request has been rejected.
							</p>
						</div>
					</div>
				) : status === "verified" && authInfo ? (
					<form onSubmit={handleApprove} className="flex flex-col gap-5">
						<div className="rounded-lg border border-border bg-muted/40 p-3.5">
							<div className="flex items-center gap-3">
								<span className="flex size-9 shrink-0 items-center justify-center rounded-md bg-background border">
									<Laptop className="size-5 text-muted-foreground" />
								</span>
								<div className="flex flex-col min-w-0">
									<span className="text-xs text-muted-foreground">
										User Verification Code
									</span>
									<span className="font-mono font-bold tracking-widest text-base">
										{authInfo.user_code}
									</span>
								</div>
							</div>
						</div>

						<div className="flex flex-col gap-2">
							<Label htmlFor="channel-name" className="text-xs">
								Channel Name
							</Label>
							<Input
								id="channel-name"
								placeholder="e.g. My Laptop"
								value={channelName}
								onChange={(e) => setChannelName(e.target.value)}
								disabled={submitting}
								required
							/>
							<p className="text-[11px] text-muted-foreground">
								Identify this CLI channel in your notification settings.
							</p>
						</div>

						{error ? (
							<p className="text-sm text-destructive" role="alert">
								{error}
							</p>
						) : null}

						<div className="flex items-center gap-3 pt-2">
							<Button
								type="button"
								variant="outline"
								className="flex-1"
								onClick={handleDeny}
								disabled={submitting}
							>
								Deny
							</Button>
							<Button type="submit" className="flex-1" disabled={submitting}>
								{submitting ? "Connecting..." : "Authorize"}
							</Button>
						</div>
					</form>
				) : (
					<form
						onSubmit={(e) => {
							e.preventDefault();
							void handleVerifyCode(code);
						}}
						className="flex flex-col gap-4"
					>
						<div className="flex flex-col gap-2">
							<Label htmlFor="device-code" className="text-xs">
								Enter One-Time Code
							</Label>
							<Input
								id="device-code"
								placeholder="e.g. WDJB-MJHT"
								value={code}
								onChange={(e) => setCode(e.target.value.toUpperCase())}
								disabled={verifying}
								autoFocus
								className="font-mono uppercase tracking-wider"
							/>
							<p className="text-[11px] text-muted-foreground">
								Enter the 8-character code displayed in your terminal.
							</p>
						</div>

						{error ? (
							<p className="text-sm text-destructive" role="alert">
								{error}
							</p>
						) : null}

						<Button
							type="submit"
							disabled={verifying || !code.trim()}
							className="w-full"
						>
							{verifying ? "Checking Code..." : "Continue"}
						</Button>
					</form>
				)}
			</AuthCard>
		</AuthLayout>
	);
}
