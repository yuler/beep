import {
	createFileRoute,
	getRouteApi,
	Link,
	useRouter,
} from "@tanstack/react-router";
import {
	Activity,
	ArrowRight,
	ArrowUpRight,
	Check,
	Clock,
	Globe,
	Radio,
	ShieldAlert,
	ShieldCheck,
	Webhook,
} from "lucide-react";
import { useState } from "react";

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
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
	type Beeper,
	type BeeperApp,
	createBeeper,
	fetchBeeperApps,
	fetchBeepers,
} from "@/lib/api/beepers";
import { ApiError } from "@/lib/api/client";
import { withAuthRedirects } from "@/lib/auth/guards";
import { browserTimezone } from "@/lib/timezone";
import { cn } from "@/lib/utils";

const accountRoute = getRouteApi("/$account_slug");

export const Route = createFileRoute("/$account_slug/beepers")({
	loader: withAuthRedirects(async ({ params }) => {
		const slug = params?.account_slug ?? "";
		const [{ beeper_apps: beeperApps }, { beepers }] = await Promise.all([
			fetchBeeperApps(),
			fetchBeepers(slug),
		]);
		return { beeperApps, beepers };
	}),
	component: BeepersPage,
});

function getBeeperAppMeta(slug: string) {
	switch (slug) {
		case "site-uptime":
			return {
				icon: Globe,
				tag: "Availability",
				accent: "text-sky-500",
				well: "bg-sky-500/12 text-sky-500",
				band: "from-sky-500/18 via-sky-500/6 to-transparent",
			};
		case "ssl-expiry":
			return {
				icon: ShieldCheck,
				tag: "Security",
				accent: "text-amber-500",
				well: "bg-amber-500/12 text-amber-500",
				band: "from-amber-500/18 via-amber-500/6 to-transparent",
			};
		case "heartbeat":
			return {
				icon: Radio,
				tag: "Cron & Worker",
				accent: "text-emerald-500",
				well: "bg-emerald-500/12 text-emerald-500",
				band: "from-emerald-500/18 via-emerald-500/6 to-transparent",
			};
		default:
			return {
				icon: Activity,
				tag: "Probe",
				accent: "text-primary",
				well: "bg-primary/12 text-primary",
				band: "from-primary/18 via-primary/6 to-transparent",
			};
	}
}

function BeeperAppCard({
	beeperApp,
	selected,
	onSelect,
}: {
	beeperApp: BeeperApp;
	selected: boolean;
	onSelect: () => void;
}) {
	const meta = getBeeperAppMeta(beeperApp.slug);
	const Icon = meta.icon;
	const cron = beeperApp.default_cron || "*/5 * * * *";
	const metricCount = beeperApp.metrics.length;

	return (
		<Card
			className={cn(
				"pt-0 transition-all duration-200",
				selected
					? "ring-2 ring-primary/25 shadow-md"
					: "hover:ring-foreground/15 hover:shadow-xs",
			)}
		>
			<div
				className={cn(
					"relative flex items-end justify-between gap-3 bg-gradient-to-br px-4 pb-3 pt-5",
					meta.band,
				)}
			>
				<div
					className={cn(
						"flex size-11 items-center justify-center rounded-xl ring-1 ring-inset ring-foreground/8",
						meta.well,
					)}
				>
					<Icon className="size-5" />
				</div>
				<div className="flex flex-wrap items-center justify-end gap-1.5">
					<Badge variant="secondary" className="font-normal">
						{meta.tag}
					</Badge>
					<Badge
						variant="outline"
						className="font-mono text-[10px] text-muted-foreground"
					>
						v{beeperApp.version}
					</Badge>
				</div>
			</div>

			<CardHeader>
				<CardTitle className="text-lg font-semibold">
					{beeperApp.name}
				</CardTitle>
				<CardDescription className="line-clamp-3 leading-relaxed">
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
							{metricCount} {metricCount === 1 ? "metric" : "metrics"}
						</span>
					) : null}
					{beeperApp.webhook_ingest ? (
						<span className="inline-flex items-center gap-1">
							<Webhook className="size-3.5" />
							Ingest
						</span>
					) : null}
				</div>
			</CardContent>

			<CardFooter>
				<Button
					size="sm"
					variant={selected ? "default" : "outline"}
					className="w-full"
					onClick={onSelect}
				>
					{selected ? (
						<>
							<Check data-icon="inline-start" />
							Configuring
						</>
					) : (
						<>
							Configure & Install
							<ArrowRight data-icon="inline-end" />
						</>
					)}
				</Button>
			</CardFooter>
		</Card>
	);
}

function BeepersPage() {
	const { account_slug: slug } = accountRoute.useParams();
	const router = useRouter();
	const { beeperApps, beepers } = Route.useLoaderData();
	const [selectedBeeperApp, setSelectedBeeperApp] = useState<BeeperApp | null>(
		null,
	);
	const [formTitle, setFormTitle] = useState("");
	const [formCron, setFormCron] = useState("");
	const [formInputs, setFormInputs] = useState<Record<string, unknown>>({});
	const [submitting, setSubmitting] = useState(false);
	const [error, setError] = useState<string | null>(null);

	function handleSelectBeeperApp(beeperApp: BeeperApp) {
		setSelectedBeeperApp(beeperApp);
		setFormTitle(`${beeperApp.name}`);
		setFormCron(beeperApp.default_cron || "*/5 * * * *");
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
				cron: formCron.trim(),
				timezone: browserTimezone(),
				beeper_app_id: selectedBeeperApp.id,
				config: formInputs,
			});

			setSelectedBeeperApp(null);
			await router.navigate({
				to: "/$account_slug/beepers/$beeperId",
				params: { account_slug: slug, beeperId: newBeeper.id },
			});
		} catch (err) {
			setError(
				err instanceof ApiError ? err.message : "Failed to install beeper.",
			);
			setSubmitting(false);
		}
	}

	return (
		<>
			<DashboardHeader
				breadcrumbs={[
					{
						label: "Home",
						to: "/$account_slug",
						params: { account_slug: slug },
					},
					{ label: "Beepers", isCurrentPage: true },
				]}
			/>

			<div className="flex flex-1 flex-col gap-6 p-4 md:p-6">
				<div className="flex flex-col gap-1">
					<h1 className="font-heading text-2xl font-bold tracking-tight sm:text-3xl">
						Beeper Gallery
					</h1>
					<p className="text-sm text-muted-foreground">
						Install monitoring probes and automated signal receivers to watch
						your services and send notifications when issues arise.
					</p>
				</div>

				{beepers.length > 0 ? (
					<div className="flex flex-col gap-3">
						<h2 className="font-heading text-lg font-semibold">Your Beepers</h2>
						<div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
							{beepers.map((beeper: Beeper) => (
								<Card
									key={beeper.id}
									size="sm"
									className="group relative overflow-hidden transition-all duration-150 hover:border-primary/40 hover:shadow-xs"
								>
									<Link
										to="/$account_slug/beepers/$beeperId"
										params={{
											account_slug: slug,
											beeperId: beeper.id,
										}}
										className="block p-4 focus:outline-none focus-visible:ring-2 focus-visible:ring-ring"
									>
										<div className="flex items-center justify-between gap-2">
											<div className="flex items-center gap-1.5">
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
												<Badge
													variant={
														beeper.alert_state === "alerting"
															? "destructive"
															: "outline"
													}
													className="gap-1 text-[10px] font-medium"
												>
													{beeper.alert_state === "alerting" ? (
														<ShieldAlert className="size-2.5" />
													) : (
														<ShieldCheck className="size-2.5 text-emerald-600 dark:text-emerald-400" />
													)}
													{beeper.alert_state.toUpperCase()}
												</Badge>
											</div>
											<span className="font-mono text-xs text-muted-foreground">
												{beeper.cron}
											</span>
										</div>

										<div className="mt-2.5 flex items-start justify-between gap-3">
											<h3 className="font-heading text-base font-semibold text-foreground transition-colors group-hover:text-primary">
												{beeper.title}
											</h3>
											<ArrowUpRight className="size-4 shrink-0 text-muted-foreground opacity-0 transition-opacity group-hover:opacity-100 group-hover:text-primary" />
										</div>

										<p className="mt-1 text-xs text-muted-foreground">
											{beeper.beeper_app?.name}
										</p>
									</Link>
								</Card>
							))}
						</div>
					</div>
				) : null}

				<div className="flex flex-col gap-3">
					<h2 className="font-heading text-lg font-semibold">Catalog</h2>
					<div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
						{beeperApps.map((beeperApp) => (
							<BeeperAppCard
								key={beeperApp.id}
								beeperApp={beeperApp}
								selected={selectedBeeperApp?.id === beeperApp.id}
								onSelect={() => handleSelectBeeperApp(beeperApp)}
							/>
						))}
					</div>
				</div>

				{selectedBeeperApp ? (
					<Card className="mt-4 max-w-2xl">
						<CardHeader>
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
									<CardTitle className="text-lg">
										Configure {selectedBeeperApp.name}
									</CardTitle>
									<CardDescription>
										Fill in the parameters below to install this beeper into
										your account.
									</CardDescription>
								</div>
							</div>
						</CardHeader>
						<form onSubmit={handleInstall}>
							<CardContent className="flex flex-col gap-4">
								<div className="flex flex-col gap-2">
									<Label htmlFor="beeper-title">Beeper Title</Label>
									<Input
										id="beeper-title"
										required
										value={formTitle}
										onChange={(e) => setFormTitle(e.target.value)}
										disabled={submitting}
									/>
								</div>

								<div className="flex flex-col gap-2">
									<Label htmlFor="beeper-cron">Cron Schedule</Label>
									<Input
										id="beeper-cron"
										required
										value={formCron}
										onChange={(e) => setFormCron(e.target.value)}
										disabled={submitting}
										className="font-mono text-sm"
									/>
									<p className="text-[11px] text-muted-foreground">
										Default: {selectedBeeperApp.default_cron || "*/5 * * * *"}
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

								{error ? (
									<p className="text-sm text-destructive" role="alert">
										{error}
									</p>
								) : null}
							</CardContent>
							<CardFooter className="justify-end gap-2">
								<Button
									type="button"
									variant="ghost"
									size="sm"
									onClick={() => setSelectedBeeperApp(null)}
									disabled={submitting}
								>
									Cancel
								</Button>
								<Button
									type="submit"
									size="sm"
									disabled={submitting || !formTitle.trim()}
								>
									{submitting ? "Installing…" : "Install Beeper"}
								</Button>
							</CardFooter>
						</form>
					</Card>
				) : null}
			</div>
		</>
	);
}
