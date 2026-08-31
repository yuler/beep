import { Link, useNavigate } from "@tanstack/react-router";
import type { ComponentProps } from "react";

import { SignInDialog } from "@/components/auth/sign-in-dialog";
import { Button, buttonVariants } from "@/components/ui/button";
import { resolveDashboardTarget } from "@/lib/auth/account";
import { navigateForTarget } from "@/lib/auth/guards";
import { useMe } from "@/lib/auth/use-me";
import { useTranslation } from "@/lib/i18n";
import { cn } from "@/lib/utils";

type ButtonProps = ComponentProps<typeof Button>;

export function SiteAuthButton({
	signInLabel,
	dashboardLabel,
	size,
	variant,
	className,
}: {
	signInLabel?: string;
	dashboardLabel?: string;
	size?: ButtonProps["size"];
	variant?: ButtonProps["variant"];
	className?: string;
}) {
	const { t } = useTranslation();
	const navigate = useNavigate();
	const { me } = useMe();
	const resolvedSignInLabel = signInLabel ?? t("auth.sign_in");
	const resolvedDashboardLabel = dashboardLabel ?? t("common.dashboard");
	const target =
		me && me.accounts.length > 0 ? resolveDashboardTarget(me.accounts) : null;

	if (!target || target.kind === "sign") {
		return (
			<SignInDialog
				label={resolvedSignInLabel}
				size={size}
				variant={variant}
				className={className}
			/>
		);
	}

	if (target.kind === "account") {
		return (
			<Link
				to="/$account_slug"
				params={{ account_slug: target.slug }}
				className={cn(buttonVariants({ size, variant }), className)}
			>
				{resolvedDashboardLabel}
			</Link>
		);
	}

	if (target.kind === "picker") {
		return (
			<Link
				to="/accounts"
				className={cn(buttonVariants({ size, variant }), className)}
			>
				{resolvedDashboardLabel}
			</Link>
		);
	}

	return (
		<Button
			size={size}
			variant={variant}
			className={className}
			onClick={() => {
				void navigateForTarget(navigate, target);
			}}
		>
			{resolvedDashboardLabel}
		</Button>
	);
}
