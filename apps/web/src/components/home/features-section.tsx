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
		color: "text-primary bg-primary/10",
	},
	{
		id: "feature-probes",
		icon: Activity,
		title: m.marketing_feature_2_title,
		description: m.marketing_feature_2_description,
		color: "text-emerald-600 dark:text-emerald-400 bg-emerald-500/10",
	},
	{
		id: "feature-dispatch",
		icon: BellRing,
		title: m.marketing_feature_3_title,
		description: m.marketing_feature_3_description,
		color: "text-amber-600 dark:text-amber-400 bg-amber-500/10",
	},
	{
		id: "feature-cron",
		icon: Clock,
		title: m.marketing_feature_4_title,
		description: m.marketing_feature_4_description,
		color: "text-sky-600 dark:text-sky-400 bg-sky-500/10",
	},
	{
		id: "feature-tenancy",
		icon: Users2,
		title: m.marketing_feature_5_title,
		description: m.marketing_feature_5_description,
		color: "text-violet-600 dark:text-violet-400 bg-violet-500/10",
	},
	{
		id: "feature-api",
		icon: Code2,
		title: m.marketing_feature_6_title,
		description: m.marketing_feature_6_description,
		color: "text-rose-600 dark:text-rose-400 bg-rose-500/10",
	},
] as const;

export function FeaturesSection() {
	return (
		<section
			id="features"
			className="border-t border-border/80 bg-muted/20 py-20 md:py-28"
		>
			<div className="mx-auto max-w-5xl px-4 sm:px-6">
				{/* Section Heading */}
				<div className="mx-auto max-w-2xl text-center">
					<h2 className="font-heading text-3xl font-bold tracking-tight text-foreground sm:text-4xl">
						{m.marketing_features_title()}
					</h2>
					<p className="mt-4 text-base text-muted-foreground sm:text-lg">
						{m.marketing_features_description()}
					</p>
				</div>

				{/* Features Grid */}
				<div className="mt-14 grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
					{features.map(({ id, icon: Icon, title, description, color }) => (
						<Card
							key={id}
							className="relative overflow-hidden transition-all duration-200 hover:-translate-y-0.5 hover:shadow-md hover:ring-primary/30"
						>
							<CardHeader>
								<div
									className={`mb-3 inline-flex size-10 items-center justify-center rounded-xl ${color}`}
								>
									<Icon className="size-5" />
								</div>
								<CardTitle className="text-lg">{title()}</CardTitle>
								<CardDescription className="mt-2 text-sm leading-relaxed">
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
