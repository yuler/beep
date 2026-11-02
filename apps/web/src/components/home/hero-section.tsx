import { ArrowRight, Sparkles } from "lucide-react";
import { HeroBrowser } from "@/components/home/hero-browser";
import { SiteAuthButton } from "@/components/layout/site-auth-button";
import { buttonVariants } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { m } from "@/locale/paraglide/messages";

export function HeroSection() {
	return (
		<section className="relative overflow-hidden pt-16 pb-16 md:pt-24 md:pb-24 bg-ground text-ink">
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

				<HeroBrowser />
			</div>
		</section>
	);
}
