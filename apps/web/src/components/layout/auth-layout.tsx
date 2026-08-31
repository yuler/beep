import { Link } from "@tanstack/react-router";
import { Loader2 } from "lucide-react";
import type { ReactNode } from "react";

import { LanguageToggle } from "@/components/language-toggle";
import { LogoMark } from "@/components/logo-mark";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
} from "@/components/ui/card";
import { useTranslation } from "@/lib/i18n";
import { cn } from "@/lib/utils";

/** Shown while client-only auth routes probe `session_id` on Core. */
export function AuthPending() {
	return (
		<div className="flex min-h-svh items-center justify-center bg-background">
			<Loader2
				className="size-6 animate-spin text-muted-foreground"
				aria-label="Checking session"
			/>
		</div>
	);
}

export function AuthLayout({ children }: { children: ReactNode }) {
	const { t } = useTranslation();

	return (
		<div className="flex min-h-svh flex-col items-center justify-center gap-6 bg-background p-6 md:p-10">
			<div className="absolute top-6 left-6 flex items-center gap-4">
				<Link
					to="/"
					className="text-sm text-muted-foreground hover:text-foreground"
				>
					{t("common.back")}
				</Link>
			</div>
			<div className="absolute top-6 right-6">
				<LanguageToggle />
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
		<Card className={cn("border border-border shadow-xs", className)}>
			<CardHeader className="flex flex-col items-center gap-2 text-center">
				<Link to="/" aria-label="beep" className="mb-1 text-foreground">
					<span className="inline-flex size-10 items-center justify-center rounded-lg bg-foreground text-background">
						<LogoMark className="size-6" />
					</span>
				</Link>
				<CardDescription>{description}</CardDescription>
			</CardHeader>
			<CardContent>{children}</CardContent>
		</Card>
	);
}
