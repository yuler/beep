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
import { m } from "@/locale/paraglide/messages";

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
	const resolvedLabel = label ?? m.auth_sign_in();
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
						{step === "email" ? m.auth_sign_in() : m.auth_check_email()}
					</DialogTitle>
					<DialogDescription>
						{step === "email"
							? m.auth_sign_in_description()
							: email
								? m.auth_verify_description_email({ email })
								: m.auth_verify_description()}
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
						{m.auth_prefer_full_page()}{" "}
						<Link
							to="/sign"
							className="underline underline-offset-4"
							onClick={() => handleOpenChange(false)}
						>
							{m.auth_open_sign_in()}
						</Link>
					</p>
				) : null}
			</DialogContent>
		</Dialog>
	);
}
