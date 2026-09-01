import { Link } from "@tanstack/react-router";
import { LocaleSwitcher } from "@/components/layout/locale-switcher";
import { SiteAuthButton } from "@/components/layout/site-auth-button";
import { LogoMark } from "@/components/logo-mark";
import { ThemeToggle } from "@/components/theme-toggle";
import { m } from "@/locale/paraglide/messages";

export function SiteHeader() {
	return (
		<header className="sticky top-0 z-50 border-b border-border/80 bg-background/80 backdrop-blur-sm">
			<div className="mx-auto flex max-w-5xl items-center justify-between gap-4 px-4 py-3 sm:px-6">
				<Link
					to="/"
					aria-label={m.term_beep()}
					className="inline-flex items-center gap-2 font-semibold tracking-tight text-foreground"
				>
					<LogoMark className="size-6" />
					<span>{m.term_beep()}</span>
				</Link>

				<div className="flex items-center gap-2">
					<LocaleSwitcher />
					<ThemeToggle />
					<SiteAuthButton />
				</div>
			</div>
		</header>
	);
}
