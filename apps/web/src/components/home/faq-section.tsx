import { ChevronDown } from "lucide-react";
import { useState } from "react";
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
		<section className="border-t border-line bg-ground py-16 md:py-24 text-ink">
			<div className="mx-auto max-w-3xl px-4 md:px-6 lg:px-8">
				{/* Section Heading */}
				<div className="mx-auto max-w-2xl text-center">
					<div className="type-kicker uppercase text-gray mb-2">
						{m.marketing_faq_badge()}
					</div>
					<h2 className="type-title text-2xl font-medium tracking-tight text-ink sm:text-3xl">
						{m.marketing_faq_title()}
					</h2>
					<p className="type-meta mt-3 text-base text-gray">
						{m.marketing_faq_description()}
					</p>
				</div>

				{/* FAQ Accordion List */}
				<div className="mt-12 border-t border-line divide-y divide-line bg-ground">
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
									className="flex w-full items-center justify-between gap-4 py-4 text-left transition-colors hover:bg-wash px-2 cursor-pointer"
									aria-expanded={isOpen}
									aria-controls={answerId}
								>
									<span className="text-base font-normal text-ink">{q()}</span>
									<ChevronDown
										className={cn(
											"size-4 shrink-0 text-gray transition-transform duration-150",
											isOpen && "rotate-180 text-ink",
										)}
										aria-hidden="true"
									/>
								</button>
								{isOpen && (
									<section
										id={answerId}
										aria-labelledby={questionId}
										className="px-2 pb-4 pt-1 type-meta text-gray leading-relaxed text-sm"
									>
										{a()}
									</section>
								)}
							</div>
						);
					})}
				</div>
			</div>
		</section>
	);
}
