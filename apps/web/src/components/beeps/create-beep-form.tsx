import { type FormEvent, useState } from "react";

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { createBeep } from "@/lib/api/beeps";
import { ApiError } from "@/lib/api/client";

function toDatetimeLocalValue(date: Date) {
	const pad = (value: number) => String(value).padStart(2, "0");
	return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}T${pad(date.getHours())}:${pad(date.getMinutes())}`;
}

export function CreateBeepForm({
	slug,
	onCreated,
}: {
	slug: string;
	onCreated: () => Promise<void> | void;
}) {
	const [message, setMessage] = useState("");
	const [runAt, setRunAt] = useState(() =>
		toDatetimeLocalValue(new Date(Date.now() + 60 * 60 * 1000)),
	);
	const [error, setError] = useState<string | null>(null);
	const [pending, setPending] = useState(false);

	async function onSubmit(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();
		setError(null);
		setPending(true);

		try {
			await createBeep(slug, {
				message: message.trim(),
				run_at: new Date(runAt).toISOString(),
			});
			setMessage("");
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
			<div className="flex flex-col gap-2">
				<Label htmlFor="beep-run-at">When</Label>
				<Input
					id="beep-run-at"
					name="run_at"
					type="datetime-local"
					required
					value={runAt}
					onChange={(event) => setRunAt(event.target.value)}
				/>
			</div>
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
