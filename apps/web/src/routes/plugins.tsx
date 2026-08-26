import {
	createFileRoute,
	getRouteApi,
	useRouter,
} from "@tanstack/react-router";
import { Activity, Check, Plus, ShieldCheck, Timer } from "lucide-react";
import { useEffect, useState } from "react";

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

const accountRoute = getRouteApi("/$account_slug");

export const Route = createFileRoute("/plugins")({
	loader: withAuthRedirects(() => fetchPlugins()),
	component: PluginsPage,
});

function getPluginIcon(slug: string) {
	if (slug === "site-uptime") return <Activity className="size-5 text-primary" />;
	if (slug === "ssl-expiry") return <ShieldCheck className="size-5 text-amber-500" />;
	if (slug === "heartbeat") return <Timer className="size-5 text-emerald-500" />;
	return <Activity className="size-5 text-primary" />;
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
			initialInputs[input.name] = input.default !== undefined ? input.default : "";
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
				to: "/$account_slug/beeps_/$beepId",
				params: { account_slug: slug, beepId: newBeep.id },
			});
		} catch (err) {
			setError(err instanceof ApiError ? err.message : "Failed to install plugin.");
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
						Install monitoring probes and automated checks to watch your services and notify you on failure.
					</p>
				</div>

				<div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
					{plugins.map((plugin) => (
						<Card
							key={plugin.id}
							className={`flex flex-col justify-between transition-all ${
								selectedPlugin?.id === plugin.id
									? "border-primary ring-2 ring-primary/20 shadow-md"
									: "hover:border-primary/40"
							}`}
						>
							<CardHeader>
								<div className="flex items-center justify-between gap-2">
									<div className="flex items-center gap-2.5">
										{getPluginIcon(plugin.slug)}
										<CardTitle className="text-base font-semibold">{plugin.name}</CardTitle>
									</div>
									<Badge variant="outline" className="text-[10px]">
										v{plugin.version}
									</Badge>
								</div>
								<CardDescription className="text-xs pt-1.5">
									{plugin.description}
								</CardDescription>
							</CardHeader>
							<CardFooter className="pt-0">
								<Button
									size="sm"
									variant={selectedPlugin?.id === plugin.id ? "default" : "outline"}
									className="w-full"
									onClick={() => handleSelectPlugin(plugin)}
								>
									{selectedPlugin?.id === plugin.id ? (
										<>
											<Check data-icon="inline-start" />
											Configuring
										</>
									) : (
										<>
											<Plus data-icon="inline-start" />
											Configure & Install
										</>
									)}
								</Button>
							</CardFooter>
						</Card>
					))}
				</div>

				{selectedPlugin ? (
					<Card className="mt-4 max-w-2xl border-primary/40 shadow-sm">
						<CardHeader>
							<div className="flex items-center gap-2">
								{getPluginIcon(selectedPlugin.slug)}
								<CardTitle className="text-lg">Configure {selectedPlugin.name}</CardTitle>
							</div>
							<CardDescription>
								Fill in the parameters below to instantiate this monitor into your Beep schedule.
							</CardDescription>
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
										<div className="flex items-center justify-between">
											<Label htmlFor={`input-${input.name}`}>
												{input.label}
												{input.required ? <span className="text-destructive ml-1">*</span> : null}
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
												const val = input.type === "number" ? Number(e.target.value) : e.target.value;
												setFormInputs((curr) => ({ ...curr, [input.name]: val }));
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
							<CardFooter className="flex justify-end gap-2 border-t pt-4">
								<Button
									type="button"
									variant="ghost"
									size="sm"
									onClick={() => setSelectedPlugin(null)}
									disabled={submitting}
								>
									Cancel
								</Button>
								<Button type="submit" size="sm" disabled={submitting || !formTitle.trim()}>
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
