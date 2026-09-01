import { Button } from "@/components/ui/button";
import { type Locale, useI18n } from "@/lib/i18n";
import * as m from "@/locale/paraglide/messages";

export function LanguageToggle() {
	const { locale, getLocalizedPath } = useI18n();

	const nextLocale: Locale = locale === "en" ? "zh" : "en";
	const label = locale === "en" ? "EN" : "中";

	return (
		<Button
			type="button"
			variant="outline"
			size="icon"
			className="text-xs font-semibold"
			aria-label={m.common_language()}
			onClick={() => {
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
			<span className="sr-only">{m.common_language()}</span>
			<span className="text-xs">{label}</span>
		</Button>
	);
}
