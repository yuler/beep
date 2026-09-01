import { ChevronDown } from "lucide-react";
import { useState } from "react";
import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import { m } from "@/locale/paraglide/messages";

const faqs = [
	{
		q: m.marketing_faq_q1,
		a: m.marketing_faq_a1,
	},
	{
		q: m.marketing_faq_q2,
		a: m.marketing_faq_a2,
	},
	{
		q: m.marketing_faq_q3,
		a: m.marketing_faq_a3,
	},
	{
		q: m.marketing_faq_q4,
		a: m.marketing_faq_a4,
	},
	{
		q: m.marketing_faq_q5,
		a: m.marketing_faq_a5,
	},
] as const;

export function FaqSection() {
	const [openIndex, setOpenIndex] = useState<number | null>(0);

	function toggleFaq(index: number) {
		setOpenIndex((current) => (current === index ? null : index));
	}

	return (
		<section className="border-t border-border/80 bg-background py-20 md:py-28">
			<div className="mx-auto max-w-3xl px-4 sm:px-6">
				{/* Section Heading */}
				<div className="mx-auto max-w-2xl text-center">
					<Badge variant="outline" className="mb-3 font-medium">
						{m.marketing_faq_badge()}
					</Badge>
					<h2 className="font-heading text-3xl font-bold tracking-tight text-foreground sm:text-4xl">
						{m.marketing_faq_title()}
					</h2>
					<p className="mt-4 text-base text-muted-foreground sm:text-lg">
						{m.marketing_faq_description()}
					</p>
				</div>

				{/* FAQ Accordion List */}
				<div className="mt-12 divide-y divide-border/60 rounded-xl border border-border bg-card shadow-xs">
					{faqs.map(({ q, a }, index) => {
						const isOpen = openIndex === index;
						const questionId = `faq-question-${index}`;
						const answerId = `faq-answer-${index}`;
						return (
							<div key={q()} className="group">
								<button
									type="button"
									id={questionId}
									onClick={() => toggleFaq(index)}
									className="flex w-full items-center justify-between gap-4 p-5 text-left transition-colors hover:bg-muted/40"
									aria-expanded={isOpen}
									aria-controls={answerId}
								>
									<span className="font-heading text-base font-medium text-foreground">
										{q()}
									</span>
									<ChevronDown
										className={cn(
											"size-4 shrink-0 text-muted-foreground transition-transform duration-200",
											isOpen && "rotate-180 text-primary",
										)}
										aria-hidden="true"
									/>
								</button>
								{isOpen && (
									<div
										id={answerId}
										role="region"
										aria-labelledby={questionId}
										className="px-5 pb-5 pt-1 text-sm leading-relaxed text-muted-foreground"
									>
										{a()}
									</div>
								)}
							</div>
						);
					})}
				</div>
			</div>
		</section>
	);
}
