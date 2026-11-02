import { createFileRoute } from "@tanstack/react-router";
import {
	CtaSection,
	FaqSection,
	FeaturesSection,
	HeroSection,
	StorySection,
	TestimonialsSection,
} from "@/components/home";
import { SiteLayout } from "@/components/layout";

export const Route = createFileRoute("/")({ component: Home });

function Home() {
	return (
		<SiteLayout header="actions">
			<main className="flex-1">
				<HeroSection />
				<StorySection />
				<FeaturesSection />
				<TestimonialsSection />
				<FaqSection />
				<CtaSection />
			</main>
		</SiteLayout>
	);
}
