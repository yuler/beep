import { createServerFn } from "@tanstack/react-start";
import { setCookie } from "@tanstack/react-start/server";

import { type ApiError, coreFetch, coreFetchWithHeaders } from "@/server/core";

export type StartSessionResponse = {
	/** Still returned for non-browser clients; browsers rely on the pending cookie. */
	pending_authentication_token: string;
	/** Dev-only OTP from `X-Magic-Link-Code` (superset of core's flash/header). */
	code?: string;
};

export type VerifySessionResponse = {
	/** Still returned for non-browser clients; browsers rely on the session cookie. */
	session_token: string;
};

export type MeResponse = {
	identity: {
		id: string;
		email: string;
		name: string;
		staff: boolean;
	};
	accounts: Array<{
		id: string;
		name: string;
		slug: string;
		personal: boolean;
	}>;
	last_account_slug: string | null;
};

/**
 * Forward one or more Set-Cookie headers onto the browser response.
 * Mode B: core is behind web, so session cookies land on the web origin and are
 * sent back with every subsequent request until signed out.
 */
function forwardCookies(headers: Headers) {
	const values = headers.getSetCookie?.() ?? [];
	if (values.length === 0) {
		const single = headers.get("set-cookie");
		if (single) values.push(single);
	}
	for (const value of values) {
		const [pair] = value.split(";");
		const eq = pair?.indexOf("=");
		if (pair && eq && eq > 0) {
			const name = pair.slice(0, eq);
			const val = pair.slice(eq + 1);
			setCookie(name, val);
		}
	}
}

export const startSession = createServerFn({ method: "POST", strict: false })
	.validator((s: { email: string }) => s)
	.handler(async ({ data }) => {
		const { data: body, headers } =
			await coreFetchWithHeaders<StartSessionResponse>("/api/v1/session", {
				method: "POST",
				body: { email: data.email },
			});
		forwardCookies(headers);
		const code = headers.get("X-Magic-Link-Code") ?? undefined;
		return code ? { ...body, code } : body;
	});

export const verifyMagicLink = createServerFn({
	method: "POST",
	strict: false,
})
	.validator((s: { code: string }) => s)
	.handler(async ({ data }) => {
		const { data: body, headers } =
			await coreFetchWithHeaders<VerifySessionResponse>(
				"/api/v1/session/magic_link",
				{
					method: "POST",
					body: { code: data.code },
				},
			);
		forwardCookies(headers);
		return body;
	});

export const destroySession = createServerFn({
	method: "POST",
	strict: false,
}).handler(async () => {
	const { headers } = await coreFetchWithHeaders<{ message: string }>(
		"/api/v1/session",
		{ method: "DELETE" },
	);
	forwardCookies(headers);
});

export const rememberLastAccount = createServerFn({
	method: "POST",
	strict: false,
})
	.validator((s: { slug: string }) => s)
	.handler(async ({ data }) => {
		return coreFetch<{ last_account_slug: string }>("/api/v1/me/last_account", {
			method: "PUT",
			body: { slug: data.slug },
		});
	});

export const fetchMe = createServerFn({ method: "GET", strict: false }).handler(
	async () => coreFetch<MeResponse>("/api/v1/me", { method: "GET" }),
);

export type { ApiError };
