import {
	createFileRoute,
	getRouteApi,
	Link,
	notFound,
	useRouter,
} from "@tanstack/react-router";
import { Activity, Pause, Play, Trash2 } from "lucide-react";
import { useState } from "react";
import { BeepMarkdown } from "@/components/beeps/beep-markdown";
import { BeepRuns } from "@/components/beeps/beep-runs";
import { BEEP_STATUS_META } from "@/components/beeps/beep-status";
import { confirm } from "@/components/confirm-dialog";
import { DashboardHeader } from "@/components/dashboard/dashboard-header";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import {
	deleteBeep,
	fetchBeep,
	fetchBeepRuns,
	pauseBeep,
	resumeBeep,
	triggerBeepRun,
} from "@/lib/api/beeps";
import { ApiError } from "@/lib/api/client";
import { withAuthRedirects } from "@/lib/auth/guards";
import { formatBeepScheduleTime } from "@/lib/beep-datetime";
import { beepStatusLabel, channelLabel } from "@/lib/i18n-labels";
import type { NotificationChannel } from "@/lib/notification-channels";
import { m } from "@/locale/paraglide/messages";

const accountRoute = getRouteApi("/$account_slug");

export const Route = createFileRoute("/$account_slug/beeps_/$beepId")({
	loader: withAuthRedirects(async ({ params }) => {
		const slug = params?.account_slug ?? "";
		const beepId = params?.beepId ?? "";
		try {
			const [beep, runsRes] = await Promise.all([
				fetchBeep(slug, beepId),
				fetchBeepRuns(slug, beepId),
			]);
			return {
				beep,
				runs: runsRes.runs,
				runsPagination: runsRes.pagination,
			};
		} catch (err) {
			if (err instanceof ApiError && err.status === 404) {
				throw notFound();
			}
			throw err;
		}
	}),
	component: BeepDetailPage,
});

function BeepDetailPage() {
	const { account_slug: slug } = accountRoute.useParams();
	const router = useRouter();
	const { beep, runs, runsPagination } = Route.useLoaderData();
	const [deleting, setDeleting] = useState(false);
	const [triggering, setTriggering] = useState(false);
	const [togglingStatus, setTogglingStatus] = useState(false);
	const [error, setError] = useState<string | null>(null);

	const isOnce = beep.kind === "once";

	async function handleTrigger() {
		if (
			isOnce &&
			beep.status !== "completed" &&
			beep.run_at &&
			Number(new Date(beep.run_at)) > Date.now() &&
			!(await confirm({
				description: m.beeps_send_now_confirm_future(),
				variant: "default",
			}))
		) {
			return;
		}
		setTriggering(true);
		setError(null);
		try {
			await triggerBeepRun(slug, beep.id);
			await router.invalidate();
		} catch (err) {
			setError(
				err instanceof ApiError ? err.message : m.beeps_trigger_failed(),
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
				err instanceof ApiError ? err.message : m.beeps_status_update_failed(),
			);
		} finally {
			setTogglingStatus(false);
		}
	}

	async function handleDelete() {
		if (!(await confirm(m.beeps_delete_confirm({ title: beep.title })))) {
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
			setError(err instanceof ApiError ? err.message : m.beeps_delete_failed());
		}
	}

	const kindLabel =
		beep.kind === "recurring" ? m.beeps_kind_recurring() : m.beeps_kind_once();

	return (
		<>
			<DashboardHeader
				breadcrumbs={[
					{
						label: m.nav_home(),
						to: "/$account_slug",
						params: { account_slug: slug },
					},
					{
						label: m.nav_beeps(),
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
							{beep.beeper ? (
								beep.beeper_id ? (
									<Link
										to="/$account_slug/beepers/$beeperId"
										params={{
											account_slug: slug,
											beeperId: beep.beeper_id,
										}}
									>
										<Badge variant="outline" className="gap-1 cursor-pointer">
											<Activity className="size-3 text-muted-foreground" />
											{beep.beeper.name}
										</Badge>
									</Link>
								) : (
									<Badge variant="outline" className="gap-1">
										<Activity className="size-3 text-muted-foreground" />
										{beep.beeper.name}
									</Badge>
								)
							) : null}
						</div>
						<p className="mt-1 text-sm text-muted-foreground">
							{kindLabel} · {beep.timezone}
						</p>
					</div>
					<div className="flex flex-wrap items-center gap-2">
						{(() => {
							const statusMeta = BEEP_STATUS_META[beep.status];
							const StatusIcon = statusMeta.icon;
							return (
								<Badge variant={statusMeta.badgeVariant} className="gap-1">
									<StatusIcon className="size-3" />
									{beepStatusLabel(beep.status)}
								</Badge>
							);
						})()}
						{beep.status === "active" || beep.status === "paused" ? (
							<Button
								variant="outline"
								size="sm"
								disabled={togglingStatus || deleting || triggering}
								aria-label={
									beep.status === "paused" ? m.beeps_resume() : m.beeps_pause()
								}
								onClick={() => void handleToggleStatus()}
							>
								{beep.status === "paused" ? (
									<>
										<Play data-icon="inline-start" />
										{togglingStatus ? m.beeps_resuming() : m.beeps_resume()}
									</>
								) : (
									<>
										<Pause data-icon="inline-start" />
										{togglingStatus ? m.beeps_pausing() : m.beeps_pause()}
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
							aria-label={
								isOnce ? m.beeps_send_beep_now() : m.beeps_trigger_run()
							}
							onClick={() => void handleTrigger()}
						>
							<Play data-icon="inline-start" />
							{triggering
								? isOnce
									? m.beeps_sending()
									: m.beeps_triggering()
								: beep.status === "firing"
									? m.beeps_firing_action()
									: isOnce
										? m.beeps_send_beep_now()
										: m.beeps_trigger_run()}
						</Button>
						<Button
							variant="destructive"
							size="sm"
							disabled={deleting || triggering || togglingStatus}
							aria-label={m.common_delete()}
							onClick={() => void handleDelete()}
						>
							<Trash2 data-icon="inline-start" />
							{deleting ? m.beeps_deleting() : m.common_delete()}
						</Button>
					</div>
				</div>

				{error ? (
					<p className="text-sm text-destructive" role="alert">
						{error}
					</p>
				) : null}

				{beep.body ? (
					<Card className="max-w-lg">
						<CardHeader>
							<CardTitle>{m.beeps_body()}</CardTitle>
						</CardHeader>
						<CardContent>
							<BeepMarkdown source={beep.body} />
						</CardContent>
					</Card>
				) : null}

				<Card className="max-w-lg">
					<CardHeader>
						<CardTitle>{m.beeps_details_channels()}</CardTitle>
					</CardHeader>
					<CardContent className="flex flex-col gap-3 text-sm">
						{beep.kind === "once" ? (
							<DetailRow
								label={m.beeps_run_at()}
								value={formatBeepScheduleTime(beep.run_at, beep.timezone)}
							/>
						) : (
							<DetailRow
								label={m.beeps_cron()}
								value={beep.cron ?? m.common_em_dash()}
							/>
						)}
						<DetailRow
							label={m.beeps_next()}
							value={formatBeepScheduleTime(beep.next_run_at, beep.timezone)}
						/>
						<DetailRow
							label={m.beeps_last()}
							value={formatBeepScheduleTime(beep.last_run_at, beep.timezone)}
						/>
						<DetailRow
							label={m.beeps_channels()}
							value={
								beep.notification_channels?.length > 0
									? beep.notification_channels
											.map((channel) =>
												channelLabel(channel as NotificationChannel),
											)
											.join(", ")
									: m.beeps_channels_none()
							}
						/>
						<DetailRow
							label={m.common_created()}
							value={formatBeepScheduleTime(beep.created_at, beep.timezone)}
						/>
						{beep.intent ? (
							<DetailRow label={m.beeps_intent()} value={beep.intent} />
						) : null}
						{beep.metadata && Object.keys(beep.metadata).length > 0 ? (
							<DetailRow
								label={m.beeps_metadata()}
								value={JSON.stringify(beep.metadata)}
							/>
						) : null}
					</CardContent>
				</Card>

				<div className="max-w-lg">
					<BeepRuns
						runs={runs}
						initialPagination={runsPagination}
						slug={slug}
						beepId={beep.id}
						timezone={beep.timezone}
					/>
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
