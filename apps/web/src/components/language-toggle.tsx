import { Languages } from "lucide-react";
import { Button } from "@/components/ui/button";
import { type Locale, useI18n } from "@/lib/i18n";

const LOCALE_NAMES: Record<Locale, string> = {
	en: "English",
	"zh-CN": "简体中文",
};

export function LanguageToggle() {
	const { locale, setLocaleUrl, t } = useI18n();

	const nextLocale: Locale = locale === "en" ? "zh-CN" : "en";

	return (
		<Button
			type="button"
			variant="outline"
			size="sm"
			className="gap-1.5 text-xs font-medium"
			aria-label={t("common.language")}
			onClick={() => {
				const targetUrl = setLocaleUrl(nextLocale);
				window.location.href = targetUrl;
			}}
		>
			<Languages className="size-3.5" />
			<span>{LOCALE_NAMES[locale]}</span>
		</Button>
	);
}
