import { Check, Copy } from "lucide-react";
import { useRef, useState } from "react";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

export function CopyButton({
	copied,
	onCopy,
	label,
	className,
}: {
	copied: boolean;
	onCopy: () => void;
	label: string;
	className?: string;
}) {
	return (
		<Button
			variant="outline"
			size="icon-xs"
			className={cn(
				"absolute top-2 right-2 opacity-0 group-hover:opacity-100 focus:opacity-100 transition-opacity bg-background/90 hover:bg-background border shadow-xs",
				className,
			)}
			onClick={onCopy}
			aria-label={label}
			title={label}
		>
			{copied ? (
				<Check className="size-3 text-emerald-500" />
			) : (
				<Copy className="size-3" />
			)}
		</Button>
	);
}

export function CopyableCode({
	code,
	copied,
	onCopy,
	label,
	className,
}: {
	code: string;
	copied: boolean;
	onCopy: () => void;
	label: string;
	className?: string;
}) {
	return (
		<div
			className={cn(
				"relative group min-w-0 max-w-full rounded-lg border bg-muted/60",
				className,
			)}
		>
			<pre className="overflow-x-auto p-3 font-mono text-xs text-foreground whitespace-pre select-all min-w-0 max-w-full">
				{code}
			</pre>
			<CopyButton copied={copied} onCopy={onCopy} label={label} />
		</div>
	);
}

export function useCopyToClipboard() {
	const [copied, setCopied] = useState(false);
	const timeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

	function copy(text: string) {
		void navigator.clipboard.writeText(text);
		setCopied(true);
		if (timeoutRef.current) {
			clearTimeout(timeoutRef.current);
		}
		timeoutRef.current = setTimeout(() => setCopied(false), 2000);
	}

	return { copied, copy };
}
