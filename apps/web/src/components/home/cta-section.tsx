import { Sparkles } from "lucide-react";
import { SiteAuthButton } from "@/components/layout/site-auth-button";
import { m } from "@/locale/paraglide/messages";

export function CtaSection() {
	return (
		<section className="border-t border-line bg-ground py-16 md:py-24 text-ink">
			<div className="mx-auto max-w-[1080px] px-4 md:px-6 lg:px-8">
				<div className="border border-line bg-ground p-8 text-center md:p-14">
					<div className="mx-auto max-w-2xl">
						<div className="inline-flex size-8 items-center justify-center border border-line text-ink mb-4">
							<Sparkles className="size-4" />
						</div>

						<h2 className="type-title text-2xl font-medium tracking-tight text-ink sm:text-3xl">
							{m.marketing_cta_title()}
						</h2>

						<p className="type-meta mt-3 text-base text-gray max-w-lg mx-auto">
							{m.marketing_cta_description()}
						</p>

						<div className="mt-8 flex justify-center">
							<SiteAuthButton
								signInLabel={m.common_get_started()}
								dashboardLabel={m.common_open_dashboard()}
								variant="solid"
								size="lg"
								className="px-8"
							/>
						</div>
					</div>
				</div>
			</div>
		</section>
	);
}
