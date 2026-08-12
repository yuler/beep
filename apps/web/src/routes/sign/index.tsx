import { createFileRoute } from "@tanstack/react-router";

import { SignInForm } from "@/components/auth/sign-in-form";
import { AuthCard } from "@/components/layout";

export const Route = createFileRoute("/sign/")({
	component: SignPage,
});

function SignPage() {
	const { return_to: returnTo } = Route.useSearch();

	return (
		<AuthCard description="Enter your email to sign in or create an account.">
			<SignInForm idPrefix="page-sign" returnTo={returnTo} />
		</AuthCard>
	);
}
