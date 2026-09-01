import { ArrowRight, Bot, Cpu, Radio } from "lucide-react";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { m } from "@/locale/paraglide/messages";

const steps = [
	{
		number: m.marketing_step_1_number,
		icon: Bot,
		title: m.marketing_step_1_title,
		description: m.marketing_step_1_description,
	},
	{
		number: m.marketing_step_2_number,
		icon: Radio,
		title: m.marketing_step_2_title,
		description: m.marketing_step_2_description,
	},
	{
		number: m.marketing_step_3_number,
		icon: Cpu,
		title: m.marketing_step_3_title,
		description: m.marketing_step_3_description,
	},
] as const;

export function HowItWorksSection() {
	return (
		<section className="border-t border-border/80 bg-background py-20 md:py-28">
			<div className="mx-auto max-w-5xl px-4 sm:px-6">
				{/* Section Heading */}
				<div className="mx-auto max-w-2xl text-center">
					<Badge variant="outline" className="mb-3 font-medium">
						{m.marketing_how_it_works_badge()}
					</Badge>
					<h2 className="font-heading text-3xl font-bold tracking-tight text-foreground sm:text-4xl">
						{m.marketing_how_it_works_title()}
					</h2>
					<p className="mt-4 text-base text-muted-foreground sm:text-lg">
						{m.marketing_how_it_works_description()}
					</p>
				</div>

				{/* Step Cards */}
				<div className="mt-14 grid gap-6 md:grid-cols-3">
					{steps.map(({ number, icon: Icon, title, description }, index) => (
						<Card
							key={number()}
							className="relative flex flex-col justify-between overflow-hidden border-border bg-card p-6 shadow-xs transition-shadow hover:shadow-md"
						>
							<div>
								<div className="flex items-center justify-between">
									<div className="flex size-10 items-center justify-center rounded-xl bg-primary/10 text-primary">
										<Icon className="size-5" />
									</div>
									<span className="font-mono text-2xl font-bold text-muted-foreground/30">
										{number()}
									</span>
								</div>
								<h3 className="font-heading mt-6 text-lg font-semibold text-foreground">
									{title()}
								</h3>
								<p className="mt-2 text-sm leading-relaxed text-muted-foreground">
									{description()}
								</p>
							</div>

							{index < steps.length - 1 && (
								<div className="mt-6 hidden items-center gap-1 text-xs font-medium text-muted-foreground md:flex">
									<span>Next step</span>
									<ArrowRight className="size-3.5" />
								</div>
							)}
						</Card>
					))}
				</div>
			</div>
		</section>
	);
}
