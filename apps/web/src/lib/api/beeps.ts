import { apiFetch } from "@/lib/api/client";

export type Beep = {
	id: string;
	message: string;
	kind: "once" | "recurring";
	status: "active" | "paused" | "completed" | "cancelled" | "firing";
	run_at: string | null;
	next_run_at: string | null;
	last_run_at: string | null;
	timezone: string;
	created_at: string;
};

export type BeepsResponse = {
	beeps: Beep[];
};

export function fetchBeeps(slug: string) {
	return apiFetch<BeepsResponse>(`/api/v1/${slug}/beeps`, {
		method: "GET",
	});
}

export function fetchBeep(slug: string, beepId: string) {
	return apiFetch<Beep>(`/api/v1/${slug}/beeps/${beepId}`, {
		method: "GET",
	});
}

export function createBeep(
	slug: string,
	body: { message: string; run_at: string },
) {
	return apiFetch<Beep>(`/api/v1/${slug}/beeps`, {
		method: "POST",
		body,
	});
}
