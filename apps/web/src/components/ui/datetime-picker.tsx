"use client";

import { format } from "date-fns";
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
		<div className={cn("flex flex-row gap-4", className)}>
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
					<PopoverContent className="w-auto overflow-hidden p-0" align="start">
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
	);
}

export { DateTimePicker };
