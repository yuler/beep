import { apiFetch } from "@/lib/api/client";

export type BeepRun = {
	id: string;
	scheduled_for: string;
	status:
		| "pending"
		| "running"
		| "succeeded"
		| "failed"
		| "skipped"
		| "expired";
	result: Record<string, unknown> | null;
	created_at: string;
};

export type Beep = {
	id: string;
	title: string;
	body: string | null;
	kind: "once" | "recurring";
	status: "active" | "paused" | "completed" | "cancelled" | "firing";
	run_at: string | null;
	next_run_at: string | null;
	last_run_at: string | null;
	timezone: string;
	created_at: string;
	runs: BeepRun[];
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
	body: {
		title: string;
		body: string | null;
		run_at: string;
	},
) {
	return apiFetch<Beep>(`/api/v1/${slug}/beeps`, {
		method: "POST",
		body,
	});
}

export type BeepProposal = {
	intent: "create" | "other";
	title: string | null;
	body: string | null;
	run_at: string | null;
	timezone: string;
	errors: {
		title?: string;
		body?: string;
		run_at?: string;
	};
	confirmable: boolean;
	message: string | null;
};

export function createBeepProposal(slug: string, prompt: string) {
	return apiFetch<BeepProposal>(`/api/v1/${slug}/beep_proposals`, {
		method: "POST",
		body: { prompt },
	});
}
