import { createFileRoute } from "@tanstack/react-router";
import { SignInForm } from "@/components/auth/sign-in-form";
import { AuthCard } from "@/components/layout";
import * as m from "@/locale/paraglide/messages";

export const Route = createFileRoute("/sign/")({
	component: SignPage,
});

function SignPage() {
	const { return_to: returnTo } = Route.useSearch();

	return (
		<AuthCard description={m.auth_sign_in_description()}>
			<SignInForm idPrefix="page-sign" returnTo={returnTo} />
		</AuthCard>
	);
}
