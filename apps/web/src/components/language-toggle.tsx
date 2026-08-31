import { Button } from "@/components/ui/button";
import { type Locale, useI18n } from "@/lib/i18n";

export function LanguageToggle() {
	const { locale, setLocaleUrl, t } = useI18n();

	const nextLocale: Locale = locale === "en" ? "zh-CN" : "en";
	const label = locale === "en" ? "EN" : "中";

	return (
		<Button
			type="button"
			variant="outline"
			size="icon"
			className="text-xs font-semibold"
			aria-label={t("common.language")}
			onClick={() => {
				const targetUrl = setLocaleUrl(nextLocale);
				window.location.href = targetUrl;
			}}
		>
			<span className="sr-only">{t("common.language")}</span>
			<span className="text-xs">{label}</span>
		</Button>
	);
}
