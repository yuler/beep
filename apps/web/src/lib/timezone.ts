import { getTimeZones } from "@vvo/tzdb";

export type TimezoneOption = {
	value: string;
	flag: string;
	countryCode: string | null;
	countryName: string | null;
	search: string;
};

export function browserTimezone() {
	try {
		return Intl.DateTimeFormat().resolvedOptions().timeZone || "UTC";
	} catch {
		return "UTC";
	}
}

export function timezoneOptions(): TimezoneOption[] {
	return tzdbZones().map((zone) => fromTzdb(zone));
}

export function timezoneOption(value: string): TimezoneOption {
	const match = tzdbZones().find(
		(zone) => zone.name === value || zone.group.includes(value),
	);
	if (match) {
		return fromTzdb(match, value);
	}

	return {
		value,
		flag: "🌐",
		countryCode: null,
		countryName: null,
		search: value,
	};
}

function tzdbZones() {
	return getTimeZones({ includeUtc: true });
}

function fromTzdb(
	zone: ReturnType<typeof getTimeZones>[number],
	value = zone.name,
): TimezoneOption {
	const countryCode = zone.countryCode || null;
	const countryName = zone.countryName || null;
	const flag = countryCode ? regionFlag(countryCode) : "🌐";
	const search = [
		value,
		zone.name,
		zone.alternativeName,
		zone.abbreviation,
		countryName,
		countryCode,
		...zone.mainCities,
		...zone.group,
	]
		.filter(Boolean)
		.join(" ");

	return { value, flag, countryCode, countryName, search };
}

function regionFlag(countryCode: string) {
	const iso = countryCode.toUpperCase();
	if (!/^[A-Z]{2}$/.test(iso)) return "🌐";
	const points = [...iso].map((letter) => 0x1f1e6 + letter.charCodeAt(0) - 65);
	return String.fromCodePoint(...points);
}
