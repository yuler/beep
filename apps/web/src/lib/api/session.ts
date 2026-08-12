import { apiFetch, apiFetchWithHeaders } from "@/lib/api/client";

export type StartSessionResponse = {
	/** Still returned for non-browser clients; browsers rely on the pending cookie. */
	pending_authentication_token: string;
	/** Dev-only OTP from `X-Magic-Link-Code` (same as Core HTML flash/header). */
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

/** Shared with router stale-time defaults and route guards. */
export const ME_STALE_MS = 30_000;
let meInflight: Promise<MeResponse> | null = null;
let meCached: { value: MeResponse; at: number } | null = null;

export function invalidateMeCache() {
	meInflight = null;
	meCached = null;
}

export async function startSession(email: string) {
	const { data, headers } = await apiFetchWithHeaders<StartSessionResponse>(
		"/api/v1/session",
		{
			method: "POST",
			body: { email },
		},
	);
	const code = headers.get("X-Magic-Link-Code") ?? undefined;
	return code ? { ...data, code } : data;
}

/** Verifies the magic-link code; pending auth comes from the HttpOnly cookie. */
export async function verifyMagicLink(code: string) {
	const result = await apiFetch<VerifySessionResponse>(
		"/api/v1/session/magic_link",
		{
			method: "POST",
			body: { code },
		},
	);
	invalidateMeCache();
	return result;
}

export function fetchMe(options?: { force?: boolean }): Promise<MeResponse> {
	// Server-side this module lives in the shared SSR process; never reuse the
	// module-level cache/in-flight state, which would leak one user's identity
	// to another across requests (each SSR request has its own cookie).
	if (import.meta.env.SSR) {
		return apiFetch<MeResponse>("/api/v1/me", { method: "GET" });
	}

	const force = options?.force === true;
	if (!force && meCached && Date.now() - meCached.at < ME_STALE_MS) {
		return Promise.resolve(meCached.value);
	}
	if (!force && meInflight) {
		return meInflight;
	}

	const request = apiFetch<MeResponse>("/api/v1/me", {
		method: "GET",
	})
		.then((me) => {
			meCached = { value: me, at: Date.now() };
			return me;
		})
		.catch((err) => {
			invalidateMeCache();
			throw err;
		})
		.finally(() => {
			if (meInflight === request) {
				meInflight = null;
			}
		});

	meInflight = request;
	return request;
}

/** Guest-friendly session probe — 401 / network errors become `null`. */
export async function fetchMeOrNull(): Promise<MeResponse | null> {
	try {
		return await fetchMe();
	} catch {
		return null;
	}
}

/** Ask Core to persist the last-account picker hint on the identity. */
export function rememberLastAccount(slug: string) {
	return apiFetch<{ last_account_slug: string }>("/api/v1/me/last_account", {
		method: "PUT",
		body: { slug },
	});
}

export async function destroySession() {
	try {
		return await apiFetch<{ message: string }>("/api/v1/session", {
			method: "DELETE",
		});
	} finally {
		invalidateMeCache();
	}
}
