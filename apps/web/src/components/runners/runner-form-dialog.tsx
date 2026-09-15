import { KeyRound, Terminal } from "lucide-react";
import { useEffect, useState } from "react";
import { CliInstallSnippet } from "@/components/runners/cli-install-snippet";
import { Button } from "@/components/ui/button";
import { CopyableCode, useCopyToClipboard } from "@/components/ui/copy-button";
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
import { cn } from "@/lib/utils";
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
	const [tab, setTab] = useState<"connect" | "manual">("connect");
	const [name, setName] = useState("");
	const [tagsInput, setTagsInput] = useState("");
	const [submitting, setSubmitting] = useState(false);
	const [error, setError] = useState<string | null>(null);
	const { copied: connectCopied, copy: copyConnect } = useCopyToClipboard();

	useEffect(() => {
		if (!open) return;
		setTab("connect");
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
			<DialogContent className="sm:max-w-lg">
				{isEdit ? (
					<form onSubmit={handleSubmit} className="flex flex-col gap-4">
						<DialogHeader>
							<DialogTitle className="text-lg">
								{m.runners_edit_runner()}
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
								{submitting ? m.common_saving() : m.runners_update_runner()}
							</Button>
						</DialogFooter>
					</form>
				) : (
					<div className="flex flex-col gap-4">
						<DialogHeader>
							<DialogTitle className="text-lg">
								{m.runners_add_runner()}
							</DialogTitle>
							<DialogDescription>{m.runners_description()}</DialogDescription>
						</DialogHeader>

						<div className="flex items-center gap-1 rounded-lg border border-input bg-muted/40 p-1">
							<button
								type="button"
								onClick={() => setTab("connect")}
								className={cn(
									"flex-1 flex items-center justify-center gap-1.5 rounded-md px-3 py-1.5 text-xs font-medium transition-all cursor-pointer",
									tab === "connect"
										? "bg-background text-foreground shadow-xs"
										: "text-muted-foreground hover:text-foreground",
								)}
							>
								<Terminal className="size-3.5" />
								<span>{m.runners_tab_connect()}</span>
								<span className="ml-1 rounded-full bg-primary/10 px-1.5 py-0.2 text-[10px] font-semibold text-primary">
									{m.runners_tab_recommended()}
								</span>
							</button>
							<button
								type="button"
								onClick={() => setTab("manual")}
								className={cn(
									"flex-1 flex items-center justify-center gap-1.5 rounded-md px-3 py-1.5 text-xs font-medium transition-all cursor-pointer",
									tab === "manual"
										? "bg-background text-foreground shadow-xs"
										: "text-muted-foreground hover:text-foreground",
								)}
							>
								<KeyRound className="size-3.5" />
								<span>{m.runners_tab_manual()}</span>
							</button>
						</div>

						{tab === "connect" ? (
							<div className="flex flex-col gap-4 py-2">
								<p className="text-xs text-muted-foreground">
									{m.runners_connect_desc()}
								</p>

								<div className="flex flex-col gap-1.5">
									<span className="text-xs font-medium text-foreground">
										{m.runners_connect_step1()}
									</span>
									<CliInstallSnippet />
								</div>

								<div className="flex flex-col gap-1.5">
									<span className="text-xs font-medium text-foreground">
										{m.runners_connect_step2()}
									</span>
									<CopyableCode
										code="beep runner connect"
										copied={connectCopied}
										onCopy={() => copyConnect("beep runner connect")}
										label={m.runners_copy()}
									/>
								</div>

								<div className="rounded-md border bg-muted/30 p-3 text-xs text-muted-foreground leading-relaxed">
									💡 {m.runners_connect_hint()}
								</div>

								<DialogFooter className="pt-2">
									<Button
										type="button"
										onClick={() => onOpenChange(false)}
										className="w-full sm:w-auto"
									>
										{m.runners_done()}
									</Button>
								</DialogFooter>
							</div>
						) : (
							<form onSubmit={handleSubmit} className="flex flex-col gap-4">
								<p className="text-xs text-muted-foreground">
									{m.runners_manual_desc()}
								</p>

								<div className="flex flex-col gap-4 py-1">
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
										{submitting ? m.common_saving() : m.runners_create_token()}
									</Button>
								</DialogFooter>
							</form>
						)}
					</div>
				)}
			</DialogContent>
		</Dialog>
	);
}
