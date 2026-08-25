import { HelpCircle } from "lucide-react";
import { type FormEvent, useState } from "react";

import { BeepMarkdown } from "@/components/beeps/beep-markdown";
import { Button } from "@/components/ui/button";
import { DateTimePicker } from "@/components/ui/datetime-picker";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
	Tooltip,
	TooltipContent,
	TooltipTrigger,
} from "@/components/ui/tooltip";
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
	const [kind, setKind] = useState<"once" | "recurring">("once");
	const [sendNow, setSendNow] = useState(true);
	const [cron, setCron] = useState("0 9 * * *");
	const [runAt, setRunAt] = useState(defaultRunAt);
	const [fieldErrors, setFieldErrors] = useState<{
		run_at?: string;
	}>({});
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
				run_at: kind === "once" && !sendNow ? runAt.toISOString() : null,
				cron: kind === "recurring" ? cron.trim() : null,
			});
			setTitle("");
			setBody("");
			setPreview(false);
			setKind("once");
			setSendNow(true);
			setRunAt(defaultRunAt());
			setCron("0 9 * * *");
			setFieldErrors({});
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
				<Label>Type</Label>
				<div className="flex rounded-lg border border-input bg-muted/30 p-1">
					<Button
						type="button"
						size="sm"
						variant={kind === "once" ? "default" : "ghost"}
						className={cn(
							"flex-1 font-medium transition-colors",
							kind === "once"
								? "bg-background text-foreground shadow-sm dark:bg-card dark:text-foreground dark:ring-1 dark:ring-border/60"
								: "text-muted-foreground hover:text-foreground",
						)}
						disabled={pending}
						onClick={() => setKind("once")}
					>
						Once
					</Button>
					<Button
						type="button"
						size="sm"
						variant={kind === "recurring" ? "default" : "ghost"}
						className={cn(
							"flex-1 font-medium transition-colors",
							kind === "recurring"
								? "bg-background text-foreground shadow-sm dark:bg-card dark:text-foreground dark:ring-1 dark:ring-border/60"
								: "text-muted-foreground hover:text-foreground",
						)}
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
						onChange={(event) => {
							setBody(event.target.value);
							event.currentTarget.style.height = "auto";
							event.currentTarget.style.height = `${event.currentTarget.scrollHeight}px`;
						}}
						placeholder={BODY_PLACEHOLDER}
						disabled={pending}
						rows={4}
						className={cn(
							"w-full min-w-0 resize-none overflow-hidden rounded-lg border border-input bg-transparent px-2.5 py-1.5 text-base transition-colors outline-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:pointer-events-none disabled:cursor-not-allowed disabled:bg-input/50 disabled:opacity-50 md:text-sm dark:bg-input/30",
						)}
					/>
				)}
			</div>

			{kind === "once" ? (
				<div className="flex flex-col gap-2">
					<div className="flex items-center justify-between gap-2">
						<Label htmlFor="beep-run-at">Run at</Label>
						<Tooltip>
							<TooltipTrigger
								type="button"
								className="text-muted-foreground hover:text-foreground"
								aria-label="Execution time information"
							>
								<HelpCircle className="size-4" />
							</TooltipTrigger>
							<TooltipContent side="top" className="max-w-xs text-xs">
								Execution time for one-off beeps. Uncheck &quot;Send
								immediately&quot; to schedule for a specific time in the future.
							</TooltipContent>
						</Tooltip>
					</div>

					<div className="flex items-center gap-2">
						<input
							type="checkbox"
							id="beep-send-now"
							name="sendNow"
							checked={sendNow}
							onChange={(event) => setSendNow(event.target.checked)}
							disabled={pending}
							className="h-4 w-4 rounded border-input text-primary focus:ring-ring"
						/>
						<Label
							htmlFor="beep-send-now"
							className="cursor-pointer text-sm font-normal"
						>
							Send immediately (now)
						</Label>
					</div>

					{!sendNow ? (
						<div className="flex flex-col gap-2">
							<DateTimePicker
								id="beep-run-at"
								value={runAt}
								onChange={(val) => {
									setRunAt(val);
									setFieldErrors((curr) => ({
										...curr,
										run_at:
											val.getTime() <= Date.now() + 60 * 1000
												? "must be at least 1 minute in the future"
												: undefined,
									}));
								}}
								disabled={pending}
							/>
							{fieldErrors.run_at ? (
								<p className="text-xs text-destructive" role="alert">
									{fieldErrors.run_at}
								</p>
							) : (
								<p className="text-xs text-muted-foreground">
									Must be at least 1 minute in the future.
								</p>
							)}
						</div>
					) : (
						<p className="text-xs text-muted-foreground">
							Will be delivered right after creation.
						</p>
					)}
				</div>
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

			{error ? (
				<p className="text-sm text-destructive" role="alert">
					{error}
				</p>
			) : null}
			<Button
				type="submit"
				disabled={
					pending ||
					(kind === "once" &&
						!sendNow &&
						(runAt.getTime() <= Date.now() + 60 * 1000 ||
							Boolean(fieldErrors.run_at)))
				}
				className="w-fit"
			>
				{pending
					? kind === "once" && sendNow
						? "Sending…"
						: "Creating…"
					: kind === "once" && sendNow
						? "Send beep now"
						: "Create beep"}
			</Button>
		</form>
	);
}
