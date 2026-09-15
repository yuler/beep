import { createFileRoute, Link } from "@tanstack/react-router";
import { CheckCircle2, Terminal, XCircle } from "lucide-react";
import { type FormEvent, useCallback, useState } from "react";
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
	approveCliDeviceAuth,
	type CliDeviceAuthInfo,
	denyCliDeviceAuth,
	verifyCliDeviceCode,
} from "@/lib/api/cli-auth";
import { ApiError } from "@/lib/api/client";
import { resolveDashboardTarget } from "@/lib/auth/account";
import { requireSession } from "@/lib/auth/guards";
import { parseDeviceSearch } from "@/lib/device-search";
import { translateError } from "@/lib/i18n-labels";

export const Route = createFileRoute("/device/cli")({
	ssr: false,
	validateSearch: parseDeviceSearch,
	beforeLoad: async ({ context, location }) => {
		const me = await requireSession({ context, location });
		return { me };
	},
	component: CliDeviceAuthPage,
});

function CliDeviceAuthPage() {
	const { me } = Route.useRouteContext();
	const search = Route.useSearch();
	const navigate = Route.useNavigate();
	const [code, setCode] = useState(search.code?.toUpperCase() ?? "");
	const [clientName, setClientName] = useState("");
	const [authInfo, setAuthInfo] = useState<CliDeviceAuthInfo | null>(null);
	const [verifying, setVerifying] = useState(false);
	const [submitting, setSubmitting] = useState(false);
	const [status, setStatus] = useState<
		"idle" | "verified" | "approved" | "denied"
	>("idle");
	const [error, setError] = useState<string | null>(null);
	const dashboardTarget = resolveDashboardTarget(me.accounts);

	const handleVerifyCode = useCallback(
		async (userCode: string) => {
			const cleaned = userCode.trim().toUpperCase();
			if (!cleaned) return;

			setVerifying(true);
			setError(null);
			try {
				const info = await verifyCliDeviceCode(cleaned);
				setAuthInfo(info);
				setClientName(info.client_name || "Beep CLI");
				setStatus("verified");
				void navigate({ search: {}, replace: true });
			} catch (err) {
				setError(err instanceof ApiError ? err.message : translateError(err));
				setStatus("idle");
			} finally {
				setVerifying(false);
			}
		},
		[navigate],
	);

	async function handleApprove(e: FormEvent) {
		e.preventDefault();
		if (!authInfo) return;

		setSubmitting(true);
		setError(null);
		try {
			await approveCliDeviceAuth({
				user_code: authInfo.user_code,
				client_name: clientName.trim() || undefined,
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
			await denyCliDeviceAuth(authInfo.user_code);
			setStatus("denied");
		} catch (err) {
			setError(err instanceof ApiError ? err.message : translateError(err));
		} finally {
			setSubmitting(false);
		}
	}

	return (
		<div className="flex min-h-[80vh] flex-1 flex-col items-center justify-center p-4 md:p-8">
			<div className="w-full max-w-md">
				<div className="mb-4 flex items-center justify-between rounded-lg border border-border bg-muted/30 px-3 py-2 text-xs">
					<div className="flex items-center gap-2 text-muted-foreground">
						<span>Authorizing as:</span>
						<span className="font-semibold text-foreground">
							{me.identity.email}
						</span>
					</div>
				</div>

				{status === "approved" ? (
					<Card className="text-center">
						<CardHeader className="flex flex-col items-center pb-2">
							<div className="mb-3 flex size-12 items-center justify-center rounded-full bg-primary/10 text-primary">
								<CheckCircle2 className="size-6" />
							</div>
							<CardTitle className="text-xl">CLI Authorized!</CardTitle>
							<CardDescription>
								Your terminal is now logged in as{" "}
								<span className="font-semibold text-foreground">
									{me.identity.email}
								</span>
								.
							</CardDescription>
						</CardHeader>
						<CardContent className="space-y-4 pt-2">
							<p className="text-xs text-muted-foreground">
								You can close this tab and return to your terminal.
							</p>
							{dashboardTarget.kind === "account" ? (
								<Link
									to="/$account_slug"
									params={{ account_slug: dashboardTarget.slug }}
									className={buttonVariants({
										variant: "outline",
										size: "sm",
									})}
								>
									Go to Dashboard
								</Link>
							) : (
								<Link
									to="/accounts"
									className={buttonVariants({
										variant: "outline",
										size: "sm",
									})}
								>
									Go to Dashboard
								</Link>
							)}
						</CardContent>
					</Card>
				) : status === "denied" ? (
					<Card className="text-center">
						<CardHeader className="flex flex-col items-center pb-2">
							<div className="mb-3 flex size-12 items-center justify-center rounded-full bg-destructive/10 text-destructive">
								<XCircle className="size-6" />
							</div>
							<CardTitle className="text-xl">Authorization Denied</CardTitle>
							<CardDescription>
								The CLI connection request has been declined.
							</CardDescription>
						</CardHeader>
						<CardContent className="pt-2">
							<Button
								variant="outline"
								size="sm"
								onClick={() => {
									setStatus("idle");
									setAuthInfo(null);
									setCode("");
								}}
							>
								Try Another Code
							</Button>
						</CardContent>
					</Card>
				) : status === "verified" && authInfo ? (
					<Card>
						<CardHeader className="text-center">
							<div className="mx-auto mb-2 flex size-10 items-center justify-center rounded-full bg-primary/10 text-primary">
								<Terminal className="size-5" />
							</div>
							<CardTitle className="text-xl">Approve CLI Login</CardTitle>
							<CardDescription>
								Confirm that you want to log in from your terminal.
							</CardDescription>
						</CardHeader>
						<CardContent>
							<form onSubmit={handleApprove} className="space-y-4">
								<div className="rounded-lg border border-border bg-muted/40 p-3 text-center">
									<div className="text-xs text-muted-foreground">
										Verification Code
									</div>
									<div className="font-mono text-2xl font-bold tracking-widest text-primary">
										{authInfo.user_code}
									</div>
								</div>

								<div className="space-y-1.5">
									<Label htmlFor="client-name" className="text-xs">
										Device / Client Name
									</Label>
									<Input
										id="client-name"
										value={clientName}
										onChange={(e) => setClientName(e.target.value)}
										placeholder="Beep CLI"
										disabled={submitting}
									/>
								</div>

								{error && (
									<p className="text-center text-xs text-destructive">
										{error}
									</p>
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
										disabled={submitting}
									>
										{submitting ? "Approving..." : "Approve"}
									</Button>
								</div>
							</form>
						</CardContent>
					</Card>
				) : (
					<Card>
						<CardHeader className="text-center">
							<div className="mx-auto mb-2 flex size-10 items-center justify-center rounded-full bg-primary/10 text-primary">
								<Terminal className="size-5" />
							</div>
							<CardTitle className="text-xl">Authorize Beep CLI</CardTitle>
							<CardDescription>
								Enter the one-time verification code shown in your terminal.
							</CardDescription>
						</CardHeader>
						<CardContent>
							<form
								onSubmit={(e) => {
									e.preventDefault();
									void handleVerifyCode(code);
								}}
								className="space-y-4"
							>
								<div className="space-y-1.5">
									<Label htmlFor="code" className="text-xs">
										Verification Code
									</Label>
									<Input
										id="code"
										value={code}
										onChange={(e) => setCode(e.target.value.toUpperCase())}
										placeholder="ABCD-EFGH"
										className="text-center font-mono text-lg uppercase tracking-widest"
										maxLength={9}
										autoFocus
										disabled={verifying}
									/>
								</div>

								{error && (
									<p className="text-center text-xs text-destructive">
										{error}
									</p>
								)}

								<Button
									type="submit"
									className="w-full"
									disabled={verifying || !code.trim()}
								>
									{verifying ? "Checking..." : "Continue"}
								</Button>
							</form>
						</CardContent>
					</Card>
				)}
			</div>
		</div>
	);
}
