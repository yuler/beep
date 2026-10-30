import { m } from "@/locale/paraglide/messages";

export function SiteFooter() {
	return (
		<footer className="border-t border-line bg-ground">
			<div className="mx-auto flex max-w-[1080px] flex-col gap-2 px-4 py-8 type-meta text-gray sm:flex-row sm:items-center sm:justify-between md:px-6 lg:px-8">
				<p>{m.marketing_footer_tagline()}</p>
				<p className="type-kicker">{m.marketing_footer_subtitle()}</p>
			</div>
		</footer>
	);
}
