import { Link } from "@tanstack/react-router";
import { LocaleSwitcher } from "@/components/layout/locale-switcher";
import { SiteAuthButton } from "@/components/layout/site-auth-button";
import { LogoMark } from "@/components/logo-mark";
import { ThemeToggle } from "@/components/theme-toggle";
import { m } from "@/locale/paraglide/messages";

export function SiteHeader() {
	return (
		<header className="sticky top-0 z-50 h-14 md:h-16 border-b border-line bg-ground">
			<div className="mx-auto flex h-full max-w-[1080px] items-center justify-between px-4 md:px-6 lg:px-8">
				<Link
					to="/"
					aria-label={m.term_beep()}
					className="inline-flex items-center gap-2.5 text-ink hover:opacity-80 transition-opacity"
				>
					<LogoMark className="size-5 text-ink" />
					<span className="type-kicker font-medium tracking-[0.18em] uppercase">
						{m.term_beep()}
					</span>
				</Link>

				<div className="flex items-center gap-1 sm:gap-2">
					<LocaleSwitcher />
					<SiteAuthButton variant="ghost" size="sm" />
					<ThemeToggle />
				</div>
			</div>
		</header>
	);
}
