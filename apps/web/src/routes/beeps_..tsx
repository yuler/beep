import {
	createFileRoute,
	getRouteApi,
	notFound,
	useRouter,
} from "@tanstack/react-router";
import { Activity, Copy, Check, Pause, Play, Trash2 } from "lucide-react";
import { useState } from "react";

import { BeepMarkdown } from "@/components/beeps/beep-markdown";
import { BeepRuns } from "@/components/beeps/beep-runs";
import { DashboardHeader } from "@/components/dashboard/dashboard-header";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
	deleteBeep,
	fetchBeep,
	pauseBeep,
	resumeBeep,
	triggerBeepRun,
} from "@/lib/api/beeps";
import { ApiError } from "@/lib/api/client";
import { withAuthRedirects } from "@/lib/auth/guards";

const accountRoute = getRouteApi("/$account_slug");

export const Route = createFileRoute("/beeps_/")({
	loader: withAuthRedirects(async ({ params }) => {
		try {
			return await fetchBeep(params?.account_slug ?? "", params?.beepId ?? "");
		} catch (err) {
			if (err instanceof ApiError && err.status === 404) {
				throw notFound();
			}
			throw err;
		}
	}),
	component: BeepDetailPage,
});

const STATUS_VARIANT: Record<
	string,
	"default" | "secondary" | "outline" | "destructive"
> = {
	active: "default",
	paused: "secondary",
	completed: "outline",
	cancelled: "destructive",
	firing: "default",
};

const ALERT_STATE_VARIANT: Record<
	string,
	"default" | "secondary" | "outline" | "destructive"
> = {
	ok: "outline",
	alerting: "destructive",
};

function formatWhen(value: string | null | undefined) {
	if (!value) return "—";
	return new Date(value).toLocaleString();
}

function BeepDetailPage() {
	const { account_slug: slug } = accountRoute.useParams();
	const router = useRouter();
	const beep = Route.useLoaderData();
	const [deleting, setDeleting] = useState(false);
	const [triggering, setTriggering] = useState(false);
	const [togglingStatus, setTogglingStatus] = useState(false);
	const [copiedToken, setCopiedToken] = useState(false);
	const [error, setError] = useState<string | null>(null);

	async function handleTrigger() {
		setTriggering(true);
		setError(null);
		try {
			await triggerBeepRun(slug, beep.id);
			await router.invalidate();
		} catch (err) {
			setError(
				err instanceof ApiError ? err.message : "Failed to trigger beep run.",
			);
		} finally {
			setTriggering(false);
		}
	}

	async function handleToggleStatus() {
		setTogglingStatus(true);
		setError(null);
		try {
			if (beep.status === "paused") {
				await resumeBeep(slug, beep.id);
			} else {
				await pauseBeep(slug, beep.id);
			}
			await router.invalidate();
		} catch (err) {
			setError(
				err instanceof ApiError ? err.message : "Failed to update beep status.",
			);
		} finally {
			setTogglingStatus(false);
		}
	}

	async function handleDelete() {
		if (!window.confirm(`Delete "${beep.title}"? This cannot be undone.`)) {
			return;
		}

		setDeleting(true);
		setError(null);
		try {
			await deleteBeep(slug, beep.id);
			await router.navigate({
				to: "/$account_slug/beeps",
				params: { account_slug: slug },
			});
			await router.invalidate();
		} catch (err) {
			setDeleting(false);
			setError(
				err instanceof ApiError ? err.message : "Failed to delete beep.",
			);
		}
	}

	const pingWebhookUrl = beep.ping_token
		? `${window.location.origin}/api/v1/ping/${beep.ping_token}`
		: null;

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
						label: "Beeps",
						to: "/$account_slug/beeps",
						params: { account_slug: slug },
					},
					{ label: beep.title, isCurrentPage: true },
				]}
			/>

			<div className="flex flex-1 flex-col gap-6 p-4 md:p-6">
				<div className="flex flex-wrap items-start justify-between gap-3">
					<div className="min-w-0">
						<div className="flex items-center gap-2">
							<h1 className="font-heading text-2xl font-semibold tracking-tight">
								{beep.title}
							</h1>
							{beep.plugin ? (
								<Badge variant="secondary" className="font-normal text-xs">
									Plugin: {beep.plugin.name}
								</Badge>
							) : null}
						</div>
						<p className="mt-1 text-sm text-muted-foreground">
							{beep.kind} · {beep.timezone}
						</p>
					</div>
					<div className="flex flex-wrap items-center gap-2">
						{beep.alert_state ? (
							<Badge variant={ALERT_STATE_VARIANT[beep.alert_state] ?? "outline"}>
								Alert: {beep.alert_state.toUpperCase()}
							</Badge>
						) : null}
						<Badge variant={STATUS_VARIANT[beep.status] ?? "secondary"}>
							Status: {beep.status}
						</Badge>
						{beep.status === "active" || beep.status === "paused" ? (
							<Button
								variant="outline"
								size="sm"
								disabled={togglingStatus || deleting || triggering}
								aria-label={`${beep.status === "paused" ? "Resume" : "Pause"} beep ${beep.title}`}
								onClick={() => void handleToggleStatus()}
							>
								{beep.status === "paused" ? (
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
							variant="outline"
							size="sm"
							disabled={
								triggering ||
								deleting ||
								togglingStatus ||
								beep.status === "firing"
							}
							aria-label={`Trigger run for beep ${beep.title}`}
							onClick={() => void handleTrigger()}
						>
							<Activity data-icon="inline-start" />
							{triggering
								? "Checking…"
								: beep.status === "firing"
									? "Firing…"
									: beep.plugin ? "Test Run" : "Trigger run"}
						</Button>
						<Button
							variant="destructive"
							size="sm"
							disabled={deleting || triggering || togglingStatus}
							aria-label={`Delete beep ${beep.title}`}
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

				{beep.plugin ? (
					<Card className="max-w-lg">
						<CardHeader>
							<CardTitle>Plugin Configuration</CardTitle>
						</CardHeader>
						<CardContent className="flex flex-col gap-3 text-sm">
							<DetailRow label="Plugin" value={beep.plugin.name} />
							<DetailRow label="Version" value={`v${beep.plugin.version}`} />
							{beep.consecutive_failures !== undefined ? (
								<DetailRow
									label="Consecutive Failures"
									value={String(beep.consecutive_failures)}
								/>
							) : null}
							{pingWebhookUrl ? (
								<div className="flex flex-col gap-1.5 pt-1">
									<span className="text-xs text-muted-foreground font-medium">Ping URL (POST)</span>
									<div className="flex items-center gap-1.5 rounded-md border bg-muted/30 p-1.5 font-mono text-xs">
										<span className="truncate flex-1">{pingWebhookUrl}</span>
										<Button
											size="xs"
											variant="ghost"
											onClick={() => {
												navigator.clipboard.writeText(pingWebhookUrl);
												setCopiedToken(true);
												setTimeout(() => setCopiedToken(false), 2000);
											}}
										>
											{copiedToken ? <Check className="size-3" /> : <Copy className="size-3" />}
										</Button>
									</div>
								</div>
							) : null}
							{beep.plugin_config && Object.keys(beep.plugin_config).length > 0 ? (
								<div className="mt-2 flex flex-col gap-1">
									<span className="text-xs text-muted-foreground font-medium">Inputs</span>
									<pre className="overflow-auto rounded bg-muted/40 p-2 text-xs font-mono">
										{JSON.stringify(beep.plugin_config, null, 2)}
									</pre>
								</div>
							) : null}
						</CardContent>
					</Card>
				) : null}

				{beep.body ? (
					<Card className="max-w-lg">
						<CardHeader>
							<CardTitle>Body</CardTitle>
						</CardHeader>
						<CardContent>
							<BeepMarkdown source={beep.body} />
						</CardContent>
					</Card>
				) : null}

				<Card className="max-w-lg">
					<CardHeader>
						<CardTitle>Schedule</CardTitle>
					</CardHeader>
					<CardContent className="flex flex-col gap-3 text-sm">
						{beep.kind === "once" ? (
							<DetailRow label="Run at" value={formatWhen(beep.run_at)} />
						) : (
							<DetailRow label="Cron" value={beep.cron ?? "—"} />
						)}
						<DetailRow label="Next" value={formatWhen(beep.next_run_at)} />
						<DetailRow label="Last" value={formatWhen(beep.last_run_at)} />
						{beep.last_ping_at ? (
							<DetailRow label="Last Ping" value={formatWhen(beep.last_ping_at)} />
						) : null}
						<DetailRow label="Created" value={formatWhen(beep.created_at)} />
					</CardContent>
				</Card>

				<div className="max-w-lg">
					<BeepRuns runs={beep.runs} />
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
