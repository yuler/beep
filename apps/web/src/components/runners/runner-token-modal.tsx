import { Check, Copy, Terminal } from "lucide-react";
import { useState } from "react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
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
	snippetKey: string;
	copiedKey: string | null;
	onCopy: (key: string, text: string) => void;
}

function CodeSnippet({
	label,
	code,
	snippetKey,
	copiedKey,
	onCopy,
}: CodeSnippetProps) {
	const isCopied = copiedKey === snippetKey;

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
					onClick={() => onCopy(snippetKey, code)}
				>
					{isCopied ? (
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
			<div className="relative group min-w-0 max-w-full rounded-lg border bg-muted/60">
				<pre className="overflow-x-auto p-3 font-mono text-xs text-foreground whitespace-pre select-all min-w-0 max-w-full">
					{code}
				</pre>
				<Button
					variant="outline"
					size="icon-xs"
					className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 focus:opacity-100 transition-opacity bg-background/90 hover:bg-background border shadow-xs"
					onClick={() => onCopy(snippetKey, code)}
					aria-label={m.runners_copy()}
					title={m.runners_copy()}
				>
					{isCopied ? (
						<Check className="size-3 text-emerald-500" />
					) : (
						<Copy className="size-3" />
					)}
				</Button>
			</div>
		</div>
	);
}

export function RunnerTokenModal({
	runner,
	open,
	onOpenChange,
}: RunnerTokenModalProps) {
	const [copiedKey, setCopiedKey] = useState<string | null>(null);

	if (!runner || !runner.token) {
		return null;
	}

	const serverUrl = publicApiOrigin();
	const token = runner.token;
	const dockerRunCmd = `docker run -d --name beep-runner --restart=always \\
  -e BEEP_SERVER=${serverUrl} \\
  -e BEEP_RUNNER_TOKEN=${token} \\
  -v beep-runner-workspace:/home/beep/.beep-runner \\
  ghcr.io/yuler/beep-runner:latest`;

	const dockerComposeYaml = `services:
  beep-runner:
    image: ghcr.io/yuler/beep-runner:latest
    container_name: beep-runner
    restart: always
    volumes:
      - beep-runner-workspace:/home/beep/.beep-runner
    environment:
      - BEEP_SERVER=${serverUrl}
      - BEEP_RUNNER_TOKEN=${token}

volumes:
  beep-runner-workspace:`;

	const cliCmd = `beep-runner config set --server ${serverUrl} --token ${token}
beep-runner up`;

	function copyToClipboard(key: string, text: string) {
		void navigator.clipboard.writeText(text);
		setCopiedKey(key);
		setTimeout(() => setCopiedKey(null), 2000);
	}

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

				<div className="flex flex-col gap-5 py-2 min-w-0 max-w-full overflow-x-hidden">
					<Alert className="border-amber-500/30 bg-amber-500/10 text-amber-900 dark:text-amber-200">
						<AlertTitle className="text-xs font-semibold uppercase tracking-wider">
							{m.common_tips()}
						</AlertTitle>
						<AlertDescription className="text-xs">
							{m.runners_token_warning()}
						</AlertDescription>
					</Alert>

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
								onClick={() => copyToClipboard("token", token)}
							>
								{copiedKey === "token" ? (
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
							<Button
								variant="outline"
								size="icon-xs"
								className="absolute top-1.5 right-1.5 opacity-0 group-hover:opacity-100 focus:opacity-100 transition-opacity bg-background/90 hover:bg-background border shadow-xs"
								onClick={() => copyToClipboard("token", token)}
								aria-label={m.runners_copy()}
								title={m.runners_copy()}
							>
								{copiedKey === "token" ? (
									<Check className="size-3 text-emerald-500" />
								) : (
									<Copy className="size-3" />
								)}
							</Button>
						</div>
					</div>

					{/* Docker Run Command */}
					<CodeSnippet
						label={m.runners_docker_command()}
						code={dockerRunCmd}
						snippetKey="docker"
						copiedKey={copiedKey}
						onCopy={copyToClipboard}
					/>

					{/* Docker Compose YAML */}
					<CodeSnippet
						label={m.runners_docker_compose()}
						code={dockerComposeYaml}
						snippetKey="compose"
						copiedKey={copiedKey}
						onCopy={copyToClipboard}
					/>

					{/* Direct Binary CLI */}
					<CodeSnippet
						label={m.runners_cli_command()}
						code={cliCmd}
						snippetKey="cli"
						copiedKey={copiedKey}
						onCopy={copyToClipboard}
					/>
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
