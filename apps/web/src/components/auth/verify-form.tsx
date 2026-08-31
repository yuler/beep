import { isRedirect, useNavigate, useRouter } from "@tanstack/react-router";
import { type FormEvent, useEffect, useState } from "react";

import { Button } from "@/components/ui/button";
import { InputOTP } from "@/components/ui/input-otp";
import { Label } from "@/components/ui/label";
import { ApiError } from "@/lib/api/client";
import { fetchMe, verifyMagicLink } from "@/lib/api/session";
import { resolvePostAuthTarget } from "@/lib/auth/account";
import { navigateForTarget } from "@/lib/auth/guards";
import { useTranslation } from "@/lib/i18n";

export function VerifyForm({
	idPrefix = "verify",
	returnTo,
	onBack,
	onVerified,
}: {
	idPrefix?: string;
	returnTo?: string;
	/** Go back to the email step (wrong address, etc.). */
	onBack?: () => void;
	/** Called after verify, session fetch, and post-auth navigation succeed. */
	onVerified?: () => void;
}) {
	const { t } = useTranslation();
	const navigate = useNavigate();
	const router = useRouter();
	const [code, setCode] = useState("");
	const [error, setError] = useState<string | null>(null);
	const [pending, setPending] = useState(false);

	// In dev, autofill the code that the server returned
	useEffect(() => {
		if (import.meta.env.DEV) {
			const devCode = sessionStorage.getItem("beep.dev_magic_link_code");
			if (devCode) setCode(devCode);
		}
	}, []);

	async function onSubmit(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();
		setError(null);
		setPending(true);

		try {
			await verifyMagicLink(code.trim());
			if (import.meta.env.DEV) {
				sessionStorage.removeItem("beep.dev_magic_link_code");
			}

			const me = await fetchMe();
			await navigateForTarget(
				navigate,
				resolvePostAuthTarget(me.accounts, returnTo),
			);
			await router.invalidate();

			onVerified?.();
		} catch (err) {
			if (isRedirect(err)) throw err;
			setError(
				err instanceof ApiError ? err.message : t("errors.something_went_wrong"),
			);
		} finally {
			setPending(false);
		}
	}

	const codeId = `${idPrefix}-code`;

	return (
		<form className="flex flex-col gap-4" onSubmit={onSubmit}>
			<div className="flex flex-col gap-2">
				<Label htmlFor={codeId}>{t("auth.one_time_code")}</Label>
				<InputOTP id={codeId} value={code} onChange={setCode} autoFocus />
				<p className="text-xs text-muted-foreground">
					{t("auth.code_hint")}
				</p>
			</div>
			{error ? (
				<p className="text-sm text-destructive" role="alert">
					{error}
				</p>
			) : null}
			<Button
				type="submit"
				disabled={pending || code.length < 6}
				className="w-full"
			>
				{pending ? t("auth.verifying") : t("auth.verify")}
			</Button>
			{onBack ? (
				<Button
					type="button"
					variant="ghost"
					className="w-full"
					disabled={pending}
					onClick={onBack}
				>
					{t("auth.use_different_email")}
				</Button>
			) : null}
		</form>
	);
}
