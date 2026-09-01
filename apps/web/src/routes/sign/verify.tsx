import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { useEffect, useState } from "react";
import { VerifyForm } from "@/components/auth/verify-form";
import { AuthCard } from "@/components/layout";
import { safeReturnTo } from "@/lib/auth/return-to";
import { m } from "@/locale/paraglide/messages";

export const Route = createFileRoute("/sign/verify")({
	component: VerifyPage,
});

function VerifyPage() {
	const navigate = useNavigate();
	const { return_to: returnTo } = Route.useSearch();
	const safe = safeReturnTo(returnTo);
	const [email, setEmail] = useState("");

	useEffect(() => {
		setEmail(sessionStorage.getItem("beep.sign_in_email") ?? "");
	}, []);

	const description = email
		? m.auth_verify_description_email({ email })
		: m.auth_verify_description();

	return (
		<AuthCard description={description}>
			<VerifyForm
				idPrefix="page-verify"
				returnTo={returnTo}
				onBack={() => {
					sessionStorage.removeItem("beep.sign_in_email");
					if (import.meta.env.DEV) {
						sessionStorage.removeItem("beep.dev_magic_link_code");
					}
					void navigate({
						to: "/sign",
						search: safe ? { return_to: safe } : {},
					});
				}}
			/>
		</AuthCard>
	);
}
