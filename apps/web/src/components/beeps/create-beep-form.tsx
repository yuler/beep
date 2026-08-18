import { type FormEvent, useState } from "react";

import { Button } from "@/components/ui/button";
import { DateTimePicker } from "@/components/ui/datetime-picker";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { createBeep } from "@/lib/api/beeps";
import { ApiError } from "@/lib/api/client";
import { cn } from "@/lib/utils";

const TITLE_MAX_LENGTH = 80;
const BODY_MAX_LENGTH = 2000;

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
	const [title, setTitle] = useState("");
	const [body, setBody] = useState("");
	const [runAt, setRunAt] = useState(defaultRunAt);
	const [error, setError] = useState<string | null>(null);
	const [pending, setPending] = useState(false);

	async function onSubmit(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();
		setError(null);
		setPending(true);

		try {
			await createBeep(slug, {
				title: title.trim(),
				body: body.trim() || null,
				run_at: runAt.toISOString(),
			});
			setTitle("");
			setBody("");
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
				<Label htmlFor="beep-title">Title</Label>
				<Input
					id="beep-title"
					name="title"
					required
					maxLength={TITLE_MAX_LENGTH}
					value={title}
					onChange={(event) => setTitle(event.target.value)}
					placeholder="Call mom"
					disabled={pending}
				/>
			</div>
			<div className="flex flex-col gap-2">
				<Label htmlFor="beep-body">Body</Label>
				<textarea
					id="beep-body"
					name="body"
					maxLength={BODY_MAX_LENGTH}
					value={body}
					onChange={(event) => setBody(event.target.value)}
					placeholder="Bring **milk** and [eggs](https://example.com)"
					disabled={pending}
					rows={4}
					className={cn(
						"w-full min-w-0 rounded-lg border border-input bg-transparent px-2.5 py-1.5 text-base transition-colors outline-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:pointer-events-none disabled:cursor-not-allowed disabled:bg-input/50 disabled:opacity-50 md:text-sm dark:bg-input/30",
					)}
				/>
				<p className="text-xs text-muted-foreground">
					Optional. **bold**, _italic_, [links](https://…), lists, and line
					breaks.
				</p>
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
