import { ArrowRight, Bot, Cpu, Radio } from "lucide-react";
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
		<section className="border-t border-line bg-ground py-16 md:py-24 text-ink">
			<div className="mx-auto max-w-[1080px] px-4 md:px-6 lg:px-8">
				{/* Section Heading */}
				<div className="mx-auto max-w-2xl text-center">
					<div className="type-kicker uppercase text-gray mb-2">
						{m.marketing_how_it_works_badge()}
					</div>
					<h2 className="type-title text-2xl font-medium tracking-tight text-ink sm:text-3xl">
						{m.marketing_how_it_works_title()}
					</h2>
					<p className="type-meta mt-3 text-base text-gray">
						{m.marketing_how_it_works_description()}
					</p>
				</div>

				{/* Step Cards */}
				<div className="mt-12 grid gap-4 md:grid-cols-3">
					{steps.map(({ number, icon: Icon, title, description }, index) => (
						<Card
							key={number()}
							className="border-t border-line sm:border sm:border-line bg-transparent p-5 flex flex-col justify-between"
						>
							<div>
								<div className="flex items-center justify-between">
									<div className="flex size-8 items-center justify-center border border-line text-ink">
										<Icon className="size-4" />
									</div>
									<span className="font-mono text-xl font-normal text-gray">
										{number()}
									</span>
								</div>
								<h3 className="type-app-title mt-5 text-base font-medium text-ink">
									{title()}
								</h3>
								<p className="type-meta mt-2 text-sm text-gray leading-relaxed">
									{description()}
								</p>
							</div>

							{index < steps.length - 1 && (
								<div className="mt-5 hidden items-center gap-1.5 type-kicker text-gray md:flex">
									<span>{m.marketing_how_it_works_next_step()}</span>
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
