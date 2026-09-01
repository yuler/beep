import { Star } from "lucide-react";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Card } from "@/components/ui/card";
import { m } from "@/locale/paraglide/messages";

const STARS = [1, 2, 3, 4, 5] as const;

const testimonials = [
	{
		id: "alex",
		quote: m.marketing_testimonial_1_quote,
		author: m.marketing_testimonial_1_author,
		role: m.marketing_testimonial_1_role,
		initials: "AC",
	},
	{
		id: "sarah",
		quote: m.marketing_testimonial_2_quote,
		author: m.marketing_testimonial_2_author,
		role: m.marketing_testimonial_2_role,
		initials: "SL",
	},
	{
		id: "marcus",
		quote: m.marketing_testimonial_3_quote,
		author: m.marketing_testimonial_3_author,
		role: m.marketing_testimonial_3_role,
		initials: "MV",
	},
] as const;

export function TestimonialsSection() {
	return (
		<section className="border-t border-border/80 bg-muted/20 py-20 md:py-28">
			<div className="mx-auto max-w-5xl px-4 sm:px-6">
				{/* Section Heading */}
				<div className="mx-auto max-w-2xl text-center">
					<Badge variant="outline" className="mb-3 font-medium">
						{m.marketing_testimonials_badge()}
					</Badge>
					<h2 className="font-heading text-3xl font-bold tracking-tight text-foreground sm:text-4xl">
						{m.marketing_testimonials_title()}
					</h2>
					<p className="mt-4 text-base text-muted-foreground sm:text-lg">
						{m.marketing_testimonials_description()}
					</p>
				</div>

				{/* Testimonial Cards */}
				<div className="mt-14 grid gap-6 md:grid-cols-3">
					{testimonials.map(({ id, quote, author, role, initials }) => (
						<Card
							key={id}
							className="relative flex flex-col justify-between overflow-hidden border-border bg-card p-6 shadow-xs transition-shadow hover:shadow-md"
						>
							<div>
								{/* Star Rating */}
								<div className="flex items-center gap-1 text-amber-500">
									{STARS.map((starNum) => (
										<Star
											key={`star-${id}-${starNum}`}
											className="size-4 fill-amber-500"
										/>
									))}
								</div>

								{/* Quote */}
								<p className="mt-4 text-sm leading-relaxed text-foreground/90 italic">
									"{quote()}"
								</p>
							</div>

							{/* Author Info */}
							<div className="mt-6 flex items-center gap-3 border-t border-border/60 pt-4">
								<Avatar size="sm">
									<AvatarFallback className="bg-primary/10 text-xs font-semibold text-primary">
										{initials}
									</AvatarFallback>
								</Avatar>
								<div className="text-left">
									<div className="text-sm font-semibold text-foreground">
										{author()}
									</div>
									<div className="text-xs text-muted-foreground">{role()}</div>
								</div>
							</div>
						</Card>
					))}
				</div>
			</div>
		</section>
	);
}
