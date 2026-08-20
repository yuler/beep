import { createServerFn } from "@tanstack/react-start";

import { coreFetch } from "@/server/core";

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

export const fetchBeeps = createServerFn({ method: "GET", strict: false })
	.validator((s: { slug: string }) => s)
	.handler(({ data }) =>
		coreFetch<BeepsResponse>(`/api/v1/${data.slug}/beeps`, { method: "GET" }),
	);

export const fetchBeep = createServerFn({ method: "GET", strict: false })
	.validator((s: { slug: string; beepId: string }) => s)
	.handler(({ data }) =>
		coreFetch<Beep>(`/api/v1/${data.slug}/beeps/${data.beepId}`, {
			method: "GET",
		}),
	);

export const createBeep = createServerFn({ method: "POST", strict: false })
	.validator(
		(s: { slug: string; title: string; body: string | null; run_at: string }) =>
			s,
	)
	.handler(({ data }) =>
		coreFetch<Beep>(`/api/v1/${data.slug}/beeps`, {
			method: "POST",
			body: { title: data.title, body: data.body, run_at: data.run_at },
		}),
	);
