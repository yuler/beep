import { Link } from "@tanstack/react-router";
import { Loader2 } from "lucide-react";
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
		<div className="flex min-h-svh items-center justify-center bg-background">
			<Loader2
				className="size-6 animate-spin text-muted-foreground"
				aria-label={m.auth_checking_session()}
			/>
		</div>
	);
}

export function AuthLayout({ children }: { children: ReactNode }) {
	return (
		<div className="relative flex min-h-svh flex-col items-center justify-center overflow-hidden bg-background p-6 md:p-10">
			{/* Subtle background ambient glow */}
			<div
				className="pointer-events-none absolute inset-x-0 top-0 -z-10 flex transform-gpu justify-center overflow-hidden blur-3xl"
				aria-hidden="true"
			>
				<div className="aspect-[1155/678] w-[72.1875rem] flex-none bg-gradient-to-tr from-primary/20 via-primary/5 to-transparent opacity-60 dark:opacity-30" />
			</div>

			<div className="absolute top-6 left-6 flex items-center gap-4">
				<Link
					to="/"
					className="text-sm text-muted-foreground transition-colors hover:text-foreground"
				>
					{m.common_back()}
				</Link>
			</div>
			<div className="absolute top-6 right-6 flex items-center gap-2">
				<LocaleSwitcher />
				<ThemeToggle />
			</div>
			<div className="flex w-full max-w-sm flex-col gap-6">{children}</div>
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
				"border border-border/80 bg-card/90 shadow-sm backdrop-blur-xs",
				className,
			)}
		>
			<CardHeader className="flex flex-col items-center gap-2 text-center">
				<Link
					to="/"
					aria-label={m.term_beep()}
					className="mb-1 inline-flex items-center gap-2 font-semibold tracking-tight text-foreground"
				>
					<LogoMark className="size-6" />
					<span>{m.term_beep()}</span>
				</Link>
				<CardDescription>{description}</CardDescription>
			</CardHeader>
			<CardContent>{children}</CardContent>
		</Card>
	);
}
