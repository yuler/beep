import {
	Activity,
	ArrowRight,
	Bell,
	CheckCircle2,
	MessageSquarePlus,
	Send,
	Sparkles,
} from "lucide-react";
import { SiteAuthButton } from "@/components/layout/site-auth-button";
import { Badge } from "@/components/ui/badge";
import { buttonVariants } from "@/components/ui/button";
import { Card } from "@/components/ui/card";
import { cn } from "@/lib/utils";
import { m } from "@/locale/paraglide/messages";

export function HeroSection() {
	return (
		<section className="relative overflow-hidden pt-12 pb-20 md:pt-20 md:pb-28">
			{/* Subtle background glow */}
			<div
				className="pointer-events-none absolute inset-x-0 top-0 -z-10 flex transform-gpu justify-center overflow-hidden blur-3xl"
				aria-hidden="true"
			>
				<div className="aspect-[1155/678] w-[72.1875rem] flex-none bg-gradient-to-tr from-primary/20 via-primary/5 to-transparent opacity-60 dark:opacity-30" />
			</div>

			<div className="mx-auto max-w-5xl px-4 sm:px-6">
				{/* Main heading & CTA */}
				<div className="mx-auto flex max-w-3xl flex-col items-center text-center">
					<div className="inline-flex items-center gap-2 rounded-full border border-primary/20 bg-primary/5 px-3.5 py-1 text-xs font-medium text-primary shadow-xs transition-colors hover:bg-primary/10">
						<Sparkles className="size-3.5 animate-pulse text-primary" />
						<span>{m.marketing_badge()}</span>
					</div>

					<h1 className="font-heading mt-6 text-4xl font-bold tracking-tight text-foreground text-balance sm:text-5xl md:text-6xl">
						{m.marketing_hero_title()}
					</h1>

					<p className="mt-6 text-lg leading-relaxed text-muted-foreground text-pretty sm:text-xl">
						{m.marketing_hero_description()}
					</p>

					<div className="mt-8 flex flex-wrap items-center justify-center gap-3.5">
						<SiteAuthButton
							signInLabel={m.common_get_started()}
							dashboardLabel={m.common_open_dashboard()}
							size="lg"
							className="h-11 px-6 shadow-sm"
						/>
						<a
							href="#features"
							className={cn(
								buttonVariants({ size: "lg", variant: "outline" }),
								"h-11 px-6 gap-2 text-muted-foreground hover:text-foreground",
							)}
						>
							<span>{m.common_view_docs()}</span>
							<ArrowRight className="size-4" />
						</a>
					</div>
				</div>

				{/* High-fidelity Product Showcase Mockup */}
				<div className="relative mx-auto mt-14 max-w-4xl">
					<div className="relative rounded-2xl border border-border/80 bg-card/60 p-2 shadow-2xl backdrop-blur-md ring-1 ring-foreground/5 md:p-3">
						{/* Window Chrome */}
						<div className="flex items-center justify-between border-b border-border/60 px-3 py-2 text-xs text-muted-foreground">
							<div className="flex items-center gap-1.5">
								<span className="size-2.5 rounded-full bg-rose-500/80" />
								<span className="size-2.5 rounded-full bg-amber-500/80" />
								<span className="size-2.5 rounded-full bg-emerald-500/80" />
							</div>
							<div className="font-mono text-[11px] tracking-wide text-muted-foreground/70">
								app.beep.dev
							</div>
							<div className="w-8" />
						</div>

						{/* Mock Dashboard Body */}
						<div className="space-y-4 p-4 md:p-6">
							{/* Mock Prompt Creator */}
							<div className="rounded-xl border border-border bg-background p-4 shadow-xs">
								<div className="flex items-center gap-2 text-xs font-semibold text-muted-foreground uppercase tracking-wider">
									<MessageSquarePlus className="size-3.5 text-primary" />
									<span>Quick Beep</span>
								</div>
								<div className="mt-3 flex flex-col gap-3 sm:flex-row sm:items-center">
									<div className="flex flex-1 items-center gap-2 rounded-lg border border-border/80 bg-muted/30 px-3 py-2 text-sm text-foreground">
										<Sparkles className="size-4 shrink-0 text-primary" />
										<span className="truncate">
											{m.marketing_hero_prompt_input()}
										</span>
									</div>
									<div className="inline-flex shrink-0 items-center justify-center gap-1.5 rounded-lg bg-primary px-3.5 py-2 text-xs font-medium text-primary-foreground shadow-xs">
										<Send className="size-3.5" />
										<span>Schedule</span>
									</div>
								</div>
								<div className="mt-2.5 flex items-center gap-2 text-xs text-muted-foreground">
									<CheckCircle2 className="size-3.5 text-emerald-500" />
									<span>{m.marketing_hero_prompt_scheduled()}</span>
								</div>
							</div>

							{/* Mock Bottom Cards */}
							<div className="grid gap-3 sm:grid-cols-2">
								{/* Probe status card */}
								<Card size="sm" className="bg-background/80">
									<div className="flex items-center justify-between p-3.5">
										<div className="flex items-center gap-3">
											<div className="flex size-8 items-center justify-center rounded-lg bg-emerald-500/10 text-emerald-600 dark:text-emerald-400">
												<Activity className="size-4" />
											</div>
											<div>
												<div className="font-heading text-sm font-semibold text-foreground">
													{m.marketing_hero_probe_name()}
												</div>
												<div className="text-xs text-muted-foreground">
													{m.marketing_hero_probe_cron()}
												</div>
											</div>
										</div>
										<Badge
											variant="outline"
											className="border-emerald-500/30 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 text-[11px]"
										>
											200 OK • 42ms
										</Badge>
									</div>
								</Card>

								{/* Notification Channels card */}
								<Card size="sm" className="bg-background/80">
									<div className="flex items-center justify-between p-3.5">
										<div className="flex items-center gap-3">
											<div className="flex size-8 items-center justify-center rounded-lg bg-primary/10 text-primary">
												<Bell className="size-4" />
											</div>
											<div>
												<div className="font-heading text-sm font-semibold text-foreground">
													{m.marketing_hero_channels_title()}
												</div>
												<div className="text-xs text-muted-foreground">
													{m.marketing_hero_dispatch_status()}
												</div>
											</div>
										</div>
										<div className="flex items-center gap-1">
											<span className="inline-flex size-2 rounded-full bg-emerald-500" />
											<span className="text-[11px] font-medium text-emerald-600 dark:text-emerald-400">
												Live
											</span>
										</div>
									</div>
								</Card>
							</div>
						</div>
					</div>
				</div>
			</div>
		</section>
	);
}
