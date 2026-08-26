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
import { createBeep } from "@/lib/api/beeps";
import { ApiError } from "@/lib/api/client";
import { fetchPlugins, type Plugin } from "@/lib/api/plugins";
import { withAuthRedirects } from "@/lib/auth/guards";
import { browserTimezone } from "@/lib/timezone";
import { cn } from "@/lib/utils";

const accountRoute = getRouteApi("/$account_slug");

export const Route = createFileRoute("/$account_slug/plugins")({
	loader: withAuthRedirects(() => fetchPlugins()),
	component: PluginsPage,
});

function getPluginMeta(slug: string) {
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

function PluginCard({
	plugin,
	selected,
	onSelect,
}: {
	plugin: Plugin;
	selected: boolean;
	onSelect: () => void;
}) {
	const meta = getPluginMeta(plugin.slug);
	const Icon = meta.icon;
	const cron = plugin.default_cron || "*/5 * * * *";
	const metricCount = plugin.metrics.length;

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
						v{plugin.version}
					</Badge>
				</div>
			</div>

			<CardHeader>
				<CardTitle className="text-lg font-semibold">{plugin.name}</CardTitle>
				<CardDescription className="line-clamp-3 leading-relaxed">
					{plugin.description}
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
					{plugin.webhook_ingest ? (
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

function PluginsPage() {
	const { account_slug: slug } = accountRoute.useParams();
	const router = useRouter();
	const { plugins } = Route.useLoaderData();
	const [selectedPlugin, setSelectedPlugin] = useState<Plugin | null>(null);
	const [formTitle, setFormTitle] = useState("");
	const [formCron, setFormCron] = useState("");
	const [formInputs, setFormInputs] = useState<Record<string, unknown>>({});
	const [submitting, setSubmitting] = useState(false);
	const [error, setError] = useState<string | null>(null);

	function handleSelectPlugin(plugin: Plugin) {
		setSelectedPlugin(plugin);
		setFormTitle(`${plugin.name} Monitor`);
		setFormCron(plugin.default_cron || "*/5 * * * *");
		const initialInputs: Record<string, unknown> = {};
		for (const input of plugin.inputs) {
			initialInputs[input.name] =
				input.default !== undefined ? input.default : "";
		}
		setFormInputs(initialInputs);
		setError(null);
	}

	async function handleInstall(event: React.FormEvent) {
		event.preventDefault();
		if (!selectedPlugin) return;

		setSubmitting(true);
		setError(null);

		try {
			const newBeep = await createBeep(slug, {
				title: formTitle.trim(),
				kind: "recurring",
				cron: formCron.trim(),
				timezone: browserTimezone(),
				plugin_id: selectedPlugin.id,
				plugin_config: formInputs,
			});

			await router.navigate({
				to: "/$account_slug/beeps/$beepId",
				params: { account_slug: slug, beepId: newBeep.id },
			});
		} catch (err) {
			setError(
				err instanceof ApiError ? err.message : "Failed to install plugin.",
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
					{ label: "Plugins", isCurrentPage: true },
				]}
			/>

			<div className="flex flex-1 flex-col gap-6 p-4 md:p-6">
				<div className="flex flex-col gap-1">
					<h1 className="font-heading text-2xl font-bold tracking-tight sm:text-3xl">
						Plugin Gallery
					</h1>
					<p className="text-sm text-muted-foreground">
						Install monitoring probes and automated checks to watch your
						services and notify you on failure.
					</p>
				</div>

				<div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
					{plugins.map((plugin) => (
						<PluginCard
							key={plugin.id}
							plugin={plugin}
							selected={selectedPlugin?.id === plugin.id}
							onSelect={() => handleSelectPlugin(plugin)}
						/>
					))}
				</div>

				{selectedPlugin ? (
					<Card className="mt-4 max-w-2xl">
						<CardHeader>
							<div className="flex items-center gap-3">
								{(() => {
									const meta = getPluginMeta(selectedPlugin.slug);
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
										Configure {selectedPlugin.name}
									</CardTitle>
									<CardDescription>
										Fill in the parameters below to instantiate this monitor
										into your Beep schedule.
									</CardDescription>
								</div>
							</div>
						</CardHeader>
						<form onSubmit={handleInstall}>
							<CardContent className="flex flex-col gap-4">
								<div className="flex flex-col gap-2">
									<Label htmlFor="plugin-title">Monitor Title</Label>
									<Input
										id="plugin-title"
										required
										value={formTitle}
										onChange={(e) => setFormTitle(e.target.value)}
										disabled={submitting}
									/>
								</div>

								<div className="flex flex-col gap-2">
									<Label htmlFor="plugin-cron">Cron Schedule</Label>
									<Input
										id="plugin-cron"
										required
										value={formCron}
										onChange={(e) => setFormCron(e.target.value)}
										disabled={submitting}
										className="font-mono text-sm"
									/>
									<p className="text-[11px] text-muted-foreground">
										Default: {selectedPlugin.default_cron || "*/5 * * * *"}
									</p>
								</div>

								{selectedPlugin.inputs.map((input) => (
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
									onClick={() => setSelectedPlugin(null)}
									disabled={submitting}
								>
									Cancel
								</Button>
								<Button
									type="submit"
									size="sm"
									disabled={submitting || !formTitle.trim()}
								>
									{submitting ? "Installing…" : "Create & Start Monitoring"}
								</Button>
							</CardFooter>
						</form>
					</Card>
				) : null}
			</div>
		</>
	);
}
