import { isRedirect, useNavigate, useRouter } from "@tanstack/react-router";
import {
	type FormEvent,
	useCallback,
	useEffect,
	useRef,
	useState,
} from "react";
import { Button } from "@/components/ui/button";
import { InputOTP } from "@/components/ui/input-otp";
import { Label } from "@/components/ui/label";
import { ApiError } from "@/lib/api/client";
import { fetchMe, verifyMagicLink } from "@/lib/api/session";
import { resolvePostAuthTarget } from "@/lib/auth/account";
import { navigateForTarget } from "@/lib/auth/guards";
import { m } from "@/locale/paraglide/messages";

const OTP_LENGTH = 6;

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
	const navigate = useNavigate();
	const router = useRouter();
	const [code, setCode] = useState("");
	const [error, setError] = useState<string | null>(null);
	const [pending, setPending] = useState(false);
	const pendingRef = useRef(false);
	const lastAutoSubmittedRef = useRef<string | null>(null);
	const navigateRef = useRef(navigate);
	const routerRef = useRef(router);
	const returnToRef = useRef(returnTo);
	const onVerifiedRef = useRef(onVerified);

	navigateRef.current = navigate;
	routerRef.current = router;
	returnToRef.current = returnTo;
	onVerifiedRef.current = onVerified;

	const submitCode = useCallback(async (rawCode: string) => {
		const trimmed = rawCode.trim();
		if (trimmed.length < OTP_LENGTH || pendingRef.current) return;

		pendingRef.current = true;
		setError(null);
		setPending(true);

		try {
			await verifyMagicLink(trimmed);
			sessionStorage.removeItem("beep.sign_in_email");
			if (import.meta.env.DEV) {
				sessionStorage.removeItem("beep.dev_magic_link_code");
			}

			const me = await fetchMe();
			await navigateForTarget(
				navigateRef.current,
				resolvePostAuthTarget(me.accounts, returnToRef.current),
			);
			await routerRef.current.invalidate();

			onVerifiedRef.current?.();
		} catch (err) {
			if (isRedirect(err)) throw err;
			setError(
				err instanceof ApiError ? err.message : m.errors_something_went_wrong(),
			);
		} finally {
			pendingRef.current = false;
			setPending(false);
		}
	}, []);

	// In dev, autofill the code that the server returned
	useEffect(() => {
		if (!import.meta.env.DEV) return;
		const devCode = sessionStorage.getItem("beep.dev_magic_link_code");
		if (devCode) setCode(devCode);
	}, []);

	// Auto-submit once the OTP reaches full length (typing, paste, or dev autofill).
	useEffect(() => {
		const trimmed = code.trim();
		if (trimmed.length < OTP_LENGTH) {
			lastAutoSubmittedRef.current = null;
			return;
		}
		if (lastAutoSubmittedRef.current === trimmed) return;
		lastAutoSubmittedRef.current = trimmed;
		void submitCode(trimmed);
	}, [code, submitCode]);

	function onSubmit(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();
		void submitCode(code);
	}

	const codeId = `${idPrefix}-code`;

	return (
		<form className="flex flex-col gap-4" onSubmit={onSubmit}>
			<div className="flex flex-col gap-2">
				<Label htmlFor={codeId}>{m.auth_one_time_code()}</Label>
				<InputOTP
					id={codeId}
					value={code}
					onChange={setCode}
					autoFocus
					disabled={pending}
				/>
				<p className="text-xs text-muted-foreground">{m.auth_code_hint()}</p>
			</div>
			{error ? (
				<p className="text-sm text-destructive" role="alert">
					{error}
				</p>
			) : null}
			<Button
				type="submit"
				disabled={pending || code.length < OTP_LENGTH}
				className="w-full"
			>
				{pending ? m.auth_verifying() : m.auth_verify()}
			</Button>
			{onBack ? (
				<Button
					type="button"
					variant="ghost"
					className="w-full"
					disabled={pending}
					onClick={onBack}
				>
					{m.auth_use_different_email()}
				</Button>
			) : null}
		</form>
	);
}
