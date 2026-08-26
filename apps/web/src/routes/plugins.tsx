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
	Sparkles,
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

export const Route = createFileRoute("/plugins")({
	loader: withAuthRedirects(() => fetchPlugins()),
	component: PluginsPage,
});

function getPluginMeta(slug: string) {
	switch (slug) {
		case "site-uptime":
			return {
				icon: Globe,
				color: "text-blue-500",
				bgColor: "bg-blue-500/10 dark:bg-blue-500/20",
				borderColor: "group-hover:border-blue-500/30",
				tag: "Availability",
			};
		case "ssl-expiry":
			return {
				icon: ShieldCheck,
				color: "text-amber-500",
				bgColor: "bg-amber-500/10 dark:bg-amber-500/20",
				borderColor: "group-hover:border-amber-500/30",
				tag: "Security",
			};
		case "heartbeat":
			return {
				icon: Radio,
				color: "text-emerald-500",
				bgColor: "bg-emerald-500/10 dark:bg-emerald-500/20",
				borderColor: "group-hover:border-emerald-500/30",
				tag: "Cron & Worker",
			};
		default:
			return {
				icon: Activity,
				color: "text-primary",
				bgColor: "bg-primary/10",
				borderColor: "group-hover:border-primary/30",
				tag: "Probe",
			};
	}
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

			<div className="flex flex-1 flex-col gap-8 p-4 md:p-8 max-w-6xl mx-auto w-full">
				{/* Top Hero Section */}
				<div className="flex flex-col gap-2">
					<div className="inline-flex items-center gap-2 text-xs font-semibold uppercase tracking-wider text-primary">
						<Sparkles className="size-3.5" />
						Monitoring Ecosystem
					</div>
					<h1 className="font-heading text-3xl font-bold tracking-tight sm:text-4xl text-foreground">
						Plugin Gallery
					</h1>
					<p className="text-sm sm:text-base text-muted-foreground max-w-2xl">
						Install reliable health checks and background probes. Beep runs
						checks automatically and only notifies you when service disruption
						occurs.
					</p>
				</div>

				{/* Plugin Cards Grid */}
				<div className="grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
					{plugins.map((plugin) => {
						const meta = getPluginMeta(plugin.slug);
						const IconComponent = meta.icon;
						const isSelected = selectedPlugin?.id === plugin.id;

						return (
							<button
								key={plugin.id}
								type="button"
								onClick={() => handleSelectPlugin(plugin)}
								className={cn(
									"group relative flex w-full flex-col justify-between rounded-2xl border bg-card p-6 text-left text-card-foreground shadow-xs transition-all duration-200 hover:shadow-md",
									isSelected
										? "border-primary ring-2 ring-primary/20 bg-primary/[0.02]"
										: "border-border/80 hover:border-foreground/20 hover:-translate-y-0.5",
								)}
							>
								<div>
									{/* Card Top: Icon & Badges */}
									<div className="flex items-start justify-between gap-3 mb-4">
										<div
											className={cn(
												"flex size-11 items-center justify-center rounded-xl transition-transform group-hover:scale-105",
												meta.bgColor,
											)}
										>
											<IconComponent className={cn("size-5.5", meta.color)} />
										</div>
										<div className="flex items-center gap-1.5">
											<Badge
												variant="secondary"
												className="font-normal text-[11px] px-2 py-0.5 bg-muted/60 text-muted-foreground"
											>
												{meta.tag}
											</Badge>
											<Badge
												variant="outline"
												className="font-mono text-[10px] px-1.5 py-0 text-muted-foreground border-border/80"
											>
												v{plugin.version}
											</Badge>
										</div>
									</div>

									{/* Title & Description */}
									<h3 className="font-heading text-lg font-semibold tracking-tight text-foreground transition-colors group-hover:text-primary">
										{plugin.name}
									</h3>
									<p className="mt-2 text-xs sm:text-sm leading-relaxed text-muted-foreground line-clamp-3">
										{plugin.description}
									</p>
								</div>

								{/* Card Bottom Meta & Button */}
								<div className="mt-6 pt-4 border-t border-border/60 flex items-center justify-between">
									<div className="flex items-center gap-1.5 text-xs text-muted-foreground">
										<Clock className="size-3.5" />
										<span className="font-mono text-[11px]">
											{plugin.default_cron || "*/5 * * * *"}
										</span>
									</div>

									<span
										className={cn(
											"inline-flex h-8 items-center rounded-lg px-3 text-xs font-medium",
											isSelected
												? "bg-primary text-primary-foreground"
												: "bg-secondary text-secondary-foreground",
										)}
									>
										{isSelected ? (
											<>
												<Check className="size-3.5 mr-1" />
												Selected
											</>
										) : (
											<>
												Install
												<ArrowRight className="size-3.5 ml-1 transition-transform group-hover:translate-x-0.5" />
											</>
										)}
									</span>
								</div>
							</button>
						);
					})}
				</div>

				{/* Configuration Modal / Inline Form */}
				{selectedPlugin ? (
					<Card className="max-w-2xl border-primary/40 shadow-lg rounded-2xl overflow-hidden animate-in fade-in-50 duration-200">
						<CardHeader className="bg-muted/30 border-b pb-4">
							<div className="flex items-center justify-between">
								<div className="flex items-center gap-3">
									{(() => {
										const meta = getPluginMeta(selectedPlugin.slug);
										const Icon = meta.icon;
										return (
											<div
												className={cn(
													"flex size-9 items-center justify-center rounded-lg",
													meta.bgColor,
												)}
											>
												<Icon className={cn("size-5", meta.color)} />
											</div>
										);
									})()}
									<div>
										<CardTitle className="text-lg font-semibold">
											{selectedPlugin.name}
										</CardTitle>
										<CardDescription className="text-xs">
											Configure your target parameters to launch this monitor.
										</CardDescription>
									</div>
								</div>
								<Badge variant="outline" className="font-mono text-xs">
									v{selectedPlugin.version}
								</Badge>
							</div>
						</CardHeader>

						<form onSubmit={handleInstall}>
							<CardContent className="flex flex-col gap-4.5 pt-6">
								<div className="flex flex-col gap-1.5">
									<Label
										htmlFor="plugin-title"
										className="text-xs font-semibold"
									>
										Monitor Title
									</Label>
									<Input
										id="plugin-title"
										required
										value={formTitle}
										onChange={(e) => setFormTitle(e.target.value)}
										disabled={submitting}
										className="rounded-lg text-sm"
										placeholder="e.g. Production API Uptime"
									/>
								</div>

								<div className="flex flex-col gap-1.5">
									<Label
										htmlFor="plugin-cron"
										className="text-xs font-semibold"
									>
										Check Interval (Cron)
									</Label>
									<Input
										id="plugin-cron"
										required
										value={formCron}
										onChange={(e) => setFormCron(e.target.value)}
										disabled={submitting}
										className="font-mono text-sm rounded-lg"
									/>
									<p className="text-[11px] text-muted-foreground">
										Standard 5-part cron syntax. Default is{" "}
										<code className="bg-muted px-1 py-0.5 rounded font-mono text-[10px]">
											{selectedPlugin.default_cron || "*/5 * * * *"}
										</code>
									</p>
								</div>

								<div className="pt-2 border-t flex flex-col gap-4">
									<span className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
										Plugin Inputs
									</span>

									{selectedPlugin.inputs.map((input) => (
										<div key={input.name} className="flex flex-col gap-1.5">
											<div className="flex items-center justify-between">
												<Label
													htmlFor={`input-${input.name}`}
													className="text-xs font-medium"
												>
													{input.label}
													{input.required ? (
														<span className="text-destructive ml-1">*</span>
													) : null}
												</Label>
											</div>
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
												className="rounded-lg text-sm"
											/>
										</div>
									))}
								</div>

								{error ? (
									<p
										className="text-xs text-destructive bg-destructive/10 p-2.5 rounded-lg border border-destructive/20"
										role="alert"
									>
										{error}
									</p>
								) : null}
							</CardContent>

							<CardFooter className="flex justify-end gap-2.5 border-t bg-muted/20 px-6 py-4">
								<Button
									type="button"
									variant="ghost"
									size="sm"
									onClick={() => setSelectedPlugin(null)}
									disabled={submitting}
									className="rounded-lg"
								>
									Cancel
								</Button>
								<Button
									type="submit"
									size="sm"
									disabled={submitting || !formTitle.trim()}
									className="rounded-lg font-medium"
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
