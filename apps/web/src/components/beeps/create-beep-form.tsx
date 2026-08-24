import { type FormEvent, useState } from "react";

import { BeepMarkdown } from "@/components/beeps/beep-markdown";
import { Button } from "@/components/ui/button";
import { DateTimePicker } from "@/components/ui/datetime-picker";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { createBeep } from "@/lib/api/beeps";
import { ApiError } from "@/lib/api/client";
import { cn } from "@/lib/utils";

const TITLE_MAX_LENGTH = 80;
const BODY_MAX_LENGTH = 2000;
const BODY_PLACEHOLDER = [
	"**bold**",
	"_italic_",
	"[link](https://…)",
	"- list",
	"line breaks",
	"# headings not supported",
].join("\n");

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
	const [preview, setPreview] = useState(false);
	const [kind, setKind] = useState<"once" | "imminent" | "recurring">("once");
	const [cron, setCron] = useState("0 9 * * *");
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
				kind,
				run_at: kind === "once" ? runAt.toISOString() : null,
				cron: kind === "recurring" ? cron.trim() : null,
			});
			setTitle("");
			setBody("");
			setPreview(false);
			setKind("once");
			setRunAt(defaultRunAt());
			setCron("0 9 * * *");
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
				<Label>Mode</Label>
				<div className="flex rounded-lg border border-input p-1">
					<Button
						type="button"
						size="sm"
						variant={kind === "once" ? "secondary" : "ghost"}
						className="flex-1"
						disabled={pending}
						onClick={() => setKind("once")}
					>
						Once
					</Button>
					<Button
						type="button"
						size="sm"
						variant={kind === "imminent" ? "secondary" : "ghost"}
						className="flex-1"
						disabled={pending}
						onClick={() => setKind("imminent")}
					>
						Imminent (Now)
					</Button>
					<Button
						type="button"
						size="sm"
						variant={kind === "recurring" ? "secondary" : "ghost"}
						className="flex-1"
						disabled={pending}
						onClick={() => setKind("recurring")}
					>
						Recurring
					</Button>
				</div>
			</div>

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
				<div className="flex items-center justify-between gap-2">
					<Label htmlFor="beep-body">Body</Label>
					<div className="flex gap-1">
						<Button
							type="button"
							size="xs"
							variant={preview ? "ghost" : "secondary"}
							aria-pressed={!preview}
							onClick={() => setPreview(false)}
							disabled={pending}
						>
							Write
						</Button>
						<Button
							type="button"
							size="xs"
							variant={preview ? "secondary" : "ghost"}
							aria-pressed={preview}
							onClick={() => setPreview(true)}
							disabled={pending}
						>
							Preview
						</Button>
					</div>
				</div>
				{preview ? (
					<div className="min-h-24 rounded-lg border border-input px-2.5 py-1.5 dark:bg-input/30">
						{body.trim() ? (
							<BeepMarkdown source={body} />
						) : (
							<p className="text-sm text-muted-foreground">
								Nothing to preview
							</p>
						)}
					</div>
				) : (
					<textarea
						id="beep-body"
						name="body"
						maxLength={BODY_MAX_LENGTH}
						value={body}
						onChange={(event) => setBody(event.target.value)}
						placeholder={BODY_PLACEHOLDER}
						disabled={pending}
						rows={6}
						className={cn(
							"w-full min-w-0 rounded-lg border border-input bg-transparent px-2.5 py-1.5 text-base transition-colors outline-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:pointer-events-none disabled:cursor-not-allowed disabled:bg-input/50 disabled:opacity-50 md:text-sm dark:bg-input/30",
						)}
					/>
				)}
			</div>

			{kind === "once" ? (
				<DateTimePicker
					id="beep-run-at"
					value={runAt}
					onChange={setRunAt}
					disabled={pending}
				/>
			) : null}

			{kind === "recurring" ? (
				<div className="flex flex-col gap-2">
					<Label htmlFor="beep-cron">Cron expression</Label>
					<Input
						id="beep-cron"
						name="cron"
						required
						value={cron}
						onChange={(event) => setCron(event.target.value)}
						placeholder="0 9 * * *"
						disabled={pending}
					/>
					<p className="text-xs text-muted-foreground">
						Standard cron format: minute hour day month day-of-week
					</p>
				</div>
			) : null}

			{kind === "imminent" ? (
				<p className="text-xs text-muted-foreground">
					This beep will be dispatched immediately to all enabled channels upon
					creation.
				</p>
			) : null}

			{error ? (
				<p className="text-sm text-destructive" role="alert">
					{error}
				</p>
			) : null}
			<Button type="submit" disabled={pending} className="w-fit">
				{pending
					? kind === "imminent"
						? "Sending…"
						: "Creating…"
					: kind === "imminent"
						? "Send now"
						: "Create beep"}
			</Button>
		</form>
	);
}
