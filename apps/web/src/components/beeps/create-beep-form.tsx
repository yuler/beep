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
import { useTranslation } from "@/lib/i18n";
import { translateError } from "@/lib/i18n-labels";
import { browserTimezone } from "@/lib/timezone";
import { cn } from "@/lib/utils";

const TITLE_MAX_LENGTH = 80;
const BODY_MAX_LENGTH = 2000;

const CRON_PRESETS = [
	{ labelKey: "beeps.cron_preset_daily_9", value: "0 9 * * *" },
	{ labelKey: "beeps.cron_preset_weekdays_9", value: "0 9 * * 1-5" },
	{ labelKey: "beeps.cron_preset_monday_9", value: "0 9 * * 1" },
	{ labelKey: "beeps.cron_preset_hourly", value: "0 * * * *" },
] as const;

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
	const { t, dict } = useTranslation();
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
				timezone: browserTimezone(),
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
			setError(
				err instanceof ApiError ? err.message : translateError(dict, t, err),
			);
		} finally {
			setPending(false);
		}
	}

	return (
		<form className="flex flex-col gap-4" onSubmit={onSubmit}>
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
						disabled={pending}
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
						disabled={pending}
						onClick={() => setKind("recurring")}
					>
						{t("beeps.kind_recurring")}
					</Button>
				</div>
			</div>

			<div className="flex flex-col gap-2">
				<Label htmlFor="beep-title">{t("beeps.title")}</Label>
				<Input
					id="beep-title"
					name="title"
					required
					maxLength={TITLE_MAX_LENGTH}
					value={title}
					onChange={(event) => setTitle(event.target.value)}
					placeholder={t("beeps.body_placeholder")}
					disabled={pending}
				/>
			</div>
			<div className="flex flex-col gap-2">
				<div className="flex items-center justify-between gap-2">
					<Label htmlFor="beep-body">{t("beeps.body")}</Label>
					<div className="flex gap-1">
						<Button
							type="button"
							size="xs"
							variant={preview ? "ghost" : "secondary"}
							aria-pressed={!preview}
							onClick={() => setPreview(false)}
							disabled={pending}
						>
							{t("beeps.write")}
						</Button>
						<Button
							type="button"
							size="xs"
							variant={preview ? "secondary" : "ghost"}
							aria-pressed={preview}
							onClick={() => setPreview(true)}
							disabled={pending}
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
						id="beep-body"
						name="body"
						maxLength={BODY_MAX_LENGTH}
						value={body}
						onChange={(event) => {
							setBody(event.target.value);
							event.currentTarget.style.height = "auto";
							event.currentTarget.style.height = `${event.currentTarget.scrollHeight}px`;
						}}
						placeholder={t("beeps.body_markdown_placeholder")}
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
						<Label htmlFor="beep-run-at">{t("beeps.run_at")}</Label>
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
							{t("beeps.send_immediately")}
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
												? t("beeps.run_at_future_error")
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
						<Label htmlFor="beep-cron">{t("beeps.cron_schedule")}</Label>
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
								disabled={pending}
								onClick={() => setCron(preset.value)}
							>
								{t(preset.labelKey)}
							</Button>
						))}
					</div>

					<Input
						id="beep-cron"
						name="cron"
						required
						value={cron}
						onChange={(event) => setCron(event.target.value)}
						placeholder="0 9 * * *"
						disabled={pending}
						className="font-mono text-sm"
					/>
					<p className="text-xs text-muted-foreground">
						{t("beeps.cron_format")}{" "}
						<code className="rounded bg-muted px-1 py-0.5 font-mono text-[11px]">
							{t("beeps.cron_format_parts")}
						</code>
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
						? t("beeps.sending")
						: t("beeps.creating")
					: kind === "once" && sendNow
						? t("beeps.send_beep_now")
						: t("beeps.create_beep")}
			</Button>
		</form>
	);
}
