import type * as React from "react";

import { cn } from "@/lib/utils";

export interface SegmentedItem<T extends string = string> {
	value: T;
	label: React.ReactNode;
	disabled?: boolean;
}

export interface SegmentedProps<T extends string = string> {
	value: T;
	onChange: (value: T) => void;
	items: SegmentedItem<T>[];
	className?: string;
}

export function Segmented<T extends string = string>({
	value,
	onChange,
	items,
	className,
}: SegmentedProps<T>) {
	return (
		<div
			role="tablist"
			className={cn(
				"inline-flex w-full items-stretch border border-line p-0 bg-transparent text-sm",
				className,
			)}
		>
			{items.map((item) => {
				const isSelected = item.value === value;
				return (
					<button
						key={item.value}
						type="button"
						role="tab"
						aria-selected={isSelected}
						disabled={item.disabled}
						onClick={() => onChange(item.value)}
						className={cn(
							"flex-1 min-h-[40px] px-3 py-2 text-center text-sm font-normal transition-colors cursor-pointer select-none border-r border-line last:border-r-0 outline-none focus-visible:outline-1 focus-visible:outline-ink",
							isSelected
								? "bg-solid text-solid-ink"
								: "bg-transparent text-ink hover:bg-wash active:bg-wash",
							item.disabled && "opacity-40 cursor-not-allowed",
						)}
					>
						{item.label}
					</button>
				);
			})}
		</div>
	);
}
