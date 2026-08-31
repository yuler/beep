import { createFileRoute } from "@tanstack/react-router";

import { SignInForm } from "@/components/auth/sign-in-form";
import { AuthCard } from "@/components/layout";
import { useTranslation } from "@/lib/i18n";

export const Route = createFileRoute("/sign/")({
	component: SignPage,
});

function SignPage() {
	const { t } = useTranslation();
	const { return_to: returnTo } = Route.useSearch();

	return (
		<AuthCard description={t("auth.sign_in_description")}>
			<SignInForm idPrefix="page-sign" returnTo={returnTo} />
		</AuthCard>
	);
}
