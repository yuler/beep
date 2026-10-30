import { Star } from "lucide-react";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Card } from "@/components/ui/card";
import { m } from "@/locale/paraglide/messages";

const STARS = [1, 2, 3, 4, 5] as const;

const testimonials = [
	{
		id: "yuler",
		quote: m.marketing_testimonial_1_quote,
		author: m.marketing_testimonial_1_author,
		role: m.marketing_testimonial_1_role,
		avatar: "/avatars/yuler.png",
		initials: "YU",
	},
	{
		id: "jianwu",
		quote: m.marketing_testimonial_2_quote,
		author: m.marketing_testimonial_2_author,
		role: m.marketing_testimonial_2_role,
		avatar: "/avatars/jianwu.png",
		initials: "JW",
	},
	{
		id: "gonghui",
		quote: m.marketing_testimonial_3_quote,
		author: m.marketing_testimonial_3_author,
		role: m.marketing_testimonial_3_role,
		avatar: "/avatars/gonghui.png",
		initials: "GH",
	},
] as const;

export function TestimonialsSection() {
	return (
		<section className="border-t border-line bg-ground py-16 md:py-24 text-ink">
			<div className="mx-auto max-w-[1080px] px-4 md:px-6 lg:px-8">
				{/* Section Heading */}
				<div className="mx-auto max-w-2xl text-center">
					<div className="type-kicker uppercase text-gray mb-2">
						{m.marketing_testimonials_badge()}
					</div>
					<h2 className="type-title text-2xl font-medium tracking-tight text-ink sm:text-3xl">
						{m.marketing_testimonials_title()}
					</h2>
					<p className="type-meta mt-3 text-base text-gray">
						{m.marketing_testimonials_description()}
					</p>
				</div>

				{/* Testimonial Cards */}
				<div className="mt-12 grid gap-4 md:grid-cols-3">
					{testimonials.map(({ id, quote, author, role, avatar, initials }) => (
						<Card
							key={id}
							className="border-t border-line sm:border sm:border-line bg-transparent p-5 flex flex-col justify-between"
						>
							<div>
								{/* Star Rating */}
								<div className="flex items-center gap-1 text-ink">
									{STARS.map((starNum) => (
										<Star
											key={`star-${id}-${starNum}`}
											className="size-3.5 fill-ink text-ink"
										/>
									))}
								</div>

								{/* Quote */}
								<p className="type-meta mt-4 text-sm text-ink leading-relaxed">
									"{quote()}"
								</p>
							</div>

							{/* Author Info */}
							<div className="mt-6 flex items-center gap-3 border-t border-line pt-4">
								<Avatar size="sm">
									<AvatarImage src={avatar} alt={author()} />
									<AvatarFallback className="text-xs font-normal">
										{initials}
									</AvatarFallback>
								</Avatar>
								<div className="text-left min-w-0">
									<div className="text-sm font-medium text-ink truncate">
										{author()}
									</div>
									<div className="type-meta text-xs text-gray truncate">
										{role()}
									</div>
								</div>
							</div>
						</Card>
					))}
				</div>
			</div>
		</section>
	);
}
