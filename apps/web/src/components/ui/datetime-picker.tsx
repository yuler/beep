"use client";

import { addMinutes, format } from "date-fns";
import { ChevronDownIcon } from "lucide-react";
import * as React from "react";

import { Button } from "@/components/ui/button";
import { Calendar } from "@/components/ui/calendar";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
	Popover,
	PopoverContent,
	PopoverTrigger,
} from "@/components/ui/popover";
import { cn } from "@/lib/utils";

const QUICK_OFFSETS = [
	{ label: "5m", minutes: 5 },
	{ label: "10m", minutes: 10 },
	{ label: "30m", minutes: 30 },
	{ label: "1h", minutes: 60 },
	{ label: "2h", minutes: 120 },
	{ label: "3h", minutes: 180 },
	{ label: "1d", minutes: 60 * 24 },
	{ label: "1w", minutes: 60 * 24 * 7 },
] as const;

function toTimeValue(date: Date) {
	const pad = (value: number) => String(value).padStart(2, "0");
	return `${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`;
}

function applyDateKeepingTime(current: Date, selected: Date) {
	const next = new Date(selected);
	next.setHours(
		current.getHours(),
		current.getMinutes(),
		current.getSeconds(),
		0,
	);
	return next;
}

function applyTimeKeepingDate(current: Date, time: string) {
	const [hours = "0", minutes = "0", seconds = "0"] = time.split(":");
	const next = new Date(current);
	next.setHours(Number(hours), Number(minutes), Number(seconds), 0);
	return next;
}

function DateTimePicker({
	value,
	onChange,
	id,
	className,
	disabled,
}: {
	value: Date;
	onChange: (date: Date) => void;
	id?: string;
	className?: string;
	disabled?: boolean;
}) {
	const [open, setOpen] = React.useState(false);
	const dateId = id ? `${id}-date` : undefined;
	const timeId = id ? `${id}-time` : undefined;

	return (
		<div className={cn("flex flex-col gap-3", className)}>
			<div className="flex flex-col gap-2">
				<Label>Quick</Label>
				<div className="flex flex-wrap gap-1.5">
					{QUICK_OFFSETS.map((offset) => (
						<Button
							key={offset.label}
							type="button"
							size="xs"
							variant="outline"
							disabled={disabled}
							aria-label={`In ${offset.label}`}
							onClick={() => onChange(addMinutes(new Date(), offset.minutes))}
						>
							{offset.label}
						</Button>
					))}
				</div>
			</div>
			<div className="flex flex-row gap-4">
				<div className="flex min-w-0 flex-1 flex-col gap-2">
					<Label htmlFor={dateId}>Date</Label>
					<Popover open={open} onOpenChange={setOpen}>
						<PopoverTrigger
							render={
								<Button
									variant="outline"
									id={dateId}
									disabled={disabled}
									className="w-full justify-between font-normal"
								/>
							}
						>
							{format(value, "PPP")}
							<ChevronDownIcon data-icon="inline-end" />
						</PopoverTrigger>
						<PopoverContent
							className="w-auto overflow-hidden p-0"
							align="start"
						>
							<Calendar
								mode="single"
								selected={value}
								captionLayout="dropdown"
								defaultMonth={value}
								onSelect={(date) => {
									if (!date) return;
									onChange(applyDateKeepingTime(value, date));
									setOpen(false);
								}}
							/>
						</PopoverContent>
					</Popover>
				</div>
				<div className="flex w-32 shrink-0 flex-col gap-2">
					<Label htmlFor={timeId}>Time</Label>
					<Input
						type="time"
						id={timeId}
						step="1"
						required
						disabled={disabled}
						value={toTimeValue(value)}
						onChange={(event) =>
							onChange(applyTimeKeepingDate(value, event.target.value))
						}
						className="appearance-none bg-background [&::-webkit-calendar-picker-indicator]:hidden [&::-webkit-calendar-picker-indicator]:appearance-none"
					/>
				</div>
			</div>
		</div>
	);
}

export { DateTimePicker };
