import { m } from "@/locale/paraglide/messages";

export function SiteFooter() {
	return (
		<footer className="border-t border-border">
			<div className="mx-auto flex max-w-5xl flex-col gap-2 px-4 py-8 text-sm text-muted-foreground sm:flex-row sm:items-center sm:justify-between sm:px-6">
				<p>{m.marketing_footer_tagline()}</p>
				<p className="text-xs">{m.marketing_footer_subtitle()}</p>
			</div>
		</footer>
	);
}
