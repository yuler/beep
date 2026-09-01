import { createFileRoute } from "@tanstack/react-router";
import {
	CtaSection,
	FaqSection,
	FeaturesSection,
	HeroSection,
	HowItWorksSection,
	TestimonialsSection,
} from "@/components/home";
import { SiteLayout } from "@/components/layout";

export const Route = createFileRoute("/")({ component: Home });

function Home() {
	return (
		<SiteLayout>
			<main className="flex-1">
				<HeroSection />
				<FeaturesSection />
				<HowItWorksSection />
				<TestimonialsSection />
				<FaqSection />
				<CtaSection />
			</main>
		</SiteLayout>
	);
}
