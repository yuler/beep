import { Link } from "@tanstack/react-router";
import { LanguageToggle } from "@/components/language-toggle";
import { SiteAuthButton } from "@/components/layout/site-auth-button";
import { LogoMark } from "@/components/logo-mark";
import { ThemeToggle } from "@/components/theme-toggle";
import { useTranslation } from "@/lib/i18n";

export function SiteHeader() {
	const { t } = useTranslation();

	return (
		<header className="sticky top-0 z-50 border-b border-border/80 bg-background/80 backdrop-blur-sm">
			<div className="mx-auto flex max-w-5xl items-center justify-between gap-4 px-4 py-3 sm:px-6">
				<Link
					to="/"
					aria-label={t("term.beep")}
					className="inline-flex items-center gap-2 font-semibold tracking-tight text-foreground"
				>
					<span className="inline-flex size-8 items-center justify-center rounded-lg bg-foreground text-background dark:bg-foreground dark:text-background">
						<LogoMark className="size-5" />
					</span>
					<span>{t("term.beep")}</span>
				</Link>

				<div className="flex items-center gap-2">
					<LanguageToggle />
					<ThemeToggle />
					<SiteAuthButton />
				</div>
			</div>
		</header>
	);
}
