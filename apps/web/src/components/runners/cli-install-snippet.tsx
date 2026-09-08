import { Check, Copy } from "lucide-react";
import { useState } from "react";
import { Button } from "@/components/ui/button";
import { m } from "@/locale/paraglide/messages";

const INSTALL_COMMAND =
	"curl -fsSL https://raw.githubusercontent.com/yuler/beep/main/install.sh | bash";

export function CliInstallSnippet() {
	const [isCopied, setIsCopied] = useState(false);

	function copyCommand() {
		void navigator.clipboard.writeText(INSTALL_COMMAND);
		setIsCopied(true);
		setTimeout(() => setIsCopied(false), 2000);
	}

	return (
		<div className="relative group min-w-0 max-w-full rounded-lg border bg-muted/60">
			<pre className="overflow-x-auto p-3 font-mono text-xs text-foreground whitespace-pre select-all min-w-0 max-w-full">
				{INSTALL_COMMAND}
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
	);
}
