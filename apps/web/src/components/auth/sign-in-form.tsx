import { useNavigate } from "@tanstack/react-router";
import { type FormEvent, useState } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Notice } from "@/components/ui/notice";
import { ApiError } from "@/lib/api/client";
import { startSession } from "@/lib/api/session";
import { safeReturnTo } from "@/lib/auth/return-to";
import { m } from "@/locale/paraglide/messages";

export function SignInForm({
	idPrefix = "sign",
	returnTo,
	initialEmail = "",
	/** When true, stay put and call onSuccess (dialog step). Default navigates to /sign/verify. */
	stayInPlace = false,
	onSuccess,
}: {
	idPrefix?: string;
	returnTo?: string;
	initialEmail?: string;
	stayInPlace?: boolean;
	onSuccess?: (info: { email: string }) => void;
}) {
	const navigate = useNavigate();
	const [email, setEmail] = useState(initialEmail);
	const [error, setError] = useState<string | null>(null);
	const [pending, setPending] = useState(false);

	async function onSubmit(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();
		setError(null);
		setPending(true);

		const trimmed = email.trim();

		try {
			const result = await startSession(trimmed);
			if (import.meta.env.DEV && result.code) {
				sessionStorage.setItem("beep.dev_magic_link_code", result.code);
			}
			onSuccess?.({ email: trimmed });
			if (!stayInPlace) {
				sessionStorage.setItem("beep.sign_in_email", trimmed);
				const safe = safeReturnTo(returnTo);
				await navigate({
					to: "/sign/verify",
					search: safe ? { return_to: safe } : {},
				});
			}
		} catch (err) {
			setError(
				err instanceof ApiError ? err.message : m.errors_something_went_wrong(),
			);
		} finally {
			setPending(false);
		}
	}

	const emailId = `${idPrefix}-email`;

	return (
		<form className="flex flex-col gap-5" onSubmit={onSubmit}>
			<div className="flex flex-col gap-2">
				<Label htmlFor={emailId} className="type-kicker text-gray">
					{m.auth_email()}
				</Label>
				<Input
					id={emailId}
					name="email"
					type="email"
					autoComplete="username"
					required
					value={email}
					onChange={(event) => setEmail(event.target.value)}
					placeholder={m.auth_email_placeholder()}
				/>
			</div>
			{error ? (
				<Notice error label="出错" role="alert">
					{error}
				</Notice>
			) : null}
			<Button
				type="submit"
				variant="solid"
				loading={pending}
				loadingText={m.auth_sending()}
				className="w-full"
			>
				{m.auth_continue()}
			</Button>
		</form>
	);
}
