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
import { channelLabel, translateError } from "@/lib/i18n-labels";
import {
	NOTIFICATION_CHANNELS,
	type NotificationChannel,
	toggleChannel,
} from "@/lib/notification-channels";
import * as m from "@/locale/paraglide/messages";

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
	const [editChannels, setEditChannels] = useState<NotificationChannel[]>(
		(beeper.notification_channels ?? []) as NotificationChannel[],
	);
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
		setEditChannels(
			(beeper.notification_channels ?? []) as NotificationChannel[],
		);
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
				notification_channels: editChannels,
				config: editInputs,
			});
			onOpenChange(false);
			onSuccess?.(updated);
		} catch (err) {
			setEditError(
				err instanceof ApiError
					? err.message
					: translateError(err) || m.beepers_update_failed(),
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
						<DialogTitle className="text-lg">
							{m.beepers_edit_beeper()}
						</DialogTitle>
						<DialogDescription>
							{m.beepers_edit_description()}
						</DialogDescription>
					</DialogHeader>

					<div className="flex flex-col gap-4 py-2">
						<div className="flex flex-col gap-2">
							<Label htmlFor="edit-beeper-title">{m.beepers_title()}</Label>
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
								<Label htmlFor="edit-beeper-body">
									{m.beepers_body_remark()}
								</Label>
								<span className="text-[11px] text-muted-foreground">
									{m.common_optional()}
								</span>
							</div>
							<textarea
								id="edit-beeper-body"
								rows={3}
								placeholder={m.beepers_body_placeholder()}
								value={editBody}
								onChange={(e) => setEditBody(e.target.value)}
								disabled={savingEdit}
								className="w-full rounded-lg border border-input bg-transparent px-2.5 py-1.5 text-sm transition-colors outline-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:pointer-events-none disabled:cursor-not-allowed disabled:bg-input/50 disabled:opacity-50 dark:bg-input/30"
							/>
						</div>

						<div className="flex flex-col gap-2">
							<Label htmlFor="edit-beeper-cron">
								{m.beepers_cron_schedule()}
							</Label>
							<Input
								id="edit-beeper-cron"
								required
								value={editCron}
								onChange={(e) => setEditCron(e.target.value)}
								disabled={savingEdit}
								className="font-mono text-sm"
							/>
						</div>

						<div className="flex flex-col gap-2">
							<Label>{m.beepers_notification_channels()}</Label>
							<div className="flex flex-col gap-2 rounded-lg border border-input p-3 dark:bg-input/20">
								{NOTIFICATION_CHANNELS.map((channel) => (
									<Label
										key={channel}
										className="flex items-center gap-2 font-normal cursor-pointer text-sm"
									>
										<input
											type="checkbox"
											className="size-4 accent-primary rounded"
											checked={editChannels.includes(channel)}
											disabled={savingEdit}
											onChange={(e) =>
												setEditChannels((curr) =>
													toggleChannel(curr, channel, e.target.checked),
												)
											}
										/>
										{channelLabel(channel)}
									</Label>
								))}
							</div>
							<p className="text-[11px] text-muted-foreground">
								{m.beepers_notification_channels_hint()}
							</p>
						</div>

						{inputs.length > 0 ? (
							<div className="flex flex-col gap-3 pt-2">
								<h3 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
									{m.beepers_configuration_parameters()}
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
								{m.beepers_edit_impact_title()}
							</p>
							<ul className="list-disc pl-4 space-y-0.5">
								<li>
									<span className="font-medium text-foreground">
										{m.beepers_edit_impact_schedule_title()}
									</span>{" "}
									{m.beepers_edit_impact_schedule_body({
										field: "next_run_at",
									})}
								</li>
								<li>
									<span className="font-medium text-foreground">
										{m.beepers_edit_impact_config_title()}
									</span>{" "}
									{m.beepers_edit_impact_config_body()}
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
							{m.common_cancel()}
						</Button>
						<Button
							type="submit"
							size="sm"
							disabled={savingEdit || !editTitle.trim() || !editCron.trim()}
						>
							{savingEdit ? m.common_saving() : m.beepers_save_changes()}
						</Button>
					</DialogFooter>
				</form>
			</DialogContent>
		</Dialog>
	);
}
