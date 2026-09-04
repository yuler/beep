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
	const allowExecFlag = runner.allow_exec ? " --allow-exec" : "";
	const allowExecEnv = runner.allow_exec
		? " \\\n  -e BEEP_ALLOW_EXEC=true"
		: "";
	const allowExecCompose = runner.allow_exec
		? "\n      - BEEP_ALLOW_EXEC=true"
		: "";

	const dockerRunCmd = `docker run -d --name beep-runner --restart=always \\
  -e BEEP_SERVER=${serverUrl} \\
  -e BEEP_RUNNER_TOKEN=${token}${allowExecEnv} \\
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
      - BEEP_RUNNER_TOKEN=${token}${allowExecCompose}

volumes:
  beep-runner-workspace:`;

	const cliCmd = `beep-runner run --server ${serverUrl} --token ${token} --workspace ~/.beep-runner${allowExecFlag}`;

	function copyToClipboard(key: string, text: string) {
		void navigator.clipboard.writeText(text);
		setCopiedKey(key);
		setTimeout(() => setCopiedKey(null), 2000);
	}

	return (
		<Dialog open={open} onOpenChange={onOpenChange}>
			<DialogContent className="sm:max-w-2xl max-h-[90vh] overflow-y-auto">
				<DialogHeader>
					<div className="flex items-center gap-2">
						<Terminal className="size-5 text-primary" />
						<DialogTitle className="text-lg">
							{m.runners_token_modal_title()} ({runner.name})
						</DialogTitle>
					</div>
					<DialogDescription>{m.runners_token_modal_desc()}</DialogDescription>
				</DialogHeader>

				<div className="flex flex-col gap-5 py-2">
					<Alert className="border-amber-500/30 bg-amber-500/10 text-amber-900 dark:text-amber-200">
						<AlertTitle className="text-xs font-semibold uppercase tracking-wider">
							{m.term_gravatar()} {/* safety margin fallback */}
							Important
						</AlertTitle>
						<AlertDescription className="text-xs">
							{m.runners_token_warning()}
						</AlertDescription>
					</Alert>

					{/* Token Block */}
					<div className="flex flex-col gap-1.5">
						<div className="flex items-center justify-between">
							<span className="text-xs font-semibold text-foreground uppercase tracking-wider">
								Token
							</span>
							<Button
								variant="ghost"
								size="sm"
								className="h-7 px-2 text-xs gap-1"
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
						<div className="rounded-lg border bg-muted/60 px-3 py-2 font-mono text-xs text-foreground select-all break-all">
							{token}
						</div>
					</div>

					{/* Docker Run Command */}
					<div className="flex flex-col gap-1.5">
						<div className="flex items-center justify-between">
							<span className="text-xs font-semibold text-foreground uppercase tracking-wider">
								{m.runners_docker_command()}
							</span>
							<Button
								variant="ghost"
								size="sm"
								className="h-7 px-2 text-xs gap-1"
								onClick={() => copyToClipboard("docker", dockerRunCmd)}
							>
								{copiedKey === "docker" ? (
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
						<pre className="rounded-lg border bg-muted/60 p-3 font-mono text-xs text-foreground overflow-x-auto whitespace-pre select-all">
							{dockerRunCmd}
						</pre>
					</div>

					{/* Docker Compose YAML */}
					<div className="flex flex-col gap-1.5">
						<div className="flex items-center justify-between">
							<span className="text-xs font-semibold text-foreground uppercase tracking-wider">
								{m.runners_docker_compose()}
							</span>
							<Button
								variant="ghost"
								size="sm"
								className="h-7 px-2 text-xs gap-1"
								onClick={() => copyToClipboard("compose", dockerComposeYaml)}
							>
								{copiedKey === "compose" ? (
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
						<pre className="rounded-lg border bg-muted/60 p-3 font-mono text-xs text-foreground overflow-x-auto whitespace-pre select-all">
							{dockerComposeYaml}
						</pre>
					</div>

					{/* Direct Binary CLI */}
					<div className="flex flex-col gap-1.5">
						<div className="flex items-center justify-between">
							<span className="text-xs font-semibold text-foreground uppercase tracking-wider">
								{m.runners_cli_command()}
							</span>
							<Button
								variant="ghost"
								size="sm"
								className="h-7 px-2 text-xs gap-1"
								onClick={() => copyToClipboard("cli", cliCmd)}
							>
								{copiedKey === "cli" ? (
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
						<pre className="rounded-lg border bg-muted/60 p-3 font-mono text-xs text-foreground overflow-x-auto whitespace-pre select-all">
							{cliCmd}
						</pre>
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
