import {
	createFileRoute,
	getRouteApi,
	useRouter,
} from "@tanstack/react-router";
import {
	Activity,
	ArrowRight,
	Clock,
	Globe,
	Radio,
	ShieldCheck,
	Webhook,
} from "lucide-react";
import { useState } from "react";
import { BeeperList } from "@/components/beepers/beeper-list";
import { RunnerRoutingPicker } from "@/components/beepers/runner-routing-picker";
import { DashboardHeader } from "@/components/dashboard/dashboard-header";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardFooter,
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
import {
	type BeeperApp,
	createBeeper,
	fetchBeeperApps,
	fetchBeepers,
} from "@/lib/api/beepers";
import { ApiError } from "@/lib/api/client";
import { fetchRunners } from "@/lib/api/runners";
import { withAuthRedirects } from "@/lib/auth/guards";
import { translateError } from "@/lib/i18n-labels";
import { browserTimezone } from "@/lib/timezone";
import { cn } from "@/lib/utils";
import { m } from "@/locale/paraglide/messages";

const accountRoute = getRouteApi("/$account_slug");

export const Route = createFileRoute("/$account_slug/beepers")({
	loader: withAuthRedirects(async ({ params }) => {
		const slug = params?.account_slug ?? "";
		const [{ beeper_apps: beeperApps }, { beepers }, { runners }] =
			await Promise.all([
				fetchBeeperApps(),
				fetchBeepers(slug),
				fetchRunners(slug),
			]);
		return { beeperApps, beepers, runners };
	}),
	component: BeepersPage,
});

function getBeeperAppMeta(slug: string) {
	switch (slug) {
		case "site-uptime":
			return {
				icon: Globe,
				tag: m.beepers_tag_availability,
				accent: "text-sky-500",
				well: "bg-sky-500/12 text-sky-500",
				band: "from-sky-500/18 via-sky-500/6 to-transparent",
			};
		case "ssl-expiry":
			return {
				icon: ShieldCheck,
				tag: m.beepers_tag_security,
				accent: "text-amber-500",
				well: "bg-amber-500/12 text-amber-500",
				band: "from-amber-500/18 via-amber-500/6 to-transparent",
			};
		case "heartbeat":
			return {
				icon: Radio,
				tag: m.beepers_tag_cron_worker,
				accent: "text-emerald-500",
				well: "bg-emerald-500/12 text-emerald-500",
				band: "from-emerald-500/18 via-emerald-500/6 to-transparent",
			};
		default:
			return {
				icon: Activity,
				tag: m.beepers_tag_probe,
				accent: "text-primary",
				well: "bg-primary/12 text-primary",
				band: "from-primary/18 via-primary/6 to-transparent",
			};
	}
}

function BeeperAppCard({
	beeperApp,
	onSelect,
}: {
	beeperApp: BeeperApp;
	onSelect: () => void;
}) {
	const meta = getBeeperAppMeta(beeperApp.slug);
	const Icon = meta.icon;
	const cron = beeperApp.default_cron || "*/5 * * * *";
	const metricCount = beeperApp.metrics.length;
	const isBuiltIn = beeperApp.official !== false;

	return (
		<Card className="relative flex h-full flex-col overflow-hidden pt-0 transition-all duration-200 hover:ring-foreground/15 hover:shadow-xs">
			<div
				className={cn(
					"relative flex items-end justify-between gap-3 bg-gradient-to-br px-4 pb-3 pt-6",
					meta.band,
				)}
			>
				{isBuiltIn ? (
					<div className="absolute right-3.5 top-2.5">
						<span className="inline-flex items-center gap-1 rounded-full border border-border/60 bg-background/80 px-2 py-0.5 text-[10px] font-medium tracking-tight text-muted-foreground backdrop-blur-xs">
							<span className="size-1.5 rounded-full bg-emerald-500" />
							{m.beepers_built_in()}
						</span>
					</div>
				) : null}

				<div
					className={cn(
						"flex size-11 items-center justify-center rounded-xl ring-1 ring-inset ring-foreground/8",
						meta.well,
					)}
				>
					<Icon className="size-5" />
				</div>
				<div className="flex flex-wrap items-center justify-end gap-1.5">
					<Badge variant="secondary" className="font-normal text-[11px]">
						{meta.tag()}
					</Badge>
					<Badge
						variant="outline"
						className="font-mono text-[10px] text-muted-foreground"
					>
						v{beeperApp.version}
					</Badge>
				</div>
			</div>

			<CardHeader className="flex-1">
				<CardTitle className="text-lg font-semibold">
					{beeperApp.name}
				</CardTitle>
				<CardDescription className="line-clamp-3 leading-relaxed min-h-[4.5rem]">
					{beeperApp.description}
				</CardDescription>
			</CardHeader>

			<CardContent>
				<div className="flex flex-wrap items-center gap-x-3 gap-y-1.5 text-xs text-muted-foreground">
					<span className="inline-flex items-center gap-1.5">
						<Clock className="size-3.5" />
						<span className="font-mono text-[11px]">{cron}</span>
					</span>
					{metricCount > 0 ? (
						<span>
							{metricCount}{" "}
							{metricCount === 1 ? m.beepers_metric() : m.beepers_metrics()}
						</span>
					) : null}
					{beeperApp.webhook_ping ? (
						<span className="inline-flex items-center gap-1">
							<Webhook className="size-3.5" />
							{m.beepers_ping()}
						</span>
					) : null}
				</div>
			</CardContent>

			<CardFooter className="mt-auto">
				<Button
					size="sm"
					variant="outline"
					className="w-full"
					onClick={onSelect}
				>
					{m.beepers_configure_install()}
					<ArrowRight data-icon="inline-end" />
				</Button>
			</CardFooter>
		</Card>
	);
}

function BeepersPage() {
	const { account_slug: slug } = accountRoute.useParams();
	const router = useRouter();
	const { beeperApps, beepers, runners } = Route.useLoaderData();
	const [selectedBeeperApp, setSelectedBeeperApp] = useState<BeeperApp | null>(
		null,
	);
	const [formTitle, setFormTitle] = useState("");
	const [formBody, setFormBody] = useState("");
	const [formCron, setFormCron] = useState("");
	const [formRunnerId, setFormRunnerId] = useState<string | null>(null);
	const [formRunnerTag, setFormRunnerTag] = useState<string | null>(null);
	const [formInputs, setFormInputs] = useState<Record<string, unknown>>({});
	const [submitting, setSubmitting] = useState(false);
	const [error, setError] = useState<string | null>(null);

	function handleSelectBeeperApp(beeperApp: BeeperApp) {
		setSelectedBeeperApp(beeperApp);
		setFormTitle(`${beeperApp.name}`);
		setFormBody("");
		setFormCron(beeperApp.default_cron || "*/5 * * * *");
		setFormRunnerId(null);
		setFormRunnerTag(null);
		const initialInputs: Record<string, unknown> = {};
		for (const input of beeperApp.inputs) {
			initialInputs[input.name] =
				input.default !== undefined ? input.default : "";
		}
		setFormInputs(initialInputs);
		setError(null);
	}

	async function handleInstall(event: React.FormEvent) {
		event.preventDefault();
		if (!selectedBeeperApp) return;

		setSubmitting(true);
		setError(null);

		try {
			const newBeeper = await createBeeper(slug, {
				title: formTitle.trim(),
				body: formBody.trim() || undefined,
				cron: formCron.trim(),
				timezone: browserTimezone(),
				beeper_app_id: selectedBeeperApp.id,
				runner_id: formRunnerId,
				runner_tag: formRunnerTag,
				config: formInputs,
			});

			setSelectedBeeperApp(null);
			await router.invalidate();
			await router.navigate({
				to: "/$account_slug/beepers/$beeperId",
				params: { account_slug: slug, beeperId: newBeeper.id },
			});
		} catch (err) {
			setError(
				err instanceof ApiError
					? err.message
					: translateError(err) || m.beepers_install_failed(),
			);
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
						params: { account_slug: slug },
					},
					{ label: m.nav_beepers(), isCurrentPage: true },
				]}
			/>

			<div className="flex flex-1 flex-col gap-8 p-4 md:p-6">
				<div className="flex flex-col gap-1">
					<h1 className="font-heading text-2xl font-bold tracking-tight sm:text-3xl">
						{m.beepers_gallery_title()}
					</h1>
					<p className="text-sm text-muted-foreground">
						{m.beepers_gallery_description()}
					</p>
				</div>

				{/* 1. Catalog / Apps Gallery on Top */}
				<div className="flex flex-col gap-3">
					<h2 className="font-heading text-lg font-semibold">
						{m.beepers_catalog()}
					</h2>
					<div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
						{beeperApps.map((beeperApp) => (
							<BeeperAppCard
								key={beeperApp.id}
								beeperApp={beeperApp}
								onSelect={() => handleSelectBeeperApp(beeperApp)}
							/>
						))}
					</div>
				</div>

				{beepers.length > 0 ? (
					<div className="flex flex-col gap-3">
						<h2 className="font-heading text-lg font-semibold">
							{m.beepers_your_beepers()}
						</h2>
						<BeeperList beepers={beepers} slug={slug} />
					</div>
				) : null}

				{/* 3. Install Configuration Modal Popup */}
				<Dialog
					open={selectedBeeperApp !== null}
					onOpenChange={(open) => {
						if (!open && !submitting) {
							setSelectedBeeperApp(null);
						}
					}}
				>
					<DialogContent className="sm:max-w-lg">
						{selectedBeeperApp ? (
							<form onSubmit={handleInstall} className="flex flex-col gap-4">
								<DialogHeader>
									<div className="flex items-center gap-3">
										{(() => {
											const meta = getBeeperAppMeta(selectedBeeperApp.slug);
											const Icon = meta.icon;
											return (
												<div
													className={cn(
														"flex size-9 items-center justify-center rounded-lg ring-1 ring-inset ring-foreground/8",
														meta.well,
													)}
												>
													<Icon className="size-4" />
												</div>
											);
										})()}
										<div>
											<DialogTitle className="text-lg">
												{m.beepers_install_title({
													name: selectedBeeperApp.name,
												})}
											</DialogTitle>
											<DialogDescription>
												{m.beepers_install_description()}
											</DialogDescription>
										</div>
									</div>
								</DialogHeader>

								<div className="flex flex-col gap-4 py-2">
									<div className="flex flex-col gap-2">
										<Label htmlFor="beeper-title">{m.beepers_title()}</Label>
										<Input
											id="beeper-title"
											required
											value={formTitle}
											onChange={(e) => setFormTitle(e.target.value)}
											disabled={submitting}
										/>
									</div>

									<div className="flex flex-col gap-2">
										<div className="flex items-center justify-between">
											<Label htmlFor="beeper-body">
												{m.beepers_body_remark()}
											</Label>
											<span className="text-[11px] text-muted-foreground">
												{m.common_optional()}
											</span>
										</div>
										<textarea
											id="beeper-body"
											rows={3}
											placeholder={m.beepers_body_placeholder()}
											value={formBody}
											onChange={(e) => setFormBody(e.target.value)}
											disabled={submitting}
											className="w-full rounded-lg border border-input bg-transparent px-2.5 py-1.5 text-sm transition-colors outline-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:pointer-events-none disabled:cursor-not-allowed disabled:bg-input/50 disabled:opacity-50 dark:bg-input/30"
										/>
									</div>

									<div className="flex flex-col gap-2">
										<Label htmlFor="beeper-cron">
											{m.beepers_cron_schedule()}
										</Label>
										<Input
											id="beeper-cron"
											required
											value={formCron}
											onChange={(e) => setFormCron(e.target.value)}
											disabled={submitting}
											className="font-mono text-sm"
										/>
										<p className="text-[11px] text-muted-foreground">
											{m.beepers_cron_default({
												cron: selectedBeeperApp.default_cron || "*/5 * * * *",
											})}
										</p>
									</div>

									{selectedBeeperApp.inputs.map((input) => (
										<div key={input.name} className="flex flex-col gap-2">
											<Label htmlFor={`input-${input.name}`}>
												{input.label}
												{input.required ? (
													<span className="text-destructive ml-1">*</span>
												) : null}
											</Label>
											<Input
												id={`input-${input.name}`}
												type={input.type === "number" ? "number" : "text"}
												required={input.required}
												min={input.min}
												max={input.max}
												placeholder={input.placeholder}
												value={String(formInputs[input.name] ?? "")}
												onChange={(e) => {
													const val =
														input.type === "number"
															? Number(e.target.value)
															: e.target.value;
													setFormInputs((curr) => ({
														...curr,
														[input.name]: val,
													}));
												}}
												disabled={submitting}
											/>
										</div>
									))}

									<RunnerRoutingPicker
										runners={runners}
										runnerId={formRunnerId}
										runnerTag={formRunnerTag}
										onChange={({ runner_id, runner_tag }) => {
											setFormRunnerId(runner_id);
											setFormRunnerTag(runner_tag);
										}}
										disabled={submitting}
									/>

									{error ? (
										<p className="text-sm text-destructive" role="alert">
											{error}
										</p>
									) : null}
								</div>

								<DialogFooter className="gap-2 sm:gap-0">
									<Button
										type="button"
										variant="ghost"
										size="sm"
										onClick={() => setSelectedBeeperApp(null)}
										disabled={submitting}
									>
										{m.common_cancel()}
									</Button>
									<Button
										type="submit"
										size="sm"
										disabled={submitting || !formTitle.trim()}
									>
										{submitting ? m.beepers_installing() : m.beepers_install()}
									</Button>
								</DialogFooter>
							</form>
						) : null}
					</DialogContent>
				</Dialog>
			</div>
		</>
	);
}
