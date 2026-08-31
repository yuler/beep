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
import { useTranslation } from "@/lib/i18n";
import { channelLabel, translateError } from "@/lib/i18n-labels";
import {
	NOTIFICATION_CHANNELS,
	type NotificationChannel,
	toggleChannel,
} from "@/lib/notification-channels";

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
	const { t, dict } = useTranslation();
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
					: translateError(dict, t, err) || t("beepers.update_failed"),
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
							{t("beepers.edit_beeper")}
						</DialogTitle>
						<DialogDescription>
							{t("beepers.edit_description")}
						</DialogDescription>
					</DialogHeader>

					<div className="flex flex-col gap-4 py-2">
						<div className="flex flex-col gap-2">
							<Label htmlFor="edit-beeper-title">{t("beepers.title")}</Label>
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
									{t("beepers.body_remark")}
								</Label>
								<span className="text-[11px] text-muted-foreground">
									{t("common.optional")}
								</span>
							</div>
							<textarea
								id="edit-beeper-body"
								rows={3}
								placeholder={t("beepers.body_placeholder")}
								value={editBody}
								onChange={(e) => setEditBody(e.target.value)}
								disabled={savingEdit}
								className="w-full rounded-lg border border-input bg-transparent px-2.5 py-1.5 text-sm transition-colors outline-none placeholder:text-muted-foreground focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:pointer-events-none disabled:cursor-not-allowed disabled:bg-input/50 disabled:opacity-50 dark:bg-input/30"
							/>
						</div>

						<div className="flex flex-col gap-2">
							<Label htmlFor="edit-beeper-cron">
								{t("beepers.cron_schedule")}
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
							<Label>{t("beepers.notification_channels")}</Label>
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
										{channelLabel(t, channel)}
									</Label>
								))}
							</div>
							<p className="text-[11px] text-muted-foreground">
								{t("beepers.notification_channels_hint")}
							</p>
						</div>

						{inputs.length > 0 ? (
							<div className="flex flex-col gap-3 pt-2">
								<h3 className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
									{t("beepers.configuration_parameters")}
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
								{t("beepers.edit_impact_title")}
							</p>
							<ul className="list-disc pl-4 space-y-0.5">
								<li>
									<span className="font-medium text-foreground">
										{t("beepers.edit_impact_schedule_title")}
									</span>{" "}
									{t("beepers.edit_impact_schedule_body", {
										field: "next_run_at",
									})}
								</li>
								<li>
									<span className="font-medium text-foreground">
										{t("beepers.edit_impact_config_title")}
									</span>{" "}
									{t("beepers.edit_impact_config_body")}
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
							{t("common.cancel")}
						</Button>
						<Button
							type="submit"
							size="sm"
							disabled={savingEdit || !editTitle.trim() || !editCron.trim()}
						>
							{savingEdit ? t("common.saving") : t("beepers.save_changes")}
						</Button>
					</DialogFooter>
				</form>
			</DialogContent>
		</Dialog>
	);
}
