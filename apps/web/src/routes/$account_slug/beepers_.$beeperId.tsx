import {
	createFileRoute,
	getRouteApi,
	notFound,
	useRouter,
} from "@tanstack/react-router";
import {
	Check,
	ChevronRight,
	Copy,
	Edit,
	Pause,
	Play,
	SlidersHorizontal,
	Trash2,
	Webhook,
} from "lucide-react";
import { useState } from "react";

import { BeepMarkdown } from "@/components/beeps/beep-markdown";
import { DashboardHeader } from "@/components/dashboard/dashboard-header";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { publicApiOrigin } from "@/config";
import {
	type BeeperRun,
	deleteBeeper,
	fetchBeeper,
	pauseBeeper,
	resumeBeeper,
	triggerBeeperRun,
	updateBeeper,
} from "@/lib/api/beepers";
import { ApiError } from "@/lib/api/client";
import { withAuthRedirects } from "@/lib/auth/guards";
import { formatBeepScheduleTime } from "@/lib/beep-datetime";
import {
	beeperHealthIsDestructive,
	beeperHealthLabel,
	isBeeperProbeBroken,
} from "@/lib/beeper-health";

const accountRoute = getRouteApi("/$account_slug");

export const Route = createFileRoute("/$account_slug/beepers_/$beeperId")({
	loader: withAuthRedirects(async ({ params }) => {
		try {
			return await fetchBeeper(
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
	component: BeeperDetailPage,
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

const SIGNAL_STATUS_VARIANT: Record<
	string,
	"default" | "secondary" | "outline" | "destructive"
> = {
	ok: "outline",
	alerting: "destructive",
	error: "destructive",
};

function BeeperDetailPage() {
	const { account_slug: slug } = accountRoute.useParams();
	const router = useRouter();
	const beeper = Route.useLoaderData();
	const [deleting, setDeleting] = useState(false);
	const [triggering, setTriggering] = useState(false);
	const [togglingStatus, setTogglingStatus] = useState(false);
	const [hasCopiedPing, setHasCopiedPing] = useState(false);
	const [error, setError] = useState<string | null>(null);

	// Edit modal state
	const [isEditDialogOpen, setIsEditDialogOpen] = useState(false);
	const [editTitle, setEditTitle] = useState(beeper.title);
	const [editBody, setEditBody] = useState(beeper.body ?? "");
	const [editCron, setEditCron] = useState(beeper.cron);
	const [editInputs, setEditInputs] = useState<Record<string, unknown>>(
		beeper.config ?? {},
	);
	const [savingEdit, setSavingEdit] = useState(false);
	const [editError, setEditError] = useState<string | null>(null);

	const pingUrl = beeper.ping_token
		? `${publicApiOrigin()}/api/v1/beeper_apps/heartbeat/pings/${beeper.ping_token}`
		: null;

	function handleOpenEdit() {
		setEditTitle(beeper.title);
		setEditBody(beeper.body ?? "");
		setEditCron(beeper.cron);
		setEditInputs({ ...(beeper.config ?? {}) });
		setEditError(null);
		setIsEditDialogOpen(true);
	}

	async function handleSaveEdit(e: React.FormEvent) {
		e.preventDefault();
		setSavingEdit(true);
		setEditError(null);
		try {
			await updateBeeper(slug, beeper.id, {
				title: editTitle.trim(),
				body: editBody.trim() || null,
				cron: editCron.trim(),
				config: editInputs,
			});
			setIsEditDialogOpen(false);
			await router.invalidate();
		} catch (err) {
			setEditError(
				err instanceof ApiError ? err.message : "Failed to update beeper.",
			);
		} finally {
			setSavingEdit(false);
		}
	}

	function handleCopyPing() {
		if (!pingUrl) return;
		void navigator.clipboard.writeText(pingUrl);
		setHasCopiedPing(true);
		setTimeout(() => setHasCopiedPing(false), 2000);
	}

	async function handleTrigger() {
		setTriggering(true);
		setError(null);
		try {
			await triggerBeeperRun(slug, beeper.id);
			await router.invalidate();
		} catch (err) {
			setError(
				err instanceof ApiError ? err.message : "Failed to trigger beeper run.",
			);
		} finally {
			setTriggering(false);
		}
	}

	async function handleToggleStatus() {
		setTogglingStatus(true);
		setError(null);
		try {
			if (beeper.status === "paused") {
				await resumeBeeper(slug, beeper.id);
			} else {
				await pauseBeeper(slug, beeper.id);
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
		if (!window.confirm(`Delete "${beeper.title}"? This cannot be undone.`)) {
			return;
		}

		setDeleting(true);
		setError(null);
		try {
			await deleteBeeper(slug, beeper.id);
			await router.navigate({
				to: "/$account_slug/beepers",
				params: { account_slug: slug },
			});
			await router.invalidate();
		} catch (err) {
			setDeleting(false);
			setError(
				err instanceof ApiError ? err.message : "Failed to delete beeper.",
			);
		}
	}

	const runs = beeper.runs ?? [];
	const inputs = beeper.beeper_app?.inputs ?? [];
	const configEntries = Object.entries(beeper.config ?? {});

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
					{ label: beeper.title, isCurrentPage: true },
				]}
			/>

			<div className="flex flex-1 flex-col gap-6 p-4 md:p-6">
				<div className="flex flex-wrap items-start justify-between gap-3">
					<div className="min-w-0">
						<div className="flex items-center gap-2">
							<h1 className="font-heading text-2xl font-semibold tracking-tight">
								{beeper.title}
							</h1>
							<Badge
								variant={
									beeperHealthIsDestructive(beeper) ? "destructive" : "outline"
								}
							>
								{beeperHealthLabel(beeper)}
							</Badge>
							{isBeeperProbeBroken(beeper.runs) ? (
								<p className="text-sm text-muted-foreground">
									Probe is failing to run — check configuration or network
									access, not target health.
								</p>
							) : null}
						</div>
						<p className="mt-1 text-sm text-muted-foreground">
							{beeper.beeper_app?.name} (v{beeper.beeper_app?.version}) ·{" "}
							{beeper.timezone}
						</p>
					</div>
					<div className="flex flex-wrap items-center gap-2">
						<Badge
							variant={
								beeper.status === "active"
									? "default"
									: beeper.status === "paused"
										? "secondary"
										: "outline"
							}
						>
							{beeper.status}
						</Badge>
						<Button
							variant="outline"
							size="sm"
							disabled={deleting || triggering || togglingStatus}
							aria-label={`Edit beeper ${beeper.title}`}
							onClick={handleOpenEdit}
						>
							<Edit data-icon="inline-start" />
							Edit
						</Button>
						{beeper.status === "active" || beeper.status === "paused" ? (
							<Button
								variant="outline"
								size="sm"
								disabled={togglingStatus || deleting || triggering}
								aria-label={`${beeper.status === "paused" ? "Resume" : "Pause"} beeper ${beeper.title}`}
								onClick={() => void handleToggleStatus()}
							>
								{beeper.status === "paused" ? (
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
								beeper.status === "firing"
							}
							aria-label={`Trigger run for beeper ${beeper.title}`}
							onClick={() => void handleTrigger()}
						>
							<Play data-icon="inline-start" />
							{triggering
								? "Triggering…"
								: beeper.status === "firing"
									? "Firing…"
									: "Trigger run"}
						</Button>
						<Button
							variant="destructive"
							size="sm"
							disabled={deleting || togglingStatus || triggering}
							aria-label={`Delete beeper ${beeper.title}`}
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

				{pingUrl ? (
					<Card className="border-primary/25 bg-primary/5">
						<CardHeader className="pb-3">
							<div className="flex flex-wrap items-center justify-between gap-3">
								<div className="flex items-center gap-2">
									<Webhook className="size-4 text-primary" />
									<CardTitle className="text-base">Webhook Ping URL</CardTitle>
								</div>
								<Button
									variant="outline"
									size="sm"
									onClick={handleCopyPing}
									className="gap-1.5 font-normal"
								>
									{hasCopiedPing ? (
										<>
											<Check className="size-3.5 text-green-500" />
											<span>Copied</span>
										</>
									) : (
										<>
											<Copy className="size-3.5" />
											<span>Copy URL</span>
										</>
									)}
								</Button>
							</div>
							<CardDescription className="text-xs">
								Send an HTTP POST request to this endpoint after each job run or
								backup.
							</CardDescription>
						</CardHeader>
						<CardContent className="pt-0">
							<div className="rounded-md bg-muted/60 px-3 py-2 font-mono text-xs text-foreground select-all break-all">
								{pingUrl}
							</div>
						</CardContent>
					</Card>
				) : null}

				{beeper.body ? (
					<Card>
						<CardHeader className="pb-2">
							<CardTitle className="text-base">Body / Remarks</CardTitle>
						</CardHeader>
						<CardContent>
							<BeepMarkdown source={beeper.body} />
						</CardContent>
					</Card>
				) : null}

				<div className="grid gap-6 md:grid-cols-2">
					<Card>
						<CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
							<CardTitle>Schedule & Status</CardTitle>
							<Button
								variant="ghost"
								size="sm"
								className="h-8 px-2 text-xs text-muted-foreground hover:text-foreground"
								onClick={handleOpenEdit}
							>
								<Edit className="size-3.5" />
								Edit
							</Button>
						</CardHeader>
						<CardContent className="flex flex-col gap-3 text-sm">
							<DetailRow label="Cron" value={beeper.cron} />
							<DetailRow
								label="Next run"
								value={formatBeepScheduleTime(
									beeper.next_run_at,
									beeper.timezone,
								)}
							/>
							<DetailRow
								label="Last run"
								value={formatBeepScheduleTime(
									beeper.last_run_at,
									beeper.timezone,
								)}
							/>
							<DetailRow
								label="Consecutive failures"
								value={String(beeper.consecutive_failures)}
							/>
							{beeper.ping_token ? (
								<DetailRow label="Ping token" value={beeper.ping_token} />
							) : null}
							{beeper.last_ping_at ? (
								<DetailRow
									label="Last ping"
									value={formatWhen(beeper.last_ping_at)}
								/>
							) : null}
							<DetailRow
								label="Channels"
								value={
									beeper.notification_channels?.length > 0
										? beeper.notification_channels.join(", ")
										: "Default"
								}
							/>
							<DetailRow
								label="Created"
								value={formatWhen(beeper.created_at)}
							/>
						</CardContent>
					</Card>

					<Card>
						<CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
							<CardTitle>Configuration</CardTitle>
							<Button
								variant="ghost"
								size="sm"
								className="h-8 px-2 text-xs text-muted-foreground hover:text-foreground"
								onClick={handleOpenEdit}
							>
								<SlidersHorizontal className="size-3.5" />
								Edit Config
							</Button>
						</CardHeader>
						<CardContent className="flex flex-col gap-3 text-sm">
							{configEntries.length === 0 ? (
								<p className="text-sm text-muted-foreground">
									No configuration parameters set.
								</p>
							) : (
								configEntries.map(([key, val]) => {
									const inputDef = inputs.find((i) => i.name === key);
									const label = inputDef?.label || key;
									const displayVal =
										val === null || val === undefined
											? "—"
											: typeof val === "boolean"
												? val
													? "true"
													: "false"
												: String(val);

									return (
										<div
											key={key}
											className="flex justify-between gap-4 border-b border-border/40 pb-2 last:border-0 last:pb-0"
										>
											<div className="flex flex-col min-w-0">
												<span className="font-medium text-foreground">
													{label}
												</span>
												<span className="font-mono text-[11px] text-muted-foreground">
													{key}
												</span>
											</div>
											<span className="font-mono text-sm text-right break-all max-w-[60%]">
												{displayVal}
											</span>
										</div>
									);
								})
							)}
						</CardContent>
					</Card>
				</div>

				{/* Edit Beeper Dialog */}
				<Dialog
					open={isEditDialogOpen}
					onOpenChange={(open) => {
						if (!open && !savingEdit) {
							setIsEditDialogOpen(false);
						}
					}}
				>
					<DialogContent className="sm:max-w-lg">
						<form onSubmit={handleSaveEdit} className="flex flex-col gap-4">
							<DialogHeader>
								<DialogTitle className="text-lg">Edit Beeper</DialogTitle>
								<DialogDescription>
									Update the title, remarks, schedule, or configuration
									parameters for this beeper.
								</DialogDescription>
							</DialogHeader>

							<div className="flex flex-col gap-4 py-2">
								<div className="flex flex-col gap-2">
									<Label htmlFor="edit-beeper-title">Beeper Title</Label>
									<Input
										id="edit-beeper-title"
										required
										value={editTitle}
										onChange={(e) => setEditTitle(e.target.value)}
										disabled={savingEdit}
									/>
								</div>

								<div className="flex flex-col gap-2">
									<div className="flex items-center justify-between">
										<Label htmlFor="edit-beeper-body">Body / Remark</Label>
										<span className="text-[11px] text-muted-foreground">
											Optional
										</span>
									</div>
									<textarea
										id="edit-beeper-body"
										rows={3}
										placeholder="Add notes, runbook links, or alert context..."
										value={editBody}
										onChange={(e) => setEditBody(e.target.value)}
										disabled={savingEdit}
										className="w-full rounded-lg border border-input bg-transparent px-2.5 py-1.5 text-sm transition-colors outline-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:pointer-events-none disabled:cursor-not-allowed disabled:bg-input/50 disabled:opacity-50 dark:bg-input/30"
									/>
								</div>

								<div className="flex flex-col gap-2">
									<Label htmlFor="edit-beeper-cron">Cron Schedule</Label>
									<Input
										id="edit-beeper-cron"
										required
										value={editCron}
										onChange={(e) => setEditCron(e.target.value)}
										disabled={savingEdit}
										className="font-mono text-sm"
									/>
								</div>

								{inputs.length > 0 ? (
									<div className="flex flex-col gap-3 pt-2">
										<h3 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
											Configuration Parameters
										</h3>
										{inputs.map((input) => (
											<div key={input.name} className="flex flex-col gap-2">
												<Label htmlFor={`edit-input-${input.name}`}>
													{input.label}
													{input.required ? (
														<span className="text-destructive ml-1">*</span>
													) : null}
												</Label>
												<Input
													id={`edit-input-${input.name}`}
													type={input.type === "number" ? "number" : "text"}
													required={input.required}
													min={input.min}
													max={input.max}
													placeholder={input.placeholder}
													value={String(editInputs[input.name] ?? "")}
													onChange={(e) => {
														const val =
															input.type === "number"
																? Number(e.target.value)
																: e.target.value;
														setEditInputs((curr) => ({
															...curr,
															[input.name]: val,
														}));
													}}
													disabled={savingEdit}
												/>
											</div>
										))}
									</div>
								) : null}

								{editError ? (
									<p className="text-sm text-destructive" role="alert">
										{editError}
									</p>
								) : null}
							</div>

							<DialogFooter className="gap-2 sm:gap-0">
								<Button
									type="button"
									variant="ghost"
									size="sm"
									onClick={() => setIsEditDialogOpen(false)}
									disabled={savingEdit}
								>
									Cancel
								</Button>
								<Button
									type="submit"
									size="sm"
									disabled={savingEdit || !editTitle.trim() || !editCron.trim()}
								>
									{savingEdit ? "Saving…" : "Save Changes"}
								</Button>
							</DialogFooter>
						</form>
					</DialogContent>
				</Dialog>

				<div>
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
												{formatBeepScheduleTime(
													run.scheduled_for,
													beeper.timezone,
												)}
											</span>
											<div className="flex items-center gap-1.5">
												{run.signal_status ? (
													<Badge
														variant={
															SIGNAL_STATUS_VARIANT[run.signal_status] ??
															"secondary"
														}
													>
														Signal: {run.signal_status.toUpperCase()}
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

										{run.signal_result ? (
											<div className="rounded bg-muted/40 p-2 text-xs">
												{run.signal_result.title ? (
													<div className="font-medium text-foreground">
														{run.signal_result.title}
													</div>
												) : null}
												{run.signal_result.message ? (
													<div className="text-muted-foreground mt-0.5">
														{run.signal_result.message}
													</div>
												) : null}
												{run.signal_result.metrics ? (
													<div className="mt-1 flex flex-wrap gap-2 text-[11px] font-mono text-muted-foreground">
														{Object.entries(run.signal_result.metrics).map(
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
