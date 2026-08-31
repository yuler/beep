import { Link } from "@tanstack/react-router";
import type { ComponentProps } from "react";
import { useState } from "react";

import { SignInForm } from "@/components/auth/sign-in-form";
import { VerifyForm } from "@/components/auth/verify-form";
import { Button } from "@/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogHeader,
	DialogTitle,
	DialogTrigger,
} from "@/components/ui/dialog";
import { useTranslation } from "@/lib/i18n";

type ButtonProps = ComponentProps<typeof Button>;
type Step = "email" | "verify";

export function SignInDialog({
	label,
	size,
	variant,
	className,
}: {
	label?: string;
	size?: ButtonProps["size"];
	variant?: ButtonProps["variant"];
	className?: string;
}) {
	const { t } = useTranslation();
	const resolvedLabel = label ?? t("auth.sign_in");
	const [open, setOpen] = useState(false);
	const [step, setStep] = useState<Step>("email");
	const [email, setEmail] = useState("");

	function reset() {
		setStep("email");
		setEmail("");
	}

	function handleOpenChange(next: boolean) {
		setOpen(next);
		if (!next) reset();
	}

	return (
		<Dialog open={open} onOpenChange={handleOpenChange} disablePointerDismissal>
			<DialogTrigger
				render={<Button size={size} variant={variant} className={className} />}
			>
				{resolvedLabel}
			</DialogTrigger>
			<DialogContent className="sm:max-w-md">
				<DialogHeader>
					<DialogTitle>
						{step === "email" ? t("auth.sign_in") : t("auth.check_email")}
					</DialogTitle>
					<DialogDescription>
						{step === "email"
							? t("auth.sign_in_description")
							: t("auth.verify_description")}
					</DialogDescription>
				</DialogHeader>
				{step === "email" ? (
					<SignInForm
						key={email || "email"}
						idPrefix="dialog-sign"
						initialEmail={email}
						stayInPlace
						onSuccess={({ email: nextEmail }) => {
							setEmail(nextEmail);
							setStep("verify");
						}}
					/>
				) : (
					<VerifyForm
						idPrefix="dialog-verify"
						onBack={() => {
							if (import.meta.env.DEV) {
								sessionStorage.removeItem("beep.dev_magic_link_code");
							}
							setStep("email");
						}}
						onVerified={() => setOpen(false)}
					/>
				)}
				{step === "email" ? (
					<p className="text-center text-xs text-muted-foreground">
						{t("auth.prefer_full_page")}{" "}
						<Link
							to="/sign"
							className="underline underline-offset-4"
							onClick={() => handleOpenChange(false)}
						>
							{t("auth.open_sign_in")}
						</Link>
					</p>
				) : null}
			</DialogContent>
		</Dialog>
	);
}
