import { useTranslation } from "@/lib/i18n";

export function SiteFooter() {
	const { t } = useTranslation();

	return (
		<footer className="border-t border-border">
			<div className="mx-auto flex max-w-5xl flex-col gap-2 px-4 py-8 text-sm text-muted-foreground sm:flex-row sm:items-center sm:justify-between sm:px-6">
				<p>{t("marketing.footer_tagline")}</p>
				<p className="text-xs">{t("marketing.footer_subtitle")}</p>
			</div>
		</footer>
	);
}
