import { HelpCircle, Sparkles } from "lucide-react";
import { type FormEvent, useState } from "react";

import { BeepMarkdown } from "@/components/beeps/beep-markdown";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { DateTimePicker } from "@/components/ui/datetime-picker";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
	Tooltip,
	TooltipContent,
	TooltipTrigger,
} from "@/components/ui/tooltip";
import { createBeep, createBeepProposal } from "@/lib/api/beeps";
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

function parseRunAt(value: string | null) {
	if (!value) return null;
	const date = new Date(value);
	if (Number.isNaN(date.getTime())) return null;
	return date;
}

export function BeepQuickCreate({
	slug,
	onCreated,
}: {
	slug: string;
	onCreated: () => Promise<void> | void;
}) {
	const [prompt, setPrompt] = useState("");
	const [kind, setKind] = useState<"once" | "recurring">("once");
	const [sendNow, setSendNow] = useState(true);
	const [title, setTitle] = useState("");
	const [body, setBody] = useState("");
	const [preview, setPreview] = useState(false);
	const [runAt, setRunAt] = useState<Date>(defaultRunAt);
	const [cron, setCron] = useState("0 9 * * *");
	const [fieldErrors, setFieldErrors] = useState<{
		title?: string;
		run_at?: string;
	}>({});
	const [proposeMessage, setProposeMessage] = useState<string | null>(null);
	const [error, setError] = useState<string | null>(null);
	const [proposing, setProposing] = useState(false);
	const [submitting, setSubmitting] = useState(false);

	async function onPropose(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();
		if (!prompt.trim()) return;

		setError(null);
		setProposeMessage(null);
		setProposing(true);

		try {
			const proposal = await createBeepProposal(slug, prompt.trim());
			if (proposal.intent === "other") {
				setFieldErrors({});
				setProposeMessage(
					proposal.message ??
						"Describe the reminder time and what to be reminded of.",
				);
				return;
			}

			const nextRunAt = parseRunAt(proposal.run_at);
			if (proposal.title) setTitle(proposal.title);
			if (proposal.body !== null && proposal.body !== undefined)
				setBody(proposal.body);
			if (nextRunAt) {
				setRunAt(nextRunAt);
				setKind("once");
				setSendNow(false);
			}

			setFieldErrors({
				title: proposal.errors.title,
				run_at:
					proposal.errors.run_at ??
					(nextRunAt && nextRunAt.getTime() <= Date.now()
						? "must be in the future"
						: undefined),
			});
			setProposeMessage("Filled in form from your prompt.");
		} catch (err) {
			setError(err instanceof ApiError ? err.message : "Something went wrong.");
		} finally {
			setProposing(false);
		}
	}

	async function onSubmit(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();
		if (submitting) return;

		setError(null);
		setProposeMessage(null);
		setSubmitting(true);

		try {
			await createBeep(slug, {
				title: title.trim(),
				body: body.trim() || null,
				kind,
				run_at: kind === "once" && !sendNow ? runAt.toISOString() : null,
				cron: kind === "recurring" ? cron.trim() : null,
			});
			setPrompt("");
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
			setSubmitting(false);
		}
	}

	const isPending = proposing || submitting;

	return (
		<Card className="max-w-105">
			<CardHeader>
				<CardTitle>Create beep</CardTitle>
			</CardHeader>
			<CardContent className="flex flex-col gap-5">
				{/* AI prompt fill section */}
				<form className="flex flex-col gap-3" onSubmit={onPropose}>
					<div className="flex flex-col gap-2">
						<Label htmlFor={`beep-prompt-${slug}`}>AI Prompt Assistant</Label>
						<div className="flex gap-2">
							<Input
								id={`beep-prompt-${slug}`}
								name="prompt"
								value={prompt}
								onChange={(event) => setPrompt(event.target.value)}
								placeholder="e.g. Remind me tomorrow at 9am to check metrics"
								disabled={isPending}
								className="flex-1"
							/>
							<Button
								type="submit"
								variant="secondary"
								disabled={isPending || prompt.trim().length === 0}
								aria-label="Auto-fill form with AI"
							>
								<Sparkles data-icon="inline-start" />
								{proposing ? "Parsing…" : "Auto-fill"}
							</Button>
						</div>
						<p className="text-xs text-muted-foreground">
							Describe what and when in plain language to populate the form
							below.
						</p>
					</div>
				</form>

				{proposeMessage ? (
					<p className="text-xs text-muted-foreground">{proposeMessage}</p>
				) : null}

				{/* Main Beep Creation Form (Always Visible) */}
				<form
					className="flex flex-col gap-4 border-t border-border pt-4"
					onSubmit={onSubmit}
				>
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
								disabled={isPending}
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
								disabled={isPending}
								onClick={() => setKind("recurring")}
							>
								Recurring
							</Button>
						</div>
					</div>

					<div className="flex flex-col gap-2">
						<Label htmlFor={`beep-title-${slug}`}>Title</Label>
						<Input
							id={`beep-title-${slug}`}
							name="title"
							required
							maxLength={TITLE_MAX_LENGTH}
							value={title}
							onChange={(event) => {
								setTitle(event.target.value);
								setFieldErrors((curr) => ({ ...curr, title: undefined }));
							}}
							placeholder="Call mom"
							disabled={isPending}
							aria-invalid={Boolean(fieldErrors.title)}
						/>
						{fieldErrors.title ? (
							<p className="text-xs text-destructive" role="alert">
								{fieldErrors.title}
							</p>
						) : null}
					</div>

					<div className="flex flex-col gap-2">
						<div className="flex items-center justify-between gap-2">
							<Label htmlFor={`beep-body-${slug}`}>Body</Label>
							<div className="flex gap-1">
								<Button
									type="button"
									size="xs"
									variant={preview ? "ghost" : "secondary"}
									aria-pressed={!preview}
									onClick={() => setPreview(false)}
									disabled={isPending}
								>
									Write
								</Button>
								<Button
									type="button"
									size="xs"
									variant={preview ? "secondary" : "ghost"}
									aria-pressed={preview}
									onClick={() => setPreview(true)}
									disabled={isPending}
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
								id={`beep-body-${slug}`}
								name="body"
								maxLength={BODY_MAX_LENGTH}
								value={body}
								onChange={(event) => setBody(event.target.value)}
								placeholder={BODY_PLACEHOLDER}
								disabled={isPending}
								rows={4}
								className={cn(
									"w-full min-w-0 rounded-lg border border-input bg-transparent px-2.5 py-1.5 text-base transition-colors outline-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:pointer-events-none disabled:cursor-not-allowed disabled:bg-input/50 disabled:opacity-50 md:text-sm dark:bg-input/30",
								)}
							/>
						)}
					</div>

					{kind === "once" ? (
						<div className="flex flex-col gap-3">
							<div className="flex items-center justify-between gap-2">
								<div className="flex items-center gap-2">
									<input
										type="checkbox"
										id={`beep-send-now-${slug}`}
										name="sendNow"
										checked={sendNow}
										onChange={(event) => setSendNow(event.target.checked)}
										disabled={isPending}
										className="h-4 w-4 rounded border-input text-primary focus:ring-ring"
									/>
									<Label
										htmlFor={`beep-send-now-${slug}`}
										className="cursor-pointer font-normal"
									>
										Send immediately (now)
									</Label>
								</div>

								<Tooltip>
									<TooltipTrigger
										type="button"
										className="text-muted-foreground hover:text-foreground"
										aria-label="Execution time information"
									>
										<HelpCircle className="size-4" />
									</TooltipTrigger>
									<TooltipContent side="top" className="max-w-xs text-xs">
										Optional scheduled time for one-off beeps. If provided, it
										must be at least 1 minute in the future. If left blank, it
										will be sent immediately upon creation.
									</TooltipContent>
								</Tooltip>
							</div>

							{!sendNow ? (
								<div className="flex flex-col gap-2">
									<DateTimePicker
										id={`beep-run-at-${slug}`}
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
										disabled={isPending}
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
							<Label htmlFor={`beep-cron-${slug}`}>Cron expression</Label>
							<Input
								id={`beep-cron-${slug}`}
								name="cron"
								required
								value={cron}
								onChange={(event) => setCron(event.target.value)}
								placeholder="0 9 * * *"
								disabled={isPending}
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
							isPending ||
							title.trim().length === 0 ||
							(kind === "once" &&
								!sendNow &&
								(runAt.getTime() <= Date.now() + 60 * 1000 ||
									Boolean(fieldErrors.run_at)))
						}
						className="w-fit"
					>
						{submitting
							? kind === "once" && sendNow
								? "Sending…"
								: "Creating…"
							: kind === "once" && sendNow
								? "Send beep now"
								: "Create beep"}
					</Button>
				</form>
			</CardContent>
		</Card>
	);
}
