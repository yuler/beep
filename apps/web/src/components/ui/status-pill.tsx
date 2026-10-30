import { cn } from "@/lib/utils";

const STATUS_PILL_STYLES = {
	emerald: "border border-solid bg-solid text-solid-ink",
	amber: "border border-ink bg-transparent text-ink",
	rose: "border border-ink bg-transparent text-ink",
	muted: "border border-line bg-transparent text-gray",
	outline: "border border-line bg-transparent text-ink",
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
				"inline-flex h-5 items-center rounded-none px-2 text-xs font-normal",
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
			<span className="w-8 text-right text-xs tabular-nums text-gray">
				{clamped}%
			</span>
			<div className="h-1 flex-1 overflow-hidden rounded-none bg-line">
				<div
					className="h-full rounded-none bg-ink transition-[width]"
					style={{ width: `${clamped}%` }}
				/>
			</div>
		</div>
	);
}
