import { createFileRoute, Link } from "@tanstack/react-router";
import { CheckCircle2, Server, XCircle } from "lucide-react";
import { type FormEvent, useCallback, useState } from "react";
import { DashboardHeader } from "@/components/dashboard/dashboard-header";
import { Badge } from "@/components/ui/badge";
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
import { ApiError } from "@/lib/api/client";
import {
	approveRunnerDeviceAuth,
	denyRunnerDeviceAuth,
	type RunnerDeviceAuthInfo,
	verifyRunnerDeviceCode,
} from "@/lib/api/runners";
import { parseDeviceSearch } from "@/lib/device-search";
import { translateError } from "@/lib/i18n-labels";
import { m } from "@/locale/paraglide/messages";

export const Route = createFileRoute("/$account_slug/device/runner")({
	ssr: false,
	validateSearch: parseDeviceSearch,
	component: AccountRunnerDeviceAuthPage,
});

function AccountRunnerDeviceAuthPage() {
	const { account } = Route.useRouteContext();
	const search = Route.useSearch();
	const navigate = Route.useNavigate();
	const [code, setCode] = useState(search.code?.toUpperCase() ?? "");
	const [runnerName, setRunnerName] = useState("");
	const [tagsString, setTagsString] = useState("");
	const [authInfo, setAuthInfo] = useState<RunnerDeviceAuthInfo | null>(null);
	const [approvedName, setApprovedName] = useState<string | null>(null);
	const [approvedTags, setApprovedTags] = useState<string[]>([]);
	const [verifying, setVerifying] = useState(false);
	const [submitting, setSubmitting] = useState(false);
	const [status, setStatus] = useState<
		"idle" | "verified" | "approved" | "denied"
	>("idle");
	const [error, setError] = useState<string | null>(null);

	const handleVerifyCode = useCallback(
		async (userCode: string) => {
			const cleaned = userCode.trim().toUpperCase();
			if (!cleaned) return;

			setVerifying(true);
			setError(null);
			try {
				const info = await verifyRunnerDeviceCode(account.slug, cleaned);
				setAuthInfo(info);
				setRunnerName(info.runner_name || "Self-Hosted Runner");
				setTagsString(
					info.tags && info.tags.length > 0 ? info.tags.join(", ") : "default",
				);
				setStatus("verified");
				void navigate({ search: {}, replace: true });
			} catch (err) {
				setError(err instanceof ApiError ? err.message : translateError(err));
				setStatus("idle");
			} finally {
				setVerifying(false);
			}
		},
		[navigate, account.slug],
	);

	// NOTE: `search.code` only pre-fills the input — it is never verified
	// automatically (phishing: a linkable code could bind an attacker's
	// runner to this account, RFC 8628 §8.4).

	async function handleApprove(e: FormEvent) {
		e.preventDefault();
		if (!authInfo) return;

		setSubmitting(true);
		setError(null);
		try {
			const parsedTags = tagsString
				.split(",")
				.map((t) => t.trim())
				.filter(Boolean);

			const res = await approveRunnerDeviceAuth(account.slug, {
				user_code: authInfo.user_code,
				runner_name: runnerName.trim() || undefined,
				tags: parsedTags,
			});
			setApprovedName(res.runner.name);
			setApprovedTags(res.runner.tags);
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
			await denyRunnerDeviceAuth(account.slug, authInfo.user_code);
			setStatus("denied");
		} catch (err) {
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
						label: m.nav_runners(),
						to: "/$account_slug/runners",
						params: { account_slug: account.slug },
					},
					{ label: "Authorize Runner", isCurrentPage: true },
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
								<CardTitle className="text-xl">Runner Connected!</CardTitle>
								<CardDescription>
									Runner{" "}
									<span className="font-semibold text-foreground">
										{approvedName}
									</span>{" "}
									has been registered to your account.
								</CardDescription>
							</CardHeader>
							<CardContent className="space-y-4 pt-2">
								{approvedTags.length > 0 && (
									<div className="flex flex-wrap justify-center gap-1.5">
										{approvedTags.map((tag) => (
											<Badge key={tag} variant="secondary" className="text-xs">
												{tag}
											</Badge>
										))}
									</div>
								)}
								<p className="text-xs text-muted-foreground">
									You can close this tab and return to your terminal. Run{" "}
									<code className="rounded bg-muted px-1.5 py-0.5 font-mono text-xs">
										beep runner up
									</code>{" "}
									or{" "}
									<code className="rounded bg-muted px-1.5 py-0.5 font-mono text-xs">
										beep up
									</code>{" "}
									to start executing scheduled tasks.
								</p>
								<div className="flex justify-center gap-2">
									<Link
										to="/$account_slug/runners"
										params={{ account_slug: account.slug }}
										className={buttonVariants({
											variant: "outline",
											size: "sm",
										})}
									>
										View Runners
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
									The runner device authorization request has been rejected.
								</CardDescription>
							</CardHeader>
							<CardContent className="space-y-4 pt-2">
								<p className="text-xs text-muted-foreground">
									You can close this tab. To connect later, run{" "}
									<code className="rounded bg-muted px-1.5 py-0.5 font-mono text-xs">
										beep runner connect
									</code>{" "}
									in your terminal again.
								</p>
							</CardContent>
						</Card>
					) : (
						<Card>
							<CardHeader className="text-center">
								<div className="mx-auto mb-2 flex size-12 items-center justify-center rounded-xl bg-primary/10 text-primary">
									<Server className="size-6" />
								</div>
								<CardTitle className="text-xl">
									Register Self-Hosted Runner
								</CardTitle>
								<CardDescription>
									Authorize your machine as a task runner for this account.
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
										<div className="rounded-lg border bg-muted/40 p-3 space-y-1.5 text-xs">
											<div className="flex justify-between text-muted-foreground">
												<span>Code</span>
												<span className="font-mono font-semibold text-foreground">
													{authInfo.user_code}
												</span>
											</div>
											{authInfo.metadata?.hostname && (
												<div className="flex justify-between text-muted-foreground">
													<span>Host</span>
													<span className="font-mono text-foreground">
														{authInfo.metadata.hostname}
													</span>
												</div>
											)}
											{(authInfo.metadata?.os || authInfo.metadata?.arch) && (
												<div className="flex justify-between text-muted-foreground">
													<span>System</span>
													<span className="text-foreground">
														{authInfo.metadata?.os}/{authInfo.metadata?.arch}
													</span>
												</div>
											)}
										</div>

										<div className="space-y-2">
											<Label htmlFor="runner-name">Runner Name</Label>
											<Input
												id="runner-name"
												type="text"
												value={runnerName}
												onChange={(e) => setRunnerName(e.target.value)}
												placeholder="e.g. Build Machine 01"
												maxLength={80}
												required
											/>
											<p className="text-xs text-muted-foreground">
												A descriptive name to identify this runner instance.
											</p>
										</div>

										<div className="space-y-2">
											<Label htmlFor="runner-tags">
												Tags (Comma-separated)
											</Label>
											<Input
												id="runner-tags"
												type="text"
												value={tagsString}
												onChange={(e) => setTagsString(e.target.value)}
												placeholder="e.g. local, macos, default"
											/>
											<p className="text-xs text-muted-foreground">
												Jobs targeting these tags will be dispatched to this
												runner.
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
												disabled={submitting || !runnerName.trim()}
											>
												{submitting ? "Approving..." : "Approve Runner"}
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
