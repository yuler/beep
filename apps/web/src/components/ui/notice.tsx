import type * as React from "react";

import { cn } from "@/lib/utils";

export interface NoticeProps extends React.ComponentProps<"div"> {
	label?: string;
	error?: boolean;
}

export function Notice({
	className,
	label,
	error,
	children,
	...props
}: NoticeProps) {
	return (
		<div
			className={cn(
				"flex items-start gap-2.5 py-2 pl-3 text-sm text-ink",
				error ? "border-l border-ink" : "border-l border-line",
				className,
			)}
			{...props}
		>
			{label && (
				<span className="type-kicker text-gray shrink-0 select-none">
					{label}
				</span>
			)}
			<div className="flex-1 min-w-0">{children}</div>
		</div>
	);
}
