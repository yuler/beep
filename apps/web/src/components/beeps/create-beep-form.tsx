import { type FormEvent, useState } from "react";

import { BeepChannelFields } from "@/components/beeps/beep-channel-fields";
import { BeepMarkdown } from "@/components/beeps/beep-markdown";
import { Button } from "@/components/ui/button";
import { DateTimePicker } from "@/components/ui/datetime-picker";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { createBeep } from "@/lib/api/beeps";
import { ApiError } from "@/lib/api/client";
import type { BeepChannel } from "@/lib/beep-channels";
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
	personal,
	onCreated,
}: {
	slug: string;
	personal: boolean;
	onCreated: () => Promise<void> | void;
}) {
	const [title, setTitle] = useState("");
	const [body, setBody] = useState("");
	const [preview, setPreview] = useState(false);
	const [runAt, setRunAt] = useState(defaultRunAt);
	const [channels, setChannels] = useState<BeepChannel[]>([
		"email",
		"web_push",
	]);
	const [error, setError] = useState<string | null>(null);
	const [pending, setPending] = useState(false);

	async function onSubmit(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();
		setError(null);

		if (personal && channels.length === 0) {
			setError("Select at least one channel.");
			return;
		}

		setPending(true);

		try {
			await createBeep(slug, {
				title: title.trim(),
				body: body.trim() || null,
				run_at: runAt.toISOString(),
				channels: personal ? channels : undefined,
			});
			setTitle("");
			setBody("");
			setPreview(false);
			setRunAt(defaultRunAt());
			setChannels(["email", "web_push"]);
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
			<DateTimePicker
				id="beep-run-at"
				value={runAt}
				onChange={setRunAt}
				disabled={pending}
			/>
			{personal ? (
				<BeepChannelFields
					channels={channels}
					onChange={setChannels}
					disabled={pending}
				/>
			) : null}
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
