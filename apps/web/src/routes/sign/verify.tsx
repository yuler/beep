import { createFileRoute, useNavigate } from "@tanstack/react-router";
import { VerifyForm } from "@/components/auth/verify-form";
import { AuthCard } from "@/components/layout";
import { safeReturnTo } from "@/lib/auth/return-to";
import * as m from "@/locale/paraglide/messages";

export const Route = createFileRoute("/sign/verify")({
	component: VerifyPage,
});

function VerifyPage() {
	const navigate = useNavigate();
	const { return_to: returnTo } = Route.useSearch();
	const safe = safeReturnTo(returnTo);

	return (
		<AuthCard description={m.auth_verify_description()}>
			<VerifyForm
				idPrefix="page-verify"
				returnTo={returnTo}
				onBack={() => {
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
