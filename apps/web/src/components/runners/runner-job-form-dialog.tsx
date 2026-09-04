import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { ApiError } from "@/lib/api/client";
import {
	createRunnerJob,
	type RunnerJob,
	updateRunnerJob,
} from "@/lib/api/runner-jobs";
import { translateError } from "@/lib/i18n-labels";
import { browserTimezone } from "@/lib/timezone";
import { m } from "@/locale/paraglide/messages";

interface RunnerJobFormDialogProps {
	slug: string;
	runnerId: string;
	job?: RunnerJob | null;
	open: boolean;
	onOpenChange: (open: boolean) => void;
	onSuccess: (job: RunnerJob) => void;
}

export function RunnerJobFormDialog({
	slug,
	runnerId,
	job,
	open,
	onOpenChange,
	onSuccess,
}: RunnerJobFormDialogProps) {
	const isEdit = Boolean(job);
	const [name, setName] = useState("");
	const [jobSlug, setJobSlug] = useState("");
	const [cron, setCron] = useState("*/5 * * * *");
	const [timezone, setTimezone] = useState("");
	const [timeoutSeconds, setTimeoutSeconds] = useState(30);
	const [description, setDescription] = useState("");
	const [submitting, setSubmitting] = useState(false);
	const [error, setError] = useState<string | null>(null);

	useEffect(() => {
		if (!open) return;
		if (job) {
			setName(job.name);
			setJobSlug(job.slug);
			setCron(job.cron || "*/5 * * * *");
			setTimezone(job.timezone || browserTimezone());
			setTimeoutSeconds(job.timeout_seconds || 30);
			const desc =
				job.config && typeof job.config.description === "string"
					? job.config.description
					: "";
			setDescription(desc);
		} else {
			setName("");
			setJobSlug("");
			setCron("*/5 * * * *");
			setTimezone(browserTimezone());
			setTimeoutSeconds(30);
			setDescription("");
		}
		setError(null);
	}, [open, job]);

	async function handleSubmit(e: React.FormEvent) {
		e.preventDefault();
		setSubmitting(true);
		setError(null);

		try {
			if (isEdit && job) {
				const res = await updateRunnerJob(slug, runnerId, job.id, {
					name: name.trim(),
					slug: jobSlug.trim(),
					cron: cron.trim(),
					timezone: timezone.trim() || browserTimezone(),
					timeout_seconds: timeoutSeconds,
					description: description.trim(),
				});
				onOpenChange(false);
				onSuccess(res);
			} else {
				const res = await createRunnerJob(slug, runnerId, {
					name: name.trim(),
					slug: jobSlug.trim(),
					cron: cron.trim(),
					timezone: timezone.trim() || browserTimezone(),
					timeout_seconds: timeoutSeconds,
					description: description.trim(),
				});
				onOpenChange(false);
				onSuccess(res);
			}
		} catch (err) {
			setError(
				err instanceof ApiError
					? err.message
					: translateError(err) ||
							(isEdit
								? m.runners_jobs_update_failed()
								: m.runners_jobs_create_failed()),
			);
		} finally {
			setSubmitting(false);
		}
	}

	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogContent className="sm:max-w-md">
				<form onSubmit={handleSubmit} className="flex flex-col gap-4">
					<DialogHeader>
						<DialogTitle className="text-lg">
							{isEdit ? m.runners_jobs_edit() : m.runners_jobs_add()}
						</DialogTitle>
						<DialogDescription>
							{isEdit ? m.runners_jobs_edit_hint() : m.runners_jobs_add_hint()}
						</DialogDescription>
					</DialogHeader>

					<div className="flex flex-col gap-3 py-2">
						<div className="flex flex-col gap-1.5">
							<Label htmlFor="job-form-name">{m.runners_jobs_name()}</Label>
							<Input
								id="job-form-name"
								required
								placeholder="Intranet Gateway Health Check"
								value={name}
								onChange={(e) => setName(e.target.value)}
								disabled={submitting}
							/>
						</div>

						<div className="flex flex-col gap-1.5">
							<Label htmlFor="job-form-slug">{m.runners_jobs_slug()}</Label>
							<Input
								id="job-form-slug"
								required
								placeholder="intranet-gateway"
								className="font-mono text-sm"
								value={jobSlug}
								onChange={(e) => setJobSlug(e.target.value)}
								disabled={submitting}
							/>
						</div>

						<div className="grid grid-cols-2 gap-3">
							<div className="flex flex-col gap-1.5">
								<Label htmlFor="job-form-cron">{m.beeps_cron()}</Label>
								<Input
									id="job-form-cron"
									required
									placeholder="*/5 * * * *"
									className="font-mono text-sm"
									value={cron}
									onChange={(e) => setCron(e.target.value)}
									disabled={submitting}
								/>
							</div>

							<div className="flex flex-col gap-1.5">
								<Label htmlFor="job-form-timeout">
									{m.runners_jobs_timeout()}
								</Label>
								<Input
									id="job-form-timeout"
									type="number"
									min={1}
									max={3600}
									value={timeoutSeconds}
									onChange={(e) =>
										setTimeoutSeconds(Number(e.target.value) || 30)
									}
									disabled={submitting}
								/>
							</div>
						</div>

						<div className="flex flex-col gap-1.5">
							<Label htmlFor="job-form-tz">{m.runners_jobs_timezone()}</Label>
							<Input
								id="job-form-tz"
								placeholder="Asia/Shanghai"
								className="font-mono text-sm"
								value={timezone}
								onChange={(e) => setTimezone(e.target.value)}
								disabled={submitting}
							/>
						</div>

						<div className="flex flex-col gap-1.5">
							<div className="flex items-center justify-between">
								<Label htmlFor="job-form-desc">
									{m.runners_jobs_description()}
								</Label>
								<span className="text-[11px] text-muted-foreground">
									{m.common_optional()}
								</span>
							</div>
							<Input
								id="job-form-desc"
								placeholder={m.runners_jobs_description_placeholder()}
								value={description}
								onChange={(e) => setDescription(e.target.value)}
								disabled={submitting}
							/>
						</div>

						{error ? (
							<p className="text-sm text-destructive" role="alert">
								{error}
							</p>
						) : null}
					</div>

					<DialogFooter className="gap-2 sm:gap-0">
						<Button
							type="button"
							variant="outline"
							onClick={() => onOpenChange(false)}
							disabled={submitting}
						>
							{m.common_cancel()}
						</Button>
						<Button type="submit" disabled={submitting}>
							{submitting
								? m.common_saving()
								: isEdit
									? m.common_save_changes()
									: m.common_create()}
						</Button>
					</DialogFooter>
				</form>
			</DialogContent>
		</Dialog>
	);
}
