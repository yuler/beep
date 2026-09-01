import { Sparkles } from "lucide-react";
import { SiteAuthButton } from "@/components/layout/site-auth-button";
import { m } from "@/locale/paraglide/messages";

export function CtaSection() {
	return (
		<section className="border-t border-border/80 bg-muted/20 py-20 md:py-28">
			<div className="mx-auto max-w-5xl px-4 sm:px-6">
				<div className="relative overflow-hidden rounded-2xl border border-primary/20 bg-gradient-to-b from-primary/10 via-primary/5 to-background p-8 text-center shadow-lg md:p-14">
					{/* Glow effect */}
					<div
						className="pointer-events-none absolute -top-24 left-1/2 -z-10 -translate-x-1/2 transform-gpu blur-2xl"
						aria-hidden="true"
					>
						<div className="aspect-square w-96 rounded-full bg-primary/20 opacity-70" />
					</div>

					<div className="mx-auto max-w-2xl">
						<div className="inline-flex size-10 items-center justify-center rounded-xl bg-primary/10 text-primary shadow-xs">
							<Sparkles className="size-5" />
						</div>

						<h2 className="font-heading mt-6 text-3xl font-bold tracking-tight text-foreground sm:text-4xl">
							{m.marketing_cta_title()}
						</h2>

						<p className="mt-4 text-base text-muted-foreground sm:text-lg">
							{m.marketing_cta_description()}
						</p>

						<div className="mt-8 flex justify-center">
							<SiteAuthButton
								signInLabel={m.common_get_started()}
								dashboardLabel={m.common_open_dashboard()}
								size="lg"
								className="h-11 px-8 shadow-md"
							/>
						</div>
					</div>
				</div>
			</div>
		</section>
	);
}
