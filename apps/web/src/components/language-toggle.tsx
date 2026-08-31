import { Button } from "@/components/ui/button";
import { type Locale, useI18n } from "@/lib/i18n";

export function LanguageToggle() {
	const { locale, setLocale, getLocalizedPath, t } = useI18n();

	const nextLocale: Locale = locale === "en" ? "zh" : "en";
	const label = locale === "en" ? "EN" : "中";

	return (
		<Button
			type="button"
			variant="outline"
			size="icon"
			className="text-xs font-semibold"
			aria-label={t("common.language")}
			onClick={() => {
				setLocale(nextLocale);
				if (typeof window !== "undefined") {
					const nextUrl = getLocalizedPath(
						window.location.pathname,
						nextLocale,
					);
					const search = window.location.search;
					window.location.href = `${nextUrl}${search}`;
				}
			}}
		>
			<span className="sr-only">{t("common.language")}</span>
			<span className="text-xs">{label}</span>
		</Button>
	);
}
