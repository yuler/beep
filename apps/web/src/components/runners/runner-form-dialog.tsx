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
	createRunner,
	type Runner,
	type RunnerWithToken,
	updateRunner,
} from "@/lib/api/runners";
import { translateError } from "@/lib/i18n-labels";
import { m } from "@/locale/paraglide/messages";

interface RunnerFormDialogProps {
	slug: string;
	runner?: Runner | null;
	open: boolean;
	onOpenChange: (open: boolean) => void;
	onSuccess: (runner: Runner | RunnerWithToken) => void;
}

export function RunnerFormDialog({
	slug,
	runner,
	open,
	onOpenChange,
	onSuccess,
}: RunnerFormDialogProps) {
	const isEdit = Boolean(runner);
	const [name, setName] = useState("");
	const [tagsInput, setTagsInput] = useState("");
	const [submitting, setSubmitting] = useState(false);
	const [error, setError] = useState<string | null>(null);

	useEffect(() => {
		if (!open) return;
		if (runner) {
			setName(runner.name);
			setTagsInput((runner.tags || []).join(", "));
		} else {
			setName("");
			setTagsInput("");
		}
		setError(null);
	}, [open, runner]);

	async function handleSubmit(e: React.FormEvent) {
		e.preventDefault();
		setSubmitting(true);
		setError(null);

		const tags = tagsInput
			.split(",")
			.map((t) => t.trim())
			.filter(Boolean);

		try {
			if (isEdit && runner) {
				const res = await updateRunner(slug, runner.id, {
					name: name.trim(),
					tags,
				});
				onOpenChange(false);
				onSuccess(res.runner);
			} else {
				const res = await createRunner(slug, {
					name: name.trim(),
					tags,
				});
				onOpenChange(false);
				onSuccess(res.runner);
			}
		} catch (err) {
			setError(
				err instanceof ApiError
					? err.message
					: translateError(err) ||
							(isEdit ? m.runners_update_failed() : m.runners_create_failed()),
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
							{isEdit ? m.runners_edit_runner() : m.runners_add_runner()}
						</DialogTitle>
						<DialogDescription>{m.runners_description()}</DialogDescription>
					</DialogHeader>

					<div className="flex flex-col gap-4 py-2">
						<div className="flex flex-col gap-2">
							<Label htmlFor="runner-name">{m.runners_name()}</Label>
							<Input
								id="runner-name"
								required
								placeholder={m.runners_name_placeholder()}
								value={name}
								onChange={(e) => setName(e.target.value)}
								disabled={submitting}
							/>
						</div>

						<div className="flex flex-col gap-2">
							<div className="flex items-center justify-between">
								<Label htmlFor="runner-tags">{m.runners_tags()}</Label>
								<span className="text-[11px] text-muted-foreground">
									{m.common_optional()}
								</span>
							</div>
							<Input
								id="runner-tags"
								placeholder={m.runners_tags_placeholder()}
								value={tagsInput}
								onChange={(e) => setTagsInput(e.target.value)}
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
									? m.runners_update_runner()
									: m.runners_create_runner()}
						</Button>
					</DialogFooter>
				</form>
			</DialogContent>
		</Dialog>
	);
}
