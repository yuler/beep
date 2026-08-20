import { createServerFn } from "@tanstack/react-start";

import { coreFetch } from "@/server/core";

export type DevLetter = {
	id: string;
	sent_at: string | null;
	subject: string | null;
	to: string | null;
	from: string | null;
};

export type DevLettersResponse = {
	letters: DevLetter[];
};

export const fetchDevLetters = createServerFn({
	method: "GET",
	strict: false,
}).handler(async () =>
	coreFetch<DevLettersResponse>("/api/v1/dev/letters", { method: "GET" }),
);

export const deleteDevLetter = createServerFn({
	method: "POST",
	strict: false,
})
	.validator((s: { id: string }) => s)
	.handler(({ data }) =>
		coreFetch<void>(`/api/v1/dev/letters/${encodeURIComponent(data.id)}`, {
			method: "DELETE",
		}),
	);

export const clearDevLetters = createServerFn({
	method: "POST",
	strict: false,
}).handler(async () =>
	coreFetch<void>("/api/v1/dev/letters/clear", { method: "DELETE" }),
);
