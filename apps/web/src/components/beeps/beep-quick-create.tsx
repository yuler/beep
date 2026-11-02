import { HelpCircle, Sparkles } from "lucide-react";
import { BeepMarkdown } from "@/components/beeps/beep-markdown";
import {
	BODY_MAX_LENGTH,
	CRON_PRESETS,
	PROMPT_SUGGESTIONS,
	TITLE_MAX_LENGTH,
	useBeepCreateForm,
} from "@/components/beeps/use-beep-create-form";
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
import { channelLabel } from "@/lib/i18n-labels";
import {
	NOTIFICATION_CHANNELS,
	type NotificationChannel,
	toggleChannel,
} from "@/lib/notification-channels";
import { cn } from "@/lib/utils";
import { m } from "@/locale/paraglide/messages";

export function BeepQuickCreate({
	slug,
	onCreated,
	defaultChannels = [],
}: {
	slug: string;
	onCreated: () => Promise<void> | void;
	defaultChannels?: NotificationChannel[];
}) {
	const {
		prompt,
		setPrompt,
		kind,
		setKind,
		sendNow,
		setSendNow,
		title,
		setTitle,
		body,
		setBody,
		intent,
		setIntent,
		metadata,
		setMetadata,
		preview,
		setPreview,
		runAt,
		setRunAt,
		cron,
		setCron,
		channels,
		setChannels,
		fieldErrors,
		setFieldErrors,
		proposeMessage,
		error,
		proposing,
		submitting,
		isPending,
		submitDisabled,
		onPropose,
		onSubmit,
	} = useBeepCreateForm({
		slug,
		defaultChannels,
		onCreated,
		resetOnSuccess: true,
	});

	return (
		<Card className="w-full">
			<CardHeader className="pb-4">
				<CardTitle className="flex items-center gap-2 text-base font-medium">
					<span>{m.beeps_create_new_beep()}</span>
				</CardTitle>
			</CardHeader>
			<CardContent className="flex flex-col gap-5">
				<form
					className="flex flex-col gap-3 rounded-md border border-line bg-wash p-3.5"
					onSubmit={onPropose}
				>
					<div className="flex flex-col gap-2">
						<div className="flex items-center justify-between">
							<Label
								htmlFor={`beep-prompt-${slug}`}
								className="flex items-center gap-1.5 text-xs font-medium text-ink"
							>
								<Sparkles className="size-3.5" />
								{m.beeps_prompt_ai_assistant()}
							</Label>
							<span className="hidden sm:flex items-center gap-1 text-[11px] text-gray">
								<kbd className="inline-flex h-4.5 min-w-4.5 items-center justify-center rounded-sm border border-line bg-ground px-1 font-mono text-[10px] font-normal text-ink">
									⌘
								</kbd>
								<span>+</span>
								<kbd className="inline-flex h-4.5 items-center justify-center rounded-sm border border-line bg-ground px-1 font-mono text-[10px] font-normal text-ink">
									Enter
								</kbd>
								<span>{m.beeps_prompt_shortcut()}</span>
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
							placeholder={m.beeps_prompt_placeholder()}
							disabled={isPending}
							rows={2}
							className={cn(
								"w-full min-w-0 resize-none overflow-hidden rounded-sm border border-line bg-ground px-3 py-2 text-sm text-ink transition-colors outline-none placeholder:text-gray focus-visible:border-ink focus-visible:outline focus-visible:outline-1 focus-visible:outline-ink focus-visible:outline-offset-2 disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-40",
							)}
						/>

						<div className="flex flex-wrap items-center gap-1.5 pt-0.5">
							{PROMPT_SUGGESTIONS.map(({ id, label }) => (
								<button
									key={id}
									type="button"
									disabled={isPending}
									onClick={() => setPrompt(label())}
									className="rounded-sm border border-line bg-ground px-2 py-0.5 text-[11px] text-gray transition-colors hover:border-ink hover:text-ink"
								>
									{label()}
								</button>
							))}
						</div>

						<div className="flex items-center justify-end pt-1">
							<Button
								type="submit"
								variant="secondary"
								size="xs"
								disabled={isPending || prompt.trim().length === 0}
								aria-label={m.beeps_prompt_autofill()}
								className="font-medium"
							>
								<Sparkles data-icon="inline-start" />
								{proposing ? m.beeps_parsing() : m.beeps_prompt_autofill()}
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
						<Label>{m.beeps_type()}</Label>
						<div className="flex rounded-sm border border-line bg-wash p-1 gap-1">
							<Button
								type="button"
								size="sm"
								variant={kind === "once" ? "solid" : "ghost"}
								className="flex-1"
								disabled={isPending}
								onClick={() => setKind("once")}
							>
								{m.beeps_kind_once()}
							</Button>
							<Button
								type="button"
								size="sm"
								variant={kind === "recurring" ? "solid" : "ghost"}
								className="flex-1"
								disabled={isPending}
								onClick={() => setKind("recurring")}
							>
								{m.beeps_kind_recurring()}
							</Button>
						</div>
					</div>

					<div className="flex flex-col gap-2">
						<Label htmlFor={`beep-title-${slug}`}>{m.beeps_title()}</Label>
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
							placeholder={m.beeps_title_placeholder()}
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
							<Label htmlFor={`beep-body-${slug}`}>{m.beeps_body()}</Label>
							<div className="flex gap-1">
								<Button
									type="button"
									size="xs"
									variant={preview ? "ghost" : "secondary"}
									aria-pressed={!preview}
									onClick={() => setPreview(false)}
									disabled={isPending}
								>
									{m.beeps_write()}
								</Button>
								<Button
									type="button"
									size="xs"
									variant={preview ? "secondary" : "ghost"}
									aria-pressed={preview}
									onClick={() => setPreview(true)}
									disabled={isPending}
								>
									{m.beeps_preview()}
								</Button>
							</div>
						</div>
						{preview ? (
							<div className="min-h-32 rounded-sm border border-line px-3 py-2">
								{body.trim() ? (
									<BeepMarkdown source={body} />
								) : (
									<p className="text-sm text-gray">
										{m.beeps_nothing_to_preview()}
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
								placeholder={m.beeps_body_markdown_placeholder()}
								disabled={isPending}
								rows={6}
								className={cn(
									"w-full min-w-0 min-h-32 resize-none overflow-hidden rounded-sm border border-line bg-transparent px-3 py-2 text-base text-ink transition-colors outline-none placeholder:text-gray focus-visible:border-ink focus-visible:outline focus-visible:outline-1 focus-visible:outline-ink focus-visible:outline-offset-2 disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-40 md:text-sm",
								)}
							/>
						)}
					</div>

					<div className="flex flex-col gap-2">
						<div className="flex items-center justify-between gap-2">
							<Label htmlFor={`beep-intent-${slug}`}>{m.beeps_intent()}</Label>
							<span className="text-[11px] text-gray">
								{m.common_optional()}
							</span>
						</div>
						<Input
							id={`beep-intent-${slug}`}
							name="intent"
							value={intent}
							onChange={(event) => setIntent(event.target.value)}
							placeholder={m.beeps_intent_placeholder()}
							disabled={isPending}
							className="font-mono text-sm"
						/>
						<p className="text-[11px] text-gray">{m.beeps_intent_hint()}</p>
					</div>

					<div className="flex flex-col gap-2">
						<div className="flex items-center justify-between gap-2">
							<Label htmlFor={`beep-metadata-${slug}`}>
								{m.beeps_metadata()}
							</Label>
							<span className="text-[11px] text-gray">
								{m.common_optional()}
							</span>
						</div>
						<textarea
							id={`beep-metadata-${slug}`}
							name="metadata"
							value={metadata}
							onChange={(event) => {
								setMetadata(event.target.value);
								setFieldErrors((curr) => ({ ...curr, metadata: undefined }));
							}}
							placeholder='{"key": "value"}'
							disabled={isPending}
							rows={2}
							className={cn(
								"w-full min-w-0 min-h-16 resize-none font-mono text-sm rounded-sm border border-line bg-transparent px-3 py-2 text-ink transition-colors outline-none placeholder:text-gray focus-visible:border-ink focus-visible:outline focus-visible:outline-1 focus-visible:outline-ink focus-visible:outline-offset-2 disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-40",
								fieldErrors.metadata &&
									"border-danger focus-visible:border-danger focus-visible:outline-danger",
							)}
						/>
						{fieldErrors.metadata ? (
							<p className="text-xs text-destructive" role="alert">
								{fieldErrors.metadata}
							</p>
						) : (
							<p className="text-[11px] text-muted-foreground">
								{m.beeps_metadata_hint()}
							</p>
						)}
					</div>

					{kind === "once" ? (
						<div className="flex flex-col gap-2">
							<div className="flex items-center justify-between gap-2">
								<Label htmlFor={`beep-run-at-${slug}`}>
									{m.beeps_run_at()}
								</Label>
								<Tooltip>
									<TooltipTrigger
										type="button"
										className="text-muted-foreground hover:text-foreground"
										aria-label={m.beeps_run_at()}
									>
										<HelpCircle className="size-4" />
									</TooltipTrigger>
									<TooltipContent side="top" className="max-w-xs text-xs">
										{m.beeps_run_at_tooltip()}
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
									{m.beeps_send_immediately()}
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
														? m.beeps_run_at_future_error()
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
											{m.beeps_run_at_future_hint()}
										</p>
									)}
								</div>
							) : (
								<p className="text-xs text-muted-foreground">
									{m.beeps_send_immediately_hint()}
								</p>
							)}
						</div>
					) : null}

					{kind === "recurring" ? (
						<div className="flex flex-col gap-2.5">
							<div className="flex items-center justify-between gap-2">
								<Label htmlFor={`beep-cron-${slug}`}>
									{m.beeps_cron_schedule()}
								</Label>
								<Tooltip>
									<TooltipTrigger
										type="button"
										className="text-muted-foreground hover:text-foreground"
										aria-label={m.beeps_cron_schedule()}
									>
										<HelpCircle className="size-4" />
									</TooltipTrigger>
									<TooltipContent side="top" className="max-w-xs text-xs">
										{m.beeps_cron_tooltip()}
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
										{preset.label()}
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
									{m.beeps_cron_format()}{" "}
									<code className="rounded bg-muted px-1 py-0.5 font-mono text-[11px]">
										{m.beeps_cron_format_parts()}
									</code>
								</p>
							)}
						</div>
					) : null}

					{/* Notification Channels */}
					<div className="flex flex-col gap-2">
						<Label>{m.beeps_channels()}</Label>
						<div className="flex flex-col gap-2 rounded-sm border border-line p-3">
							{NOTIFICATION_CHANNELS.map((channel) => (
								<Label
									key={channel}
									className="flex items-center gap-2 font-normal cursor-pointer text-sm"
								>
									<input
										type="checkbox"
										className="size-4 accent-primary rounded-sm"
										checked={channels.includes(channel)}
										disabled={isPending}
										onChange={(e) =>
											setChannels((curr) =>
												toggleChannel(curr, channel, e.target.checked),
											)
										}
									/>
									{channelLabel(channel)}
								</Label>
							))}
						</div>
						{channels.length === 0 ? (
							<p className="text-xs text-destructive" role="alert">
								{m.beeps_channels_required()}
							</p>
						) : (
							<p className="text-[11px] text-muted-foreground">
								{m.beeps_notification_channels_hint()}
							</p>
						)}
					</div>

					{error ? (
						<p className="text-sm text-destructive" role="alert">
							{error}
						</p>
					) : null}

					<Button
						type="submit"
						disabled={submitDisabled}
						className="w-full sm:w-fit"
					>
						{submitting
							? kind === "once" && sendNow
								? m.beeps_sending()
								: m.beeps_creating()
							: kind === "once" && sendNow
								? m.beeps_send_beep_now()
								: m.beeps_create_beep()}
					</Button>
				</form>
			</CardContent>
		</Card>
	);
}
