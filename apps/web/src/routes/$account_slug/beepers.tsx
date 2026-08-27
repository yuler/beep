import {
	createFileRoute,
	getRouteApi,
	useRouter,
} from "@tanstack/react-router";
import {
	Activity,
	ArrowRight,
	Check,
	Clock,
	Globe,
	Radio,
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
	createBeeperInstall,
	fetchBeeperInstalls,
	fetchBeepers,
	type Beeper,
	type BeeperInstall,
} from "@/lib/api/beepers";
import { ApiError } from "@/lib/api/client";
import { withAuthRedirects } from "@/lib/auth/guards";
import { browserTimezone } from "@/lib/timezone";
import { cn } from "@/lib/utils";

const accountRoute = getRouteApi("/$account_slug");

export const Route = createFileRoute("/$account_slug/beepers")({
	loader: withAuthRedirects(async ({ params }) => {
		const slug = params?.account_slug ?? "";
		const [{ beepers }, { beeper_installs: installs }] = await Promise.all([
			fetchBeepers(),
			fetchBeeperInstalls(slug),
		]);
		return { beepers, installs };
	}),
	component: BeepersPage,
});

function getBeeperMeta(slug: string) {
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

function BeeperCard({
	beeper,
	selected,
	onSelect,
}: {
	beeper: Beeper;
	selected: boolean;
	onSelect: () => void;
}) {
	const meta = getBeeperMeta(beeper.slug);
	const Icon = meta.icon;
	const cron = beeper.default_cron || "*/5 * * * *";
	const metricCount = beeper.metrics.length;

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
						v{beeper.version}
					</Badge>
				</div>
			</div>

			<CardHeader>
				<CardTitle className="text-lg font-semibold">{beeper.name}</CardTitle>
				<CardDescription className="line-clamp-3 leading-relaxed">
					{beeper.description}
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
					{beeper.webhook_ingest ? (
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
	const { beepers, installs } = Route.useLoaderData();
	const [selectedBeeper, setSelectedBeeper] = useState<Beeper | null>(null);
	const [formTitle, setFormTitle] = useState("");
	const [formCron, setFormCron] = useState("");
	const [formInputs, setFormInputs] = useState<Record<string, unknown>>({});
	const [submitting, setSubmitting] = useState(false);
	const [error, setError] = useState<string | null>(null);

	function handleSelectBeeper(beeper: Beeper) {
		setSelectedBeeper(beeper);
		setFormTitle(`${beeper.name}`);
		setFormCron(beeper.default_cron || "*/5 * * * *");
		const initialInputs: Record<string, unknown> = {};
		for (const input of beeper.inputs) {
			initialInputs[input.name] =
				input.default !== undefined ? input.default : "";
		}
		setFormInputs(initialInputs);
		setError(null);
	}

	async function handleInstall(event: React.FormEvent) {
		event.preventDefault();
		if (!selectedBeeper) return;

		setSubmitting(true);
		setError(null);

		try {
			await createBeeperInstall(slug, {
				title: formTitle.trim(),
				cron: formCron.trim(),
				timezone: browserTimezone(),
				beeper_id: selectedBeeper.id,
				config: formInputs,
			});

			setSelectedBeeper(null);
			await router.invalidate();
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
						Install monitoring probes and automated checkers to watch your
						services and send notifications when issues arise.
					</p>
				</div>

				{installs.length > 0 ? (
					<div className="flex flex-col gap-3">
						<h2 className="font-heading text-lg font-semibold">Your Beepers</h2>
						<div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
							{installs.map((install: BeeperInstall) => (
								<Card key={install.id} size="sm">
									<CardHeader>
										<div className="flex items-center justify-between">
											<CardTitle className="text-base">{install.title}</CardTitle>
											<Badge
												variant={
													install.alert_state === "alerting"
														? "destructive"
														: "outline"
												}
											>
												{install.alert_state}
											</Badge>
										</div>
										<CardDescription className="text-xs">
											{install.beeper?.name} · Cron: {install.cron}
										</CardDescription>
									</CardHeader>
								</Card>
							))}
						</div>
					</div>
				) : null}

				<div className="flex flex-col gap-3">
					<h2 className="font-heading text-lg font-semibold">Catalog</h2>
					<div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
						{beepers.map((beeper) => (
							<BeeperCard
								key={beeper.id}
								beeper={beeper}
								selected={selectedBeeper?.id === beeper.id}
								onSelect={() => handleSelectBeeper(beeper)}
							/>
						))}
					</div>
				</div>

				{selectedBeeper ? (
					<Card className="mt-4 max-w-2xl">
						<CardHeader>
							<div className="flex items-center gap-3">
								{(() => {
									const meta = getBeeperMeta(selectedBeeper.slug);
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
										Configure {selectedBeeper.name}
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
									<Label htmlFor="beeper-title">Install Title</Label>
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
										Default: {selectedBeeper.default_cron || "*/5 * * * *"}
									</p>
								</div>

								{selectedBeeper.inputs.map((input) => (
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
									onClick={() => setSelectedBeeper(null)}
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
