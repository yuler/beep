import { Link } from "@tanstack/react-router";
import type { ReactNode } from "react";
import { LocaleSwitcher } from "@/components/layout/locale-switcher";
import { LogoMark } from "@/components/logo-mark";
import { ThemeToggle } from "@/components/theme-toggle";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
} from "@/components/ui/card";
import { cn } from "@/lib/utils";
import { m } from "@/locale/paraglide/messages";

/** Shown while client-only auth routes probe `session_id` on Core. */
export function AuthPending() {
	return (
		<div className="flex min-h-svh items-center justify-center bg-ground">
			<span className="is-loading-spin" aria-hidden="true" />
			<span className="sr-only">{m.auth_checking_session()}</span>
		</div>
	);
}

export function AuthLayout({ children }: { children: ReactNode }) {
	return (
		<div className="relative flex min-h-svh flex-col items-center justify-center bg-ground px-4 py-8 sm:px-6">
			<header className="fixed top-0 inset-x-0 z-50 h-14 md:h-16 border-b border-line bg-ground">
				<div className="mx-auto flex h-full max-w-[1080px] items-center justify-between px-4 md:px-6 lg:px-8">
					<Link
						to="/"
						className="type-kicker text-gray hover:text-ink transition-colors"
					>
						← {m.common_back()}
					</Link>
					<div className="flex items-center gap-1 sm:gap-2">
						<LocaleSwitcher />
						<ThemeToggle />
					</div>
				</div>
			</header>

			<div className="w-full max-w-[440px] pt-12">{children}</div>
		</div>
	);
}

export function AuthCard({
	description,
	children,
	className,
}: {
	description: string;
	children: ReactNode;
	className?: string;
}) {
	return (
		<Card
			className={cn(
				"border border-line bg-ground p-2 sm:p-4 text-ink shadow-none",
				className,
			)}
		>
			<CardHeader className="flex flex-col items-center gap-2 text-center pb-6">
				<Link
					to="/"
					aria-label={m.term_beep()}
					className="mb-2 inline-flex items-center gap-2.5 text-ink hover:opacity-80 transition-opacity"
				>
					<LogoMark className="size-6 text-ink" />
					<span className="type-kicker font-medium tracking-[0.18em] uppercase">
						{m.term_beep()}
					</span>
				</Link>
				<CardDescription className="type-meta text-gray">
					{description}
				</CardDescription>
			</CardHeader>
			<CardContent className="px-2 sm:px-4">{children}</CardContent>
		</Card>
	);
}
