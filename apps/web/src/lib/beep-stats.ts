import type { Beep } from "@/lib/api/beeps";

export const BEEP_TIMEZONE = "Asia/Shanghai";

function dateKey(value: Date | string, timeZone = BEEP_TIMEZONE) {
	return new Intl.DateTimeFormat("en-CA", {
		timeZone,
		year: "numeric",
		month: "2-digit",
		day: "2-digit",
	}).format(new Date(value));
}

export function beepRunAt(beep: Beep) {
	return beep.next_run_at ?? beep.run_at;
}

export function beepStats(beeps: Beep[], now = new Date()) {
	const today = dateKey(now);
	let active = 0;
	let dueToday = 0;
	let firing = 0;

	for (const beep of beeps) {
		if (beep.status === "firing") firing += 1;
		if (beep.status === "active") active += 1;
		if (beep.status !== "active" && beep.status !== "firing") continue;
		const runAt = beepRunAt(beep);
		if (runAt && dateKey(runAt) === today) dueToday += 1;
	}

	return { active, dueToday, firing };
}

export function upcomingBeeps(beeps: Beep[], limit = 5) {
	return beeps
		.filter((beep) => beep.status === "active" || beep.status === "firing")
		.slice()
		.sort((left, right) => {
			const leftAt = beepRunAt(left);
			const rightAt = beepRunAt(right);
			if (!leftAt) return 1;
			if (!rightAt) return -1;
			return new Date(leftAt).getTime() - new Date(rightAt).getTime();
		})
		.slice(0, limit);
}
