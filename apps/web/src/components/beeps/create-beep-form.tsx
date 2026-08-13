import { type FormEvent, useState } from "react";

import { Button } from "@/components/ui/button";
import { DateTimePicker } from "@/components/ui/datetime-picker";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { createBeep } from "@/lib/api/beeps";
import { ApiError } from "@/lib/api/client";

function defaultRunAt() {
	return new Date(Date.now() + 60 * 60 * 1000);
}

export function CreateBeepForm({
	slug,
	onCreated,
}: {
	slug: string;
	onCreated: () => Promise<void> | void;
}) {
	const [message, setMessage] = useState("");
	const [runAt, setRunAt] = useState(defaultRunAt);
	const [error, setError] = useState<string | null>(null);
	const [pending, setPending] = useState(false);

	async function onSubmit(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();
		setError(null);
		setPending(true);

		try {
			await createBeep(slug, {
				message: message.trim(),
				run_at: runAt.toISOString(),
			});
			setMessage("");
			setRunAt(defaultRunAt());
			await onCreated();
		} catch (err) {
			setError(err instanceof ApiError ? err.message : "Something went wrong.");
		} finally {
			setPending(false);
		}
	}

	return (
		<form className="flex flex-col gap-4" onSubmit={onSubmit}>
			<div className="flex flex-col gap-2">
				<Label htmlFor="beep-message">Message</Label>
				<Input
					id="beep-message"
					name="message"
					required
					maxLength={500}
					value={message}
					onChange={(event) => setMessage(event.target.value)}
					placeholder="Call mom"
				/>
			</div>
			<DateTimePicker
				id="beep-run-at"
				value={runAt}
				onChange={setRunAt}
				disabled={pending}
			/>
			{error ? (
				<p className="text-sm text-destructive" role="alert">
					{error}
				</p>
			) : null}
			<Button type="submit" disabled={pending} className="w-fit">
				{pending ? "Creating…" : "Create beep"}
			</Button>
		</form>
	);
}
