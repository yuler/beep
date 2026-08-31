import { cn } from "@/lib/utils";

const STATUS_PILL_STYLES = {
	emerald:
		"bg-emerald-500/12 text-emerald-700 dark:bg-emerald-500/15 dark:text-emerald-400",
	amber:
		"bg-amber-500/12 text-amber-700 dark:bg-amber-500/15 dark:text-amber-400",
	rose: "bg-rose-500/12 text-rose-700 dark:bg-rose-500/15 dark:text-rose-400",
	muted: "bg-muted text-muted-foreground",
	outline: "border border-border bg-transparent text-muted-foreground",
} as const;

export type StatusPillTone = keyof typeof STATUS_PILL_STYLES;

export function StatusPill({
	label,
	tone = "muted",
	className,
}: {
	label: string;
	tone?: StatusPillTone;
	className?: string;
}) {
	return (
		<span
			className={cn(
				"inline-flex h-6 items-center rounded-full px-2.5 text-xs font-medium",
				STATUS_PILL_STYLES[tone],
				className,
			)}
		>
			{label}
		</span>
	);
}

export function ProgressBar({
	value,
	className,
}: {
	value: number;
	className?: string;
}) {
	const clamped = Math.max(0, Math.min(100, value));
	return (
		<div className={cn("flex min-w-24 items-center gap-2", className)}>
			<span className="w-8 text-right text-xs tabular-nums text-muted-foreground">
				{clamped}%
			</span>
			<div className="h-1 flex-1 overflow-hidden rounded-full bg-muted">
				<div
					className="h-full rounded-full bg-emerald-500 transition-[width]"
					style={{ width: `${clamped}%` }}
				/>
			</div>
		</div>
	);
}
