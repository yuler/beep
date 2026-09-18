type BeepScheduleFormatStyle = "full" | "short";

function partValue(
	parts: Intl.DateTimeFormatPart[],
	type: Intl.DateTimeFormatPartTypes,
) {
	return parts.find((part) => part.type === type)?.value ?? "";
}

function formatInTimeZone(date: Date, timeZone: string) {
	const parts = new Intl.DateTimeFormat("en-CA", {
		timeZone,
		year: "numeric",
		month: "2-digit",
		day: "2-digit",
		hour: "2-digit",
		minute: "2-digit",
		second: "2-digit",
		hourCycle: "h23",
	}).formatToParts(date);

	const year = partValue(parts, "year");
	const month = partValue(parts, "month");
	const day = partValue(parts, "day");
	const hour = partValue(parts, "hour");
	const minute = partValue(parts, "minute");
	const second = partValue(parts, "second");

	return `${year}-${month}-${day} ${hour}:${minute}:${second}`;
}

export function formatBeepScheduleTime(
	value: string | Date | null | undefined,
	timezone: string,
	_style: BeepScheduleFormatStyle = "full",
): string {
	if (!value) return "—";

	const date = value instanceof Date ? value : new Date(value);
	if (Number.isNaN(date.getTime())) return "—";

	try {
		return formatInTimeZone(date, timezone || "UTC");
	} catch {
		return formatInTimeZone(date, "UTC");
	}
}
