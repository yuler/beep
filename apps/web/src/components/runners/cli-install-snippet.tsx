import { Check, Copy } from "lucide-react";
import { useState } from "react";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
	detectUserOS,
	getCliInstallCommand,
	type SupportedOS,
} from "@/lib/os-detect";
import { m } from "@/locale/paraglide/messages";

const OS_OPTIONS: Array<{ id: SupportedOS; label: string }> = [
	{ id: "macos", label: "macOS" },
	{ id: "linux", label: "Linux" },
	{ id: "windows", label: "Windows" },
];

export function CliInstallSnippet() {
	const detectedOS = detectUserOS();
	const [selectedOS, setSelectedOS] = useState<SupportedOS>(detectedOS);
	const [isCopied, setIsCopied] = useState(false);

	const command = getCliInstallCommand(selectedOS);

	function copyCommand() {
		void navigator.clipboard.writeText(command);
		setIsCopied(true);
		setTimeout(() => setIsCopied(false), 2000);
	}

	return (
		<div className="flex flex-col gap-2.5 min-w-0 max-w-full">
			<div className="flex flex-wrap items-center justify-between gap-2">
				<div className="flex items-center gap-1.5 p-0.5 rounded-lg bg-muted border">
					{OS_OPTIONS.map((os) => {
						const isSelected = selectedOS === os.id;
						const isDetected = detectedOS === os.id;
						return (
							<button
								key={os.id}
								type="button"
								onClick={() => setSelectedOS(os.id)}
								className={`inline-flex items-center gap-1.5 px-2.5 py-1 text-xs font-medium rounded-md transition-colors cursor-pointer ${
									isSelected
										? "bg-background text-foreground shadow-xs"
										: "text-muted-foreground hover:text-foreground"
								}`}
							>
								<span>{os.label}</span>
								{isDetected ? (
									<Badge
										variant="secondary"
										className="h-4 px-1 text-[10px] uppercase tracking-wider font-normal bg-primary/10 text-primary border-0"
									>
										{m.runners_current_os()}
									</Badge>
								) : null}
							</button>
						);
					})}
				</div>

				<Button
					variant="ghost"
					size="sm"
					className="h-7 px-2 text-xs gap-1 shrink-0"
					onClick={copyCommand}
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
					{command}
				</pre>
				<Button
					variant="outline"
					size="icon-xs"
					className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 focus:opacity-100 transition-opacity bg-background/90 hover:bg-background border shadow-xs"
					onClick={copyCommand}
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
