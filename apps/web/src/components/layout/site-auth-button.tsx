import { getRouteApi, Link, useNavigate } from "@tanstack/react-router";
import type { ComponentProps } from "react";

import { SignInDialog } from "@/components/auth/sign-in-dialog";
import { Button, buttonVariants } from "@/components/ui/button";
import { resolveDashboardTarget } from "@/lib/auth/account";
import { navigateForTarget } from "@/lib/auth/guards";
import { cn } from "@/lib/utils";

const rootRoute = getRouteApi("__root__");

type ButtonProps = ComponentProps<typeof Button>;

export function SiteAuthButton({
	signInLabel = "Sign in",
	dashboardLabel = "Dashboard",
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
	const navigate = useNavigate();
	const { me } = rootRoute.useRouteContext();
	const target =
		me && me.accounts.length > 0 ? resolveDashboardTarget(me.accounts) : null;

	if (!target || target.kind === "sign") {
		return (
			<SignInDialog
				label={signInLabel}
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
				{dashboardLabel}
			</Link>
		);
	}

	if (target.kind === "picker") {
		return (
			<Link
				to="/accounts"
				className={cn(buttonVariants({ size, variant }), className)}
			>
				{dashboardLabel}
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
			{dashboardLabel}
		</Button>
	);
}
