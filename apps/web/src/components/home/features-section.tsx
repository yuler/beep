import {
	Activity,
	BellRing,
	Clock,
	Code2,
	MessageSquareQuote,
	Users2,
} from "lucide-react";
import {
	Card,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { m } from "@/locale/paraglide/messages";

const features = [
	{
		id: "feature-nlp",
		icon: MessageSquareQuote,
		title: m.marketing_feature_1_title,
		description: m.marketing_feature_1_description,
	},
	{
		id: "feature-probes",
		icon: Activity,
		title: m.marketing_feature_2_title,
		description: m.marketing_feature_2_description,
	},
	{
		id: "feature-dispatch",
		icon: BellRing,
		title: m.marketing_feature_3_title,
		description: m.marketing_feature_3_description,
	},
	{
		id: "feature-cron",
		icon: Clock,
		title: m.marketing_feature_4_title,
		description: m.marketing_feature_4_description,
	},
	{
		id: "feature-tenancy",
		icon: Users2,
		title: m.marketing_feature_5_title,
		description: m.marketing_feature_5_description,
	},
	{
		id: "feature-api",
		icon: Code2,
		title: m.marketing_feature_6_title,
		description: m.marketing_feature_6_description,
	},
] as const;

export function FeaturesSection() {
	return (
		<section
			id="features"
			className="border-t border-line bg-ground py-16 md:py-24 text-ink"
		>
			<div className="mx-auto max-w-[1080px] px-4 md:px-6 lg:px-8">
				{/* Section Heading */}
				<div className="mx-auto max-w-2xl text-center">
					<h2 className="type-title text-2xl font-medium tracking-tight text-ink sm:text-3xl">
						{m.marketing_features_title()}
					</h2>
					<p className="type-meta mt-3 text-base text-gray">
						{m.marketing_features_description()}
					</p>
				</div>

				{/* Features Grid */}
				<div className="mt-12 grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
					{features.map(({ id, icon: Icon, title, description }) => (
						<Card
							key={id}
							className="border-t border-line sm:border sm:border-line bg-transparent p-5 hover:bg-wash transition-colors"
						>
							<CardHeader className="p-0">
								<div className="mb-3 inline-flex size-8 items-center justify-center border border-line text-ink">
									<Icon className="size-4" />
								</div>
								<CardTitle className="type-app-title text-base font-medium text-ink">
									{title()}
								</CardTitle>
								<CardDescription className="type-meta mt-2 text-sm text-gray leading-relaxed">
									{description()}
								</CardDescription>
							</CardHeader>
						</Card>
					))}
				</div>
			</div>
		</section>
	);
}
