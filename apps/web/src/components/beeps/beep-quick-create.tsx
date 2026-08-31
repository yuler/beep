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
import { useTranslation } from "@/lib/i18n";
import { translateError } from "@/lib/i18n-labels";
import { browserTimezone } from "@/lib/timezone";
import { cn } from "@/lib/utils";

const TITLE_MAX_LENGTH = 80;
const BODY_MAX_LENGTH = 2000;

const PROMPT_SUGGESTIONS = [
	"beeps.prompt_example_1",
	"beeps.prompt_example_2",
	"beeps.prompt_example_3",
] as const;

const CRON_PRESETS = [
	{ labelKey: "beeps.cron_preset_daily_9", value: "0 9 * * *" },
	{ labelKey: "beeps.cron_preset_weekdays_9", value: "0 9 * * 1-5" },
	{ labelKey: "beeps.cron_preset_monday_9", value: "0 9 * * 1" },
	{ labelKey: "beeps.cron_preset_hourly", value: "0 * * * *" },
] as const;

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
	const { t, dict } = useTranslation();
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
		cron?: string;
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
			const proposal = await createBeepProposal(
				slug,
				prompt.trim(),
				browserTimezone(),
			);
			if (proposal.intent === "other") {
				setFieldErrors({});
				setProposeMessage(
					proposal.message ?? t("beeps.prompt_autofill_failed"),
				);
				return;
			}

			const nextRunAt = parseRunAt(proposal.run_at);
			if (proposal.title) setTitle(proposal.title);
			if (proposal.body !== null && proposal.body !== undefined)
				setBody(proposal.body);

			if (proposal.kind === "recurring" || proposal.cron) {
				setKind("recurring");
				setCron(proposal.cron ?? "");
			} else if (nextRunAt) {
				setRunAt(nextRunAt);
				setKind("once");
				setSendNow(false);
			} else {
				setRunAt(defaultRunAt());
				setKind("once");
				setSendNow(true);
			}

			setFieldErrors({
				title: proposal.errors.title,
				cron: proposal.errors.cron,
				run_at:
					proposal.errors.run_at ??
					(proposal.kind !== "recurring" &&
					nextRunAt &&
					nextRunAt.getTime() <= Date.now() + 60 * 1000
						? t("beeps.run_at_future_error")
						: undefined),
			});
			setProposeMessage(t("beeps.prompt_filled"));
		} catch (err) {
			setError(
				err instanceof ApiError
					? err.message
					: translateError(dict, t, err),
			);
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
				timezone: browserTimezone(),
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
			setError(
				err instanceof ApiError
					? err.message
					: translateError(dict, t, err),
			);
		} finally {
			setSubmitting(false);
		}
	}

	const isPending = proposing || submitting;

	return (
		<Card className="w-full shadow-xs">
			<CardHeader className="pb-4">
				<CardTitle className="flex items-center gap-2 text-base font-semibold">
					<span>{t("beeps.create_new_beep")}</span>
				</CardTitle>
			</CardHeader>
			<CardContent className="flex flex-col gap-5">
				<form
					className="flex flex-col gap-3 rounded-xl border border-primary/20 bg-primary/[0.03] p-3.5 dark:border-primary/30 dark:bg-primary/[0.06]"
					onSubmit={onPropose}
				>
					<div className="flex flex-col gap-2">
						<div className="flex items-center justify-between">
							<Label
								htmlFor={`beep-prompt-${slug}`}
								className="flex items-center gap-1.5 text-xs font-semibold text-primary"
							>
								<Sparkles className="size-3.5" />
								{t("beeps.prompt_ai_assistant")}
							</Label>
							<span className="flex items-center gap-1 text-[11px] text-muted-foreground">
								<kbd className="inline-flex h-4.5 min-w-4.5 items-center justify-center rounded border border-border bg-muted/80 px-1 font-mono text-[10px] font-medium text-foreground shadow-2xs">
									⌘
								</kbd>
								<span>+</span>
								<kbd className="inline-flex h-4.5 items-center justify-center rounded border border-border bg-muted/80 px-1 font-mono text-[10px] font-medium text-foreground shadow-2xs">
									Enter
								</kbd>
								<span>{t("beeps.prompt_shortcut")}</span>
							</span>
						</div>
						<textarea
							id={`beep-prompt-${slug}`}
							name="prompt"
							value={prompt}
							onChange={(event) => {
								setPrompt(event.target.value);
								event.currentTarget.style.height = "auto";
								event.currentTarget.style.height = `${event.currentTarget.scrollHeight}px`;
							}}
							onKeyDown={(event) => {
								if ((event.metaKey || event.ctrlKey) && event.key === "Enter") {
									event.preventDefault();
									if (!isPending && prompt.trim().length > 0) {
										event.currentTarget.form?.requestSubmit();
									}
								}
							}}
							placeholder={t("beeps.prompt_placeholder")}
							disabled={isPending}
							rows={2}
							className={cn(
								"w-full min-w-0 resize-none overflow-hidden rounded-lg border border-input bg-background/80 px-2.5 py-1.5 text-sm transition-colors outline-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-2 focus-visible:ring-ring/40 disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-50 dark:bg-input/20",
							)}
						/>

						<div className="flex flex-wrap items-center gap-1.5 pt-0.5">
							{PROMPT_SUGGESTIONS.map((key) => (
								<button
									key={key}
									type="button"
									disabled={isPending}
									onClick={() => setPrompt(t(key))}
									className="rounded-md border border-input/60 bg-background/60 px-2 py-0.5 text-[11px] text-muted-foreground transition-colors hover:border-primary/40 hover:bg-background hover:text-foreground dark:bg-input/10"
								>
									{t(key)}
								</button>
							))}
						</div>

						<div className="flex items-center justify-end pt-1">
							<Button
								type="submit"
								variant="secondary"
								size="xs"
								disabled={isPending || prompt.trim().length === 0}
								aria-label={t("beeps.prompt_autofill")}
								className="font-medium"
							>
								<Sparkles data-icon="inline-start" />
								{proposing ? t("beeps.parsing") : t("beeps.prompt_autofill")}
							</Button>
						</div>
					</div>
				</form>

				{proposeMessage ? (
					<p className="text-xs text-muted-foreground">{proposeMessage}</p>
				) : null}

				<form
					className="flex flex-col gap-4 border-t border-border/80 pt-4"
					onSubmit={onSubmit}
				>
					<div className="flex flex-col gap-2">
						<Label>{t("beeps.type")}</Label>
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
								{t("beeps.kind_once")}
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
								{t("beeps.kind_recurring")}
							</Button>
						</div>
					</div>

					<div className="flex flex-col gap-2">
						<Label htmlFor={`beep-title-${slug}`}>{t("beeps.title")}</Label>
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
							placeholder={t("beeps.body_placeholder")}
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
							<Label htmlFor={`beep-body-${slug}`}>{t("beeps.body")}</Label>
							<div className="flex gap-1">
								<Button
									type="button"
									size="xs"
									variant={preview ? "ghost" : "secondary"}
									aria-pressed={!preview}
									onClick={() => setPreview(false)}
									disabled={isPending}
								>
									{t("beeps.write")}
								</Button>
								<Button
									type="button"
									size="xs"
									variant={preview ? "secondary" : "ghost"}
									aria-pressed={preview}
									onClick={() => setPreview(true)}
									disabled={isPending}
								>
									{t("beeps.preview")}
								</Button>
							</div>
						</div>
						{preview ? (
							<div className="min-h-24 rounded-lg border border-input px-2.5 py-1.5 dark:bg-input/30">
								{body.trim() ? (
									<BeepMarkdown source={body} />
								) : (
									<p className="text-sm text-muted-foreground">
										{t("beeps.nothing_to_preview")}
									</p>
								)}
							</div>
						) : (
							<textarea
								id={`beep-body-${slug}`}
								name="body"
								maxLength={BODY_MAX_LENGTH}
								value={body}
								onChange={(event) => {
									setBody(event.target.value);
									event.currentTarget.style.height = "auto";
									event.currentTarget.style.height = `${event.currentTarget.scrollHeight}px`;
								}}
								placeholder={t("beeps.body_markdown_placeholder")}
								disabled={isPending}
								rows={3}
								className={cn(
									"w-full min-w-0 resize-none overflow-hidden rounded-lg border border-input bg-transparent px-2.5 py-1.5 text-base transition-colors outline-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:pointer-events-none disabled:cursor-not-allowed disabled:bg-input/50 disabled:opacity-50 md:text-sm dark:bg-input/30",
								)}
							/>
						)}
					</div>

					{kind === "once" ? (
						<div className="flex flex-col gap-2">
							<div className="flex items-center justify-between gap-2">
								<Label htmlFor={`beep-run-at-${slug}`}>
									{t("beeps.run_at")}
								</Label>
								<Tooltip>
									<TooltipTrigger
										type="button"
										className="text-muted-foreground hover:text-foreground"
										aria-label={t("beeps.run_at")}
									>
										<HelpCircle className="size-4" />
									</TooltipTrigger>
									<TooltipContent side="top" className="max-w-xs text-xs">
										{t("beeps.run_at_tooltip")}
									</TooltipContent>
								</Tooltip>
							</div>

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
									className="cursor-pointer text-sm font-normal"
								>
									{t("beeps.send_immediately")}
								</Label>
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
														? t("beeps.run_at_future_error")
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
											{t("beeps.run_at_future_hint")}
										</p>
									)}
								</div>
							) : (
								<p className="text-xs text-muted-foreground">
									{t("beeps.send_immediately_hint")}
								</p>
							)}
						</div>
					) : null}

					{kind === "recurring" ? (
						<div className="flex flex-col gap-2.5">
							<div className="flex items-center justify-between gap-2">
								<Label htmlFor={`beep-cron-${slug}`}>
									{t("beeps.cron_schedule")}
								</Label>
								<Tooltip>
									<TooltipTrigger
										type="button"
										className="text-muted-foreground hover:text-foreground"
										aria-label={t("beeps.cron_schedule")}
									>
										<HelpCircle className="size-4" />
									</TooltipTrigger>
									<TooltipContent side="top" className="max-w-xs text-xs">
										{t("beeps.cron_tooltip")}
									</TooltipContent>
								</Tooltip>
							</div>

							<div className="flex flex-wrap gap-1.5">
								{CRON_PRESETS.map((preset) => (
									<Button
										key={preset.value}
										type="button"
										size="sm"
										variant={cron === preset.value ? "secondary" : "outline"}
										className="h-7 text-xs font-normal"
										disabled={isPending}
										onClick={() => setCron(preset.value)}
									>
										{t(preset.labelKey)}
									</Button>
								))}
							</div>

							<Input
								id={`beep-cron-${slug}`}
								name="cron"
								required
								value={cron}
								onChange={(event) => {
									setCron(event.target.value);
									setFieldErrors((curr) => ({ ...curr, cron: undefined }));
								}}
								placeholder="0 9 * * *"
								disabled={isPending}
								className="font-mono text-sm"
								aria-invalid={Boolean(fieldErrors.cron)}
							/>
							{fieldErrors.cron ? (
								<p className="text-xs text-destructive" role="alert">
									{fieldErrors.cron}
								</p>
							) : (
								<p className="text-xs text-muted-foreground">
									{t("beeps.cron_format")}{" "}
									<code className="rounded bg-muted px-1 py-0.5 font-mono text-[11px]">
										{t("beeps.cron_format_parts")}
									</code>
								</p>
							)}
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
									Boolean(fieldErrors.run_at))) ||
							(kind === "recurring" && Boolean(fieldErrors.cron))
						}
						className="w-fit"
					>
						{submitting
							? kind === "once" && sendNow
								? t("beeps.sending")
								: t("beeps.creating")
							: kind === "once" && sendNow
								? t("beeps.send_beep_now")
								: t("beeps.create_beep")}
					</Button>
				</form>
			</CardContent>
		</Card>
	);
}
