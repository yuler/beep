import {
	createFileRoute,
	getRouteApi,
	notFound,
	useRouter,
} from "@tanstack/react-router";
import { ChevronRight, Pause, Play, Trash2 } from "lucide-react";
import { useState } from "react";

import { DashboardHeader } from "@/components/dashboard/dashboard-header";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
	type BeeperRun,
	deleteBeeperInstall,
	fetchBeeperInstall,
	pauseBeeperInstall,
	resumeBeeperInstall,
} from "@/lib/api/beepers";
import { ApiError } from "@/lib/api/client";
import { withAuthRedirects } from "@/lib/auth/guards";
import { cn } from "@/lib/utils";

const accountRoute = getRouteApi("/$account_slug");

export const Route = createFileRoute("/$account_slug/beepers_/$beeperId")({
	loader: withAuthRedirects(async ({ params }) => {
		try {
			return await fetchBeeperInstall(
				params?.account_slug ?? "",
				params?.beeperId ?? "",
			);
		} catch (err) {
			if (err instanceof ApiError && err.status === 404) {
				throw notFound();
			}
			throw err;
		}
	}),
	component: BeeperInstallDetailPage,
});

function formatWhen(value: string | null | undefined) {
	if (!value) return "—";
	return new Date(value).toLocaleString();
}

const RUN_STATUS_VARIANT: Record<
	string,
	"default" | "secondary" | "outline" | "destructive"
> = {
	pending: "secondary",
	running: "default",
	succeeded: "outline",
	failed: "destructive",
	skipped: "secondary",
	expired: "destructive",
};

const CHECK_STATUS_VARIANT: Record<
	string,
	"default" | "secondary" | "outline" | "destructive"
> = {
	ok: "outline",
	alerting: "destructive",
	error: "destructive",
};

function BeeperInstallDetailPage() {
	const { account_slug: slug } = accountRoute.useParams();
	const router = useRouter();
	const install = Route.useLoaderData();
	const [deleting, setDeleting] = useState(false);
	const [togglingStatus, setTogglingStatus] = useState(false);
	const [error, setError] = useState<string | null>(null);

	async function handleToggleStatus() {
		setTogglingStatus(true);
		setError(null);
		try {
			if (install.status === "paused") {
				await resumeBeeperInstall(slug, install.id);
			} else {
				await pauseBeeperInstall(slug, install.id);
			}
			await router.invalidate();
		} catch (err) {
			setError(
				err instanceof ApiError
					? err.message
					: "Failed to update beeper status.",
			);
		} finally {
			setTogglingStatus(false);
		}
	}

	async function handleDelete() {
		if (!window.confirm(`Delete "${install.title}"? This cannot be undone.`)) {
			return;
		}

		setDeleting(true);
		setError(null);
		try {
			await deleteBeeperInstall(slug, install.id);
			await router.navigate({
				to: "/$account_slug/beepers",
				params: { account_slug: slug },
			});
			await router.invalidate();
		} catch (err) {
			setDeleting(false);
			setError(
				err instanceof ApiError
					? err.message
					: "Failed to delete beeper install.",
			);
		}
	}

	const runs = install.runs ?? [];

	return (
		<>
			<DashboardHeader
				breadcrumbs={[
					{
						label: "Home",
						to: "/$account_slug",
						params: { account_slug: slug },
					},
					{
						label: "Beepers",
						to: "/$account_slug/beepers",
						params: { account_slug: slug },
					},
					{ label: install.title, isCurrentPage: true },
				]}
			/>

			<div className="flex flex-1 flex-col gap-6 p-4 md:p-6">
				<div className="flex flex-wrap items-start justify-between gap-3">
					<div className="min-w-0">
						<div className="flex items-center gap-2">
							<h1 className="font-heading text-2xl font-semibold tracking-tight">
								{install.title}
							</h1>
							<Badge
								variant={
									install.alert_state === "alerting" ? "destructive" : "outline"
								}
							>
								{install.alert_state}
							</Badge>
						</div>
						<p className="mt-1 text-sm text-muted-foreground">
							{install.beeper?.name} (v{install.beeper?.version}) ·{" "}
							{install.timezone}
						</p>
					</div>
					<div className="flex flex-wrap items-center gap-2">
						<Badge
							variant={
								install.status === "active"
									? "default"
									: install.status === "paused"
										? "secondary"
										: "outline"
							}
						>
							{install.status}
						</Badge>
						{install.status === "active" || install.status === "paused" ? (
							<Button
								variant="outline"
								size="sm"
								disabled={togglingStatus || deleting}
								aria-label={`${install.status === "paused" ? "Resume" : "Pause"} beeper ${install.title}`}
								onClick={() => void handleToggleStatus()}
							>
								{install.status === "paused" ? (
									<>
										<Play data-icon="inline-start" />
										{togglingStatus ? "Resuming…" : "Resume"}
									</>
								) : (
									<>
										<Pause data-icon="inline-start" />
										{togglingStatus ? "Pausing…" : "Pause"}
									</>
								)}
							</Button>
						) : null}
						<Button
							variant="destructive"
							size="sm"
							disabled={deleting || togglingStatus}
							aria-label={`Delete beeper ${install.title}`}
							onClick={() => void handleDelete()}
						>
							<Trash2 data-icon="inline-start" />
							{deleting ? "Deleting…" : "Delete"}
						</Button>
					</div>
				</div>

				{error ? (
					<p className="text-sm text-destructive" role="alert">
						{error}
					</p>
				) : null}

				<Card className="max-w-lg">
					<CardHeader>
						<CardTitle>Schedule & Details</CardTitle>
					</CardHeader>
					<CardContent className="flex flex-col gap-3 text-sm">
						<DetailRow label="Cron" value={install.cron} />
						<DetailRow
							label="Next run"
							value={formatWhen(install.next_run_at)}
						/>
						<DetailRow
							label="Last run"
							value={formatWhen(install.last_run_at)}
						/>
						<DetailRow
							label="Consecutive failures"
							value={String(install.consecutive_failures)}
						/>
						<DetailRow
							label="Channels"
							value={
								install.notification_channels?.length > 0
									? install.notification_channels.join(", ")
									: "Default"
							}
						/>
						<DetailRow label="Created" value={formatWhen(install.created_at)} />
					</CardContent>
				</Card>

				{install.config && Object.keys(install.config).length > 0 ? (
					<Card className="max-w-lg">
						<CardHeader>
							<CardTitle>Configuration</CardTitle>
						</CardHeader>
						<CardContent>
							<pre className="max-h-48 overflow-auto rounded bg-muted/50 p-2 text-xs font-mono leading-snug whitespace-pre-wrap">
								{JSON.stringify(install.config, null, 2)}
							</pre>
						</CardContent>
					</Card>
				) : null}

				<div className="max-w-lg">
					<details className="group/runs rounded-lg border bg-muted/20 text-sm">
						<summary className="flex cursor-pointer list-none items-center gap-2 px-3 py-2 marker:hidden [&::-webkit-details-marker]:hidden">
							<ChevronRight className="size-3.5 shrink-0 text-muted-foreground transition-transform group-open/runs:rotate-90" />
							<span className="font-medium">Execution Runs</span>
							<span className="text-muted-foreground">{runs.length}</span>
						</summary>
						{runs.length === 0 ? (
							<p className="border-t px-3 py-2 text-xs text-muted-foreground">
								No runs yet.
							</p>
						) : (
							<ul className="flex flex-col gap-2 border-t px-3 py-2">
								{runs.map((run: BeeperRun) => (
									<li
										key={run.id}
										className="flex flex-col gap-2 rounded-md bg-background p-2.5 ring-1 ring-foreground/10"
									>
										<div className="flex flex-wrap items-center justify-between gap-2">
											<span className="tabular-nums text-xs text-muted-foreground">
												{formatWhen(run.scheduled_for)}
											</span>
											<div className="flex items-center gap-1.5">
												{run.check_status ? (
													<Badge
														variant={
															CHECK_STATUS_VARIANT[run.check_status] ??
															"secondary"
														}
													>
														Check: {run.check_status.toUpperCase()}
													</Badge>
												) : null}
												<Badge
													variant={
														RUN_STATUS_VARIANT[run.status] ?? "secondary"
													}
												>
													{run.status}
												</Badge>
											</div>
										</div>

										{run.check_result ? (
											<div className="rounded bg-muted/40 p-2 text-xs">
												{run.check_result.title ? (
													<div className="font-medium text-foreground">
														{run.check_result.title}
													</div>
												) : null}
												{run.check_result.message ? (
													<div className="text-muted-foreground mt-0.5">
														{run.check_result.message}
													</div>
												) : null}
												{run.check_result.metrics ? (
													<div className="mt-1 flex flex-wrap gap-2 text-[11px] font-mono text-muted-foreground">
														{Object.entries(run.check_result.metrics).map(
															([k, v]) => (
																<span
																	key={k}
																	className="rounded bg-background px-1.5 py-0.5 border"
																>
																	{k}: {String(v)}
																</span>
															),
														)}
													</div>
												) : null}
											</div>
										) : null}
									</li>
								))}
							</ul>
						)}
					</details>
				</div>
			</div>
		</>
	);
}

function DetailRow({ label, value }: { label: string; value: string }) {
	return (
		<div className="flex justify-between gap-4">
			<span className="text-muted-foreground">{label}</span>
			<span className="text-right tabular-nums">{value}</span>
		</div>
	);
}
