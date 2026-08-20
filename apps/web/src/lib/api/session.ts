import {
	type MeResponse,
	destroySession as serverDestroySession,
	fetchMe as serverFetchMe,
	rememberLastAccount as serverRememberLastAccount,
	startSession as serverStartSession,
	verifyMagicLink as serverVerifyMagicLink,
} from "@/server/session";

export type {
	MeResponse,
	StartSessionResponse,
	VerifySessionResponse,
} from "@/server/session";

/** Shared with router stale-time defaults and route guards. */
export const ME_STALE_MS = 30_000;
let meInflight: Promise<MeResponse> | null = null;
let meCached: { value: MeResponse; at: number } | null = null;

function invalidateMeCacheInternal() {
	meInflight = null;
	meCached = null;
}

export function invalidateMeCache() {
	invalidateMeCacheInternal();
}

export async function startSession(email: string) {
	return serverStartSession({ data: { email } });
}

/** Verifies the magic-link code; the session cookie is set by the server. */
export async function verifyMagicLink(code: string) {
	const result = await serverVerifyMagicLink({ data: { code } });
	invalidateMeCacheInternal();
	return result;
}

export function fetchMe(options?: { force?: boolean }): Promise<MeResponse> {
	// Server-side this module lives in the shared SSR process; never reuse the
	// module-level cache/in-flight state, which would leak one user's identity
	// to another across requests (each SSR request has its own cookie).
	if (import.meta.env.SSR) {
		return serverFetchMe({});
	}

	const force = options?.force === true;
	if (!force && meCached && Date.now() - meCached.at < ME_STALE_MS) {
		return Promise.resolve(meCached.value);
	}
	if (!force && meInflight) {
		return meInflight;
	}

	const request = serverFetchMe({})
		.then((me) => {
			meCached = { value: me, at: Date.now() };
			return me;
		})
		.catch((err) => {
			invalidateMeCacheInternal();
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

/** Guest-friendly session probe — CoreApiError / network errors become `null`. */
export async function fetchMeOrNull(): Promise<MeResponse | null> {
	try {
		return await fetchMe();
	} catch {
		return null;
	}
}

/** Ask Core to persist the last-account picker hint on the identity. */
export function rememberLastAccount(slug: string) {
	return serverRememberLastAccount({ data: { slug } });
}

export async function destroySession() {
	try {
		await serverDestroySession({});
	} finally {
		invalidateMeCacheInternal();
	}
}
