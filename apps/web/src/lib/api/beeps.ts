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
	cron: string | null;
	run_at: string | null;
	next_run_at: string | null;
	last_run_at: string | null;
	timezone: string;
	notification_channels: string[];
	beeper_id?: string | null;
	beeper?: {
		slug: string;
		name: string;
	} | null;
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
		body?: string | null;
		kind?: "once" | "recurring";
		run_at?: string | null;
		cron?: string | null;
		timezone?: string;
		notification_channels?: string[];
	},
) {
	return apiFetch<Beep>(`/api/v1/${slug}/beeps`, {
		method: "POST",
		body,
	});
}

export function triggerBeepRun(slug: string, beepId: string) {
	return apiFetch<BeepRun>(`/api/v1/${slug}/beeps/${beepId}/runs`, {
		method: "POST",
	});
}

export function pauseBeep(slug: string, beepId: string) {
	return apiFetch<Beep>(`/api/v1/${slug}/beeps/${beepId}/pause`, {
		method: "POST",
	});
}

export function resumeBeep(slug: string, beepId: string) {
	return apiFetch<Beep>(`/api/v1/${slug}/beeps/${beepId}/pause`, {
		method: "DELETE",
	});
}

export function deleteBeep(slug: string, beepId: string) {
	return apiFetch<void>(`/api/v1/${slug}/beeps/${beepId}`, {
		method: "DELETE",
	});
}

export type BeepProposal = {
	intent: "create" | "other";
	kind?: "once" | "recurring";
	title: string | null;
	body: string | null;
	run_at: string | null;
	cron: string | null;
	timezone: string;
	errors: {
		title?: string;
		body?: string;
		run_at?: string;
		cron?: string;
	};
	confirmable: boolean;
	message: string | null;
};

export function createBeepProposal(
	slug: string,
	prompt: string,
	timezone?: string,
) {
	return apiFetch<BeepProposal>(`/api/v1/${slug}/beep_proposals`, {
		method: "POST",
		body: { prompt, timezone },
	});
}
