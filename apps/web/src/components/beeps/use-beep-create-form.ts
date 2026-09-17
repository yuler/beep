import {
	type FormEvent,
	useCallback,
	useEffect,
	useRef,
	useState,
} from "react";
import { createBeep, createBeepProposal } from "@/lib/api/beeps";
import { ApiError } from "@/lib/api/client";
import { translateError } from "@/lib/i18n-labels";
import {
	NOTIFICATION_CHANNELS,
	type NotificationChannel,
} from "@/lib/notification-channels";
import { browserTimezone } from "@/lib/timezone";
import { m } from "@/locale/paraglide/messages";

export const TITLE_MAX_LENGTH = 80;
export const BODY_MAX_LENGTH = 2000;

export const PROMPT_SUGGESTIONS = [
	{ id: "example-1", label: m.beeps_prompt_example_1 },
	{ id: "example-2", label: m.beeps_prompt_example_2 },
	{ id: "example-3", label: m.beeps_prompt_example_3 },
] as const;

export const CRON_PRESETS = [
	{ label: m.beeps_cron_preset_daily_9, value: "0 9 * * *" },
	{ label: m.beeps_cron_preset_weekdays_9, value: "0 9 * * 1-5" },
	{ label: m.beeps_cron_preset_monday_9, value: "0 9 * * 1" },
	{ label: m.beeps_cron_preset_hourly, value: "0 * * * *" },
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

function resolvedChannels(
	defaultChannels: NotificationChannel[],
): NotificationChannel[] {
	return defaultChannels.length > 0 ? defaultChannels : ["email"];
}

export function useBeepCreateForm({
	slug,
	defaultChannels = ["email"],
	onCreated,
	resetOnSuccess = false,
}: {
	slug: string;
	defaultChannels?: NotificationChannel[];
	onCreated: () => Promise<void> | void;
	resetOnSuccess?: boolean;
}) {
	const [prompt, setPrompt] = useState("");
	const [kind, setKind] = useState<"once" | "recurring">("once");
	const [sendNow, setSendNow] = useState(true);
	const [title, setTitle] = useState("");
	const [body, setBody] = useState("");
	const [preview, setPreview] = useState(false);
	const [runAt, setRunAt] = useState<Date>(defaultRunAt);
	const [cron, setCron] = useState("0 9 * * *");
	const [channels, setChannels] = useState<NotificationChannel[]>(() =>
		resolvedChannels(defaultChannels),
	);
	const [fieldErrors, setFieldErrors] = useState<{
		title?: string;
		run_at?: string;
		cron?: string;
	}>({});
	const [proposeMessage, setProposeMessage] = useState<string | null>(null);
	const [error, setError] = useState<string | null>(null);
	const [proposing, setProposing] = useState(false);
	const [submitting, setSubmitting] = useState(false);
	const onCreatedRef = useRef(onCreated);
	onCreatedRef.current = onCreated;

	const reset = useCallback(() => {
		setPrompt("");
		setTitle("");
		setBody("");
		setPreview(false);
		setKind("once");
		setSendNow(true);
		setRunAt(defaultRunAt());
		setCron("0 9 * * *");
		setChannels(resolvedChannels(defaultChannels));
		setFieldErrors({});
		setError(null);
		setProposeMessage(null);
	}, [defaultChannels]);

	useEffect(() => {
		setChannels(resolvedChannels(defaultChannels));
	}, [defaultChannels]);

	async function handlePropose() {
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
				setProposeMessage(proposal.message ?? m.beeps_prompt_autofill_failed());
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

			if (
				proposal.notification_channels &&
				proposal.notification_channels.length > 0
			) {
				const validChannels = proposal.notification_channels.filter(
					(c): c is NotificationChannel =>
						(NOTIFICATION_CHANNELS as readonly string[]).includes(c),
				);
				if (validChannels.length > 0) {
					setChannels(validChannels);
				}
			}

			setFieldErrors({
				title: proposal.errors.title,
				cron: proposal.errors.cron,
				run_at:
					proposal.errors.run_at ??
					(proposal.kind !== "recurring" &&
					nextRunAt &&
					nextRunAt.getTime() <= Date.now() + 60 * 1000
						? m.beeps_run_at_future_error()
						: undefined),
			});
			setProposeMessage(m.beeps_prompt_filled());
		} catch (err) {
			setError(err instanceof ApiError ? err.message : translateError(err));
		} finally {
			setProposing(false);
		}
	}

	async function onPropose(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();
		await handlePropose();
	}

	const isPending = proposing || submitting;
	const submitDisabled =
		isPending ||
		title.trim().length === 0 ||
		channels.length === 0 ||
		(kind === "once" &&
			!sendNow &&
			(runAt.getTime() <= Date.now() + 60 * 1000 ||
				Boolean(fieldErrors.run_at))) ||
		(kind === "recurring" && (!cron.trim() || Boolean(fieldErrors.cron)));

	async function onSubmit(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();
		if (submitting || channels.length === 0 || submitDisabled) return;

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
				notification_channels: channels,
			});
			if (resetOnSuccess) {
				reset();
			}
			setSubmitting(false);
			await onCreatedRef.current();
		} catch (err) {
			setError(err instanceof ApiError ? err.message : translateError(err));
			setSubmitting(false);
		}
	}

	return {
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
		reset,
		handlePropose,
		onPropose,
		onSubmit,
	};
}
