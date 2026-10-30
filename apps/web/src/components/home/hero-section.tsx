import { ArrowRight, Sparkles } from "lucide-react";
import { SiteAuthButton } from "@/components/layout/site-auth-button";
import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { m } from "@/locale/paraglide/messages";

export function HeroSection() {
	return (
		<section className="relative overflow-hidden pt-12 pb-16 md:pt-20 md:pb-24 bg-ground text-ink">
			<div className="mx-auto max-w-[1080px] px-4 md:px-6 lg:px-8">
				{/* Main heading & CTA */}
				<div className="mx-auto flex max-w-3xl flex-col items-center text-center">
					<div className="inline-flex items-center gap-2 border border-line px-3 py-1 type-kicker text-gray">
						<Sparkles className="size-3 text-ink" />
						<span>{m.marketing_badge()}</span>
					</div>

					<h1 className="type-title mt-6 text-3xl font-medium tracking-tight text-ink sm:text-4xl md:text-5xl">
						{m.marketing_hero_title()}
					</h1>

					<p className="type-meta mt-5 text-base sm:text-lg text-gray max-w-2xl">
						{m.marketing_hero_description()}
					</p>

					<div className="mt-8 flex flex-wrap items-center justify-center gap-3">
						<SiteAuthButton
							signInLabel={m.common_get_started()}
							dashboardLabel={m.common_open_dashboard()}
							variant="solid"
							size="lg"
							className="px-6"
						/>
						<a
							href="#features"
							className={cn(
								buttonVariants({ size: "lg", variant: "ghost" }),
								"px-6 gap-2 text-ink",
							)}
						>
							<span>{m.marketing_explore_features()}</span>
							<ArrowRight className="size-4" />
						</a>
					</div>
				</div>

				{/* Product Showcase Mockup */}
				<div className="relative mx-auto mt-14 max-w-3xl">
					<div className="border border-line bg-ground p-3 md:p-4 text-left">
						{/* Window Chrome */}
						<div className="flex items-center justify-between border-b border-line px-2 pb-2 text-xs text-gray font-mono">
							<div className="flex items-center gap-2">
								<span className="size-2 border border-line bg-transparent" />
								<span className="size-2 border border-line bg-transparent" />
								<span className="size-2 border border-line bg-transparent" />
							</div>
							<div className="type-kicker text-gray">app.beep.dev</div>
							<div className="w-8" />
						</div>

						{/* Mock Dashboard Body */}
						<div className="space-y-4 pt-4">
							<div className="border border-line p-4">
								<div className="type-kicker text-gray">
									{m.marketing_hero_mock_title()}
								</div>
								<div className="mt-3 flex flex-col gap-3 sm:flex-row sm:items-center">
									<div className="flex flex-1 items-center gap-2 border-b border-line py-2 text-sm text-ink">
										<span className="truncate">
											{m.marketing_hero_prompt_input()}
										</span>
									</div>
									<div className="inline-flex shrink-0 items-center justify-center gap-1.5 border border-solid bg-solid px-3.5 py-2 text-xs font-normal text-solid-ink">
										<span>{m.marketing_hero_mock_schedule()}</span>
									</div>
								</div>
								<div className="mt-2.5 flex items-center gap-2 type-meta text-gray text-xs">
									<span>{m.marketing_hero_channels_title()}</span>
									<span className="font-mono text-ink">
										{m.marketing_hero_dispatch_status()}
									</span>
								</div>
							</div>

							<div className="grid grid-cols-2 gap-3 sm:grid-cols-3">
								<div className="border border-line p-3">
									<span className="type-kicker text-gray block">
										{m.marketing_hero_mock_title()}
									</span>
									<span className="type-app-title font-medium text-ink block mt-1">
										24/7
									</span>
								</div>
								<div className="border border-line p-3">
									<span className="type-kicker text-gray block">
										{m.marketing_hero_live_status()}
									</span>
									<span className="type-app-title font-medium text-ink block mt-1">
										99.9%
									</span>
								</div>
								<div className="col-span-2 border border-line p-3 sm:col-span-1">
									<span className="type-kicker text-gray block">Status</span>
									<span className="text-sm text-ink block mt-1">
										{m.marketing_hero_dispatch_status()}
									</span>
								</div>
							</div>
						</div>
					</div>
				</div>
			</div>
		</section>
	);
}
