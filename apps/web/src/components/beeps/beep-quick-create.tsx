import { type FormEvent, useState } from "react";

import { CreateBeepForm } from "@/components/beeps/create-beep-form";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { DateTimePicker } from "@/components/ui/datetime-picker";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { createBeep, createBeepProposal } from "@/lib/api/beeps";
import { ApiError } from "@/lib/api/client";
import { cn } from "@/lib/utils";

const TITLE_MAX_LENGTH = 80;
const BODY_MAX_LENGTH = 2000;

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
	const [title, setTitle] = useState("");
	const [body, setBody] = useState("");
	const [runAt, setRunAt] = useState<Date | null>(null);
	const [fallbackRunAt] = useState(() => new Date(Date.now() + 60 * 60 * 1000));
	const [fieldErrors, setFieldErrors] = useState<{
		title?: string;
		run_at?: string;
	}>({});
	const [message, setMessage] = useState<string | null>(null);
	const [error, setError] = useState<string | null>(null);
	const [showPreview, setShowPreview] = useState(false);
	const [manualOpen, setManualOpen] = useState(false);
	const [proposing, setProposing] = useState(false);
	const [confirming, setConfirming] = useState(false);

	const canConfirm =
		title.trim().length > 0 &&
		runAt instanceof Date &&
		runAt.getTime() > Date.now() &&
		!fieldErrors.title &&
		!fieldErrors.run_at;

	async function onPropose(event: FormEvent<HTMLFormElement>) {
		event.preventDefault();
		setError(null);
		setMessage(null);
		setProposing(true);

		try {
			const proposal = await createBeepProposal(slug, prompt.trim());
			if (proposal.intent === "other") {
				setShowPreview(false);
				setFieldErrors({});
				setMessage(
					proposal.message ??
						"Describe the reminder time and what to be reminded of.",
				);
				return;
			}

			const nextRunAt = parseRunAt(proposal.run_at);
			setTitle(proposal.title ?? "");
			setBody(proposal.body ?? "");
			setRunAt(nextRunAt);
			setFieldErrors({
				title: proposal.errors.title,
				run_at:
					proposal.errors.run_at ??
					(nextRunAt && nextRunAt.getTime() <= Date.now()
						? "must be in the future"
						: undefined),
			});
			setShowPreview(true);
		} catch (err) {
			setShowPreview(false);
			setManualOpen(true);
			setError(err instanceof ApiError ? err.message : "Something went wrong.");
		} finally {
			setProposing(false);
		}
	}

	async function onConfirm() {
		if (!canConfirm || !runAt) return;
		setError(null);
		setConfirming(true);

		try {
			await createBeep(slug, {
				title: title.trim(),
				body: body.trim() || null,
				run_at: runAt.toISOString(),
			});
			setPrompt("");
			setTitle("");
			setBody("");
			setRunAt(null);
			setFieldErrors({});
			setShowPreview(false);
			setMessage(null);
			await onCreated();
		} catch (err) {
			setError(err instanceof ApiError ? err.message : "Something went wrong.");
		} finally {
			setConfirming(false);
		}
	}

	return (
		<div className="flex flex-col gap-4">
			<Card className="max-w-105">
				<CardHeader>
					<CardTitle>Quick create</CardTitle>
				</CardHeader>
				<CardContent className="flex flex-col gap-4">
					<form className="flex flex-col gap-3" onSubmit={onPropose}>
						<div className="flex flex-col gap-2">
							<Label htmlFor={`beep-prompt-${slug}`}>Reminder</Label>
							<textarea
								id={`beep-prompt-${slug}`}
								name="prompt"
								value={prompt}
								onChange={(event) => setPrompt(event.target.value)}
								placeholder="Tomorrow 9am call mom"
								disabled={proposing || confirming}
								rows={3}
								required
								className={cn(
									"w-full min-w-0 rounded-lg border border-input bg-transparent px-2.5 py-1.5 text-base transition-colors outline-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:pointer-events-none disabled:cursor-not-allowed disabled:bg-input/50 disabled:opacity-50 md:text-sm dark:bg-input/30",
								)}
							/>
						</div>
						<Button
							type="submit"
							disabled={proposing || confirming || prompt.trim().length === 0}
							className="w-fit"
						>
							{proposing ? "Proposing…" : "Propose beep"}
						</Button>
					</form>

					{message ? (
						<p className="text-sm text-muted-foreground">{message}</p>
					) : null}
					{error ? (
						<p className="text-sm text-destructive" role="alert">
							{error}
						</p>
					) : null}

					{showPreview ? (
						<div className="flex flex-col gap-4 border-t border-border pt-4">
							<div className="flex flex-col gap-2">
								<Label htmlFor={`beep-proposal-title-${slug}`}>Title</Label>
								<Input
									id={`beep-proposal-title-${slug}`}
									value={title}
									maxLength={TITLE_MAX_LENGTH}
									onChange={(event) => {
										setTitle(event.target.value);
										setFieldErrors((current) => ({
											...current,
											title: undefined,
										}));
									}}
									disabled={confirming}
									aria-invalid={Boolean(fieldErrors.title)}
								/>
								{fieldErrors.title ? (
									<p className="text-sm text-destructive" role="alert">
										{fieldErrors.title}
									</p>
								) : null}
							</div>
							<div className="flex flex-col gap-2">
								<Label htmlFor={`beep-proposal-body-${slug}`}>Body</Label>
								<textarea
									id={`beep-proposal-body-${slug}`}
									value={body}
									maxLength={BODY_MAX_LENGTH}
									onChange={(event) => setBody(event.target.value)}
									disabled={confirming}
									rows={4}
									className={cn(
										"w-full min-w-0 rounded-lg border border-input bg-transparent px-2.5 py-1.5 text-base transition-colors outline-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:pointer-events-none disabled:cursor-not-allowed disabled:bg-input/50 disabled:opacity-50 md:text-sm dark:bg-input/30",
									)}
								/>
							</div>
							<div className="flex flex-col gap-2">
								<DateTimePicker
									id={`beep-proposal-run-at-${slug}`}
									value={runAt ?? fallbackRunAt}
									onChange={(value) => {
										setRunAt(value);
										setFieldErrors((current) => ({
											...current,
											run_at:
												value.getTime() <= Date.now()
													? "must be in the future"
													: undefined,
										}));
									}}
									disabled={confirming}
								/>
								{fieldErrors.run_at ? (
									<p className="text-sm text-destructive" role="alert">
										{fieldErrors.run_at}
									</p>
								) : null}
							</div>
							<div className="flex flex-wrap gap-2">
								<Button
									type="button"
									onClick={onConfirm}
									disabled={!canConfirm || confirming}
								>
									{confirming ? "Creating…" : "Confirm"}
								</Button>
								<Button
									type="button"
									variant="ghost"
									disabled={confirming}
									onClick={() => {
										setShowPreview(false);
										setFieldErrors({});
									}}
								>
									Cancel
								</Button>
							</div>
						</div>
					) : null}
				</CardContent>
			</Card>

			<div className="max-w-105">
				<Button
					type="button"
					variant="ghost"
					size="sm"
					aria-expanded={manualOpen}
					onClick={() => setManualOpen((open) => !open)}
				>
					{manualOpen ? "Hide form" : "Fill in manually"}
				</Button>
				{manualOpen ? (
					<Card className="mt-3">
						<CardHeader>
							<CardTitle>New beep</CardTitle>
						</CardHeader>
						<CardContent>
							<CreateBeepForm slug={slug} onCreated={onCreated} />
						</CardContent>
					</Card>
				) : null}
			</div>
		</div>
	);
}
