export function browserTimezone() {
	try {
		return Intl.DateTimeFormat().resolvedOptions().timeZone || "UTC";
	} catch {
		return "UTC";
	}
}

export function ianaTimezones() {
	try {
		if ("supportedValuesOf" in Intl) {
			return Intl.supportedValuesOf("timeZone");
		}
	} catch {
		// Fall through to UTC-only.
	}

	const detected = browserTimezone();
	if (detected !== "UTC") {
		return ["UTC", detected].sort();
	}

	return ["UTC"];
}
