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
import { type Beeper, updateBeeper } from "@/lib/api/beepers";
import { ApiError } from "@/lib/api/client";

interface EditBeeperDialogProps {
	beeper: Beeper;
	slug: string;
	open: boolean;
	onOpenChange: (open: boolean) => void;
	onSuccess?: (updated: Beeper) => void;
}

export function EditBeeperDialog({
	beeper,
	slug,
	open,
	onOpenChange,
	onSuccess,
}: EditBeeperDialogProps) {
	const [editTitle, setEditTitle] = useState(beeper.title);
	const [editBody, setEditBody] = useState(beeper.body ?? "");
	const [editCron, setEditCron] = useState(beeper.cron);
	const [editInputs, setEditInputs] = useState<Record<string, unknown>>(
		beeper.config ?? {},
	);
	const [savingEdit, setSavingEdit] = useState(false);
	const [editError, setEditError] = useState<string | null>(null);

	const inputs = beeper.beeper_app?.inputs ?? [];

	useEffect(() => {
		if (!open) return;
		setEditTitle(beeper.title);
		setEditBody(beeper.body ?? "");
		setEditCron(beeper.cron);
		setEditInputs({ ...(beeper.config ?? {}) });
		setEditError(null);
	}, [open, beeper]);

	function handleOpenChange(isOpen: boolean) {
		if (!savingEdit) {
			onOpenChange(isOpen);
		}
	}

	async function handleSave(e: React.FormEvent) {
		e.preventDefault();
		setSavingEdit(true);
		setEditError(null);
		try {
			const updated = await updateBeeper(slug, beeper.id, {
				title: editTitle.trim(),
				body: editBody.trim() || null,
				cron: editCron.trim(),
				config: editInputs,
			});
			onOpenChange(false);
			onSuccess?.(updated);
		} catch (err) {
			setEditError(
				err instanceof ApiError ? err.message : "Failed to update beeper.",
			);
		} finally {
			setSavingEdit(false);
		}
	}

	return (
		<Dialog open={open} onOpenChange={handleOpenChange}>
			<DialogContent className="sm:max-w-lg">
				<form onSubmit={handleSave} className="flex flex-col gap-4">
					<DialogHeader>
						<DialogTitle className="text-lg">Edit Beeper</DialogTitle>
						<DialogDescription>
							Update the title, remarks, schedule, or configuration parameters
							for this beeper.
						</DialogDescription>
					</DialogHeader>

					<div className="flex flex-col gap-4 py-2">
						<div className="flex flex-col gap-2">
							<Label htmlFor="edit-beeper-title">Beeper Title</Label>
							<Input
								id="edit-beeper-title"
								required
								value={editTitle}
								onChange={(e) => setEditTitle(e.target.value)}
								disabled={savingEdit}
							/>
						</div>

						<div className="flex flex-col gap-2">
							<div className="flex items-center justify-between">
								<Label htmlFor="edit-beeper-body">Body / Remark</Label>
								<span className="text-[11px] text-muted-foreground">
									Optional
								</span>
							</div>
							<textarea
								id="edit-beeper-body"
								rows={3}
								placeholder="Add notes, runbook links, or alert context..."
								value={editBody}
								onChange={(e) => setEditBody(e.target.value)}
								disabled={savingEdit}
								className="w-full rounded-lg border border-input bg-transparent px-2.5 py-1.5 text-sm transition-colors outline-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:pointer-events-none disabled:cursor-not-allowed disabled:bg-input/50 disabled:opacity-50 dark:bg-input/30"
							/>
						</div>

						<div className="flex flex-col gap-2">
							<Label htmlFor="edit-beeper-cron">Cron Schedule</Label>
							<Input
								id="edit-beeper-cron"
								required
								value={editCron}
								onChange={(e) => setEditCron(e.target.value)}
								disabled={savingEdit}
								className="font-mono text-sm"
							/>
						</div>

						{inputs.length > 0 ? (
							<div className="flex flex-col gap-3 pt-2">
								<h3 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
									Configuration Parameters
								</h3>
								{inputs.map((input) => (
									<div key={input.name} className="flex flex-col gap-2">
										<Label htmlFor={`edit-input-${input.name}`}>
											{input.label}
											{input.required ? (
												<span className="text-destructive ml-1">*</span>
											) : null}
										</Label>
										<Input
											id={`edit-input-${input.name}`}
											type={input.type === "number" ? "number" : "text"}
											required={input.required}
											min={input.min}
											max={input.max}
											placeholder={input.placeholder}
											value={String(editInputs[input.name] ?? "")}
											onChange={(e) => {
												const raw = e.target.value;
												const val =
													input.type === "number"
														? raw === ""
															? ""
															: Number(raw)
														: raw;
												setEditInputs((curr) => ({
													...curr,
													[input.name]: val,
												}));
											}}
											disabled={savingEdit}
										/>
									</div>
								))}
							</div>
						) : null}

						<div className="rounded-lg border border-primary/20 bg-primary/5 p-3 text-xs leading-relaxed text-muted-foreground">
							<p className="font-medium text-foreground mb-1">
								💡 Impact of changes:
							</p>
							<ul className="list-disc pl-4 space-y-0.5">
								<li>
									<span className="font-medium text-foreground">
										Schedule (Cron):
									</span>{" "}
									The next execution time (
									<code className="font-mono text-[11px]">next_run_at</code>)
									will be immediately recalculated without needing to restart.
								</li>
								<li>
									<span className="font-medium text-foreground">
										Configuration:
									</span>{" "}
									New probe parameters take effect on the next scheduled run.
									Past run logs remain unchanged.
								</li>
							</ul>
						</div>

						{editError ? (
							<p className="text-sm text-destructive" role="alert">
								{editError}
							</p>
						) : null}
					</div>

					<DialogFooter className="gap-2 sm:gap-0">
						<Button
							type="button"
							variant="ghost"
							size="sm"
							onClick={() => onOpenChange(false)}
							disabled={savingEdit}
						>
							Cancel
						</Button>
						<Button
							type="submit"
							size="sm"
							disabled={savingEdit || !editTitle.trim() || !editCron.trim()}
						>
							{savingEdit ? "Saving…" : "Save Changes"}
						</Button>
					</DialogFooter>
				</form>
			</DialogContent>
		</Dialog>
	);
}
