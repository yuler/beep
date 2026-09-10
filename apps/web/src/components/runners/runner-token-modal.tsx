import { Check, Copy, Terminal } from "lucide-react";
import { CliInstallSnippet } from "@/components/runners/cli-install-snippet";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import {
	CopyableCode,
	CopyButton,
	useCopyToClipboard,
} from "@/components/ui/copy-button";
import {
	Dialog,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
} from "@/components/ui/dialog";
import { publicApiOrigin } from "@/config";
import type { RunnerWithToken } from "@/lib/api/runners";
import { m } from "@/locale/paraglide/messages";

interface RunnerTokenModalProps {
	runner: RunnerWithToken | null;
	open: boolean;
	onOpenChange: (open: boolean) => void;
}

interface CodeSnippetProps {
	label: string;
	code: string;
	copied: boolean;
	onCopy: () => void;
}

function CodeSnippet({ label, code, copied, onCopy }: CodeSnippetProps) {
	return (
		<div className="flex flex-col gap-1.5 min-w-0 max-w-full">
			<div className="flex items-center justify-between gap-2">
				<span className="text-xs font-semibold text-foreground uppercase tracking-wider truncate">
					{label}
				</span>
				<Button
					variant="ghost"
					size="sm"
					className="h-7 px-2 text-xs gap-1 shrink-0"
					onClick={onCopy}
				>
					{copied ? (
						<>
							<Check className="size-3.5 text-emerald-500" />
							<span className="text-emerald-500">{m.runners_copied()}</span>
						</>
					) : (
						<>
							<Copy className="size-3.5" />
							<span>{m.runners_copy()}</span>
						</>
					)}
				</Button>
			</div>
			<CopyableCode
				code={code}
				copied={copied}
				onCopy={onCopy}
				label={m.runners_copy()}
			/>
		</div>
	);
}

export function RunnerTokenModal({
	runner,
	open,
	onOpenChange,
}: RunnerTokenModalProps) {
	const tokenCopy = useCopyToClipboard();
	const cliCopy = useCopyToClipboard();

	if (!runner || !runner.token) {
		return null;
	}

	const token = runner.token;
	const serverUrl = publicApiOrigin();

	const cliCmd = `beep runner config set --server ${serverUrl} --token ${token}
beep runner up`;

	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogContent className="sm:max-w-2xl max-h-[90vh] overflow-y-auto overflow-x-hidden min-w-0">
				<DialogHeader>
					<div className="flex items-center gap-2">
						<Terminal className="size-5 text-primary" />
						<DialogTitle className="text-lg">
							{m.runners_token_modal_title()} ({runner.name})
						</DialogTitle>
					</div>
					<DialogDescription>{m.runners_token_modal_desc()}</DialogDescription>
				</DialogHeader>

				<div className="flex flex-col gap-6 py-2 min-w-0 max-w-full overflow-x-hidden">
					<Alert className="border-amber-500/30 bg-amber-500/10 text-amber-900 dark:text-amber-200">
						<AlertTitle className="text-xs font-semibold uppercase tracking-wider">
							{m.common_tips()}
						</AlertTitle>
						<AlertDescription className="text-xs">
							{m.runners_token_warning()}
						</AlertDescription>
					</Alert>

					{/* Step 1: Install Beep CLI */}
					<div className="flex flex-col gap-2 min-w-0 max-w-full">
						<div className="flex flex-col gap-0.5">
							<span className="text-xs font-semibold text-foreground uppercase tracking-wider">
								{m.runners_install_cli_step()}
							</span>
							<p className="text-xs text-muted-foreground">
								{m.runners_install_cli_desc()}
							</p>
						</div>
						<CliInstallSnippet />
					</div>

					{/* Step 2: Configure & Start Runner */}
					<div className="flex flex-col gap-4 min-w-0 max-w-full border-t pt-4">
						<div className="flex flex-col gap-0.5">
							<span className="text-xs font-semibold text-foreground uppercase tracking-wider">
								{m.runners_connect_step()}
							</span>
						</div>

						{/* Token Block */}
						<div className="flex flex-col gap-1.5 min-w-0 max-w-full">
							<div className="flex items-center justify-between gap-2">
								<span className="text-xs font-semibold text-foreground uppercase tracking-wider truncate">
									Token
								</span>
								<Button
									variant="ghost"
									size="sm"
									className="h-7 px-2 text-xs gap-1 shrink-0"
									onClick={() => tokenCopy.copy(token)}
								>
									{tokenCopy.copied ? (
										<>
											<Check className="size-3.5 text-emerald-500" />
											<span className="text-emerald-500">
												{m.runners_copied()}
											</span>
										</>
									) : (
										<>
											<Copy className="size-3.5" />
											<span>{m.runners_copy()}</span>
										</>
									)}
								</Button>
							</div>
							<div className="relative group min-w-0 max-w-full rounded-lg border bg-muted/60 px-3 py-2 font-mono text-xs text-foreground select-all break-all pr-9">
								{token}
								<CopyButton
									copied={tokenCopy.copied}
									onCopy={() => tokenCopy.copy(token)}
									label={m.runners_copy()}
									className="top-1.5 right-1.5"
								/>
							</div>
						</div>

						{/* Direct Binary CLI */}
						<CodeSnippet
							label={m.runners_cli_command()}
							code={cliCmd}
							copied={cliCopy.copied}
							onCopy={() => cliCopy.copy(cliCmd)}
						/>
					</div>
				</div>

				<DialogFooter>
					<Button type="button" size="sm" onClick={() => onOpenChange(false)}>
						{m.runners_done()}
					</Button>
				</DialogFooter>
			</DialogContent>
		</Dialog>
	);
}
