type BeepScheduleFormatStyle = "full" | "short";

const SCHEDULE_FORMAT_OPTIONS: Record<
	BeepScheduleFormatStyle,
	Intl.DateTimeFormatOptions
> = {
	full: {
		year: "numeric",
		month: "short",
		day: "numeric",
		hour: "2-digit",
		minute: "2-digit",
		hour12: false,
	},
	short: {
		month: "short",
		day: "numeric",
		hour: "2-digit",
		minute: "2-digit",
		hour12: false,
	},
};

export function formatBeepScheduleTime(
	value: string | Date | null | undefined,
	timezone: string,
	style: BeepScheduleFormatStyle = "full",
): string {
	if (!value) return "—";

	const date = new Date(value);
	const timeZone = timezone || "UTC";
	const options = { ...SCHEDULE_FORMAT_OPTIONS[style], timeZone };

	try {
		return new Intl.DateTimeFormat(undefined, options).format(date);
	} catch {
		return new Intl.DateTimeFormat(undefined, {
			...options,
			timeZone: "UTC",
		}).format(date);
	}
}
