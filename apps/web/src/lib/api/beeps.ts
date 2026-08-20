import {
	createBeep as serverCreateBeep,
	fetchBeep as serverFetchBeep,
	fetchBeeps as serverFetchBeeps,
} from "@/server/beeps";

export type { Beep, BeepRun, BeepsResponse } from "@/server/beeps";

export function fetchBeeps(slug: string) {
	return serverFetchBeeps({ data: { slug } });
}

export function fetchBeep(slug: string, beepId: string) {
	return serverFetchBeep({ data: { slug, beepId } });
}

export function createBeep(
	slug: string,
	body: {
		title: string;
		body: string | null;
		run_at: string;
	},
) {
	return serverCreateBeep({ data: { slug, ...body } });
}
