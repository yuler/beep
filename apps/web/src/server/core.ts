import { getRequestHeader } from "@tanstack/react-start/server";

import { ApiError } from "@/lib/api/client";

/**
 * Mode B: the browser only ever talks to this TanStack Start (Node) server.
 * All Rails-core calls originate here. The session cookie is forwarded from the
 * incoming request so core sees the same identity the browser has.
 *
 * This module is server-only (imported from `createServerFn` bodies and SSR
 * loaders). It must never be imported from a client component.
 */

const CORE_INTERNAL_URL = process.env.CORE_INTERNAL_URL ?? "";

export type { ApiError };

type CoreOptions = Omit<RequestInit, "body"> & {
	body?: unknown;
};

async function coreRequest(
	path: string,
	{ body, headers, method = "GET", ...init }: CoreOptions = {},
): Promise<Response> {
	if (!CORE_INTERNAL_URL) {
		throw new Error(
			"CORE_INTERNAL_URL is required. Copy .env.example to .env.",
		);
	}

	const requestHeaders = new Headers(headers);
	if (body !== undefined) {
		requestHeaders.set("Content-Type", "application/json");
	}
	requestHeaders.set("Accept", "application/json");

	const host = getRequestHeader("host");
	const cookie = getRequestHeader("cookie");

	// First-party semantics for core: forward Origin so CSRF follows the same
	// rules the old dev proxy used. Server-to-server fetches carry no
	// Sec-Fetch-Site, which Rails' header-only protection treats as allowed.
	if (host) requestHeaders.set("Origin", `http://${host}`);

	if (cookie && !requestHeaders.has("Cookie")) {
		requestHeaders.set("Cookie", cookie);
	}

	const response = await fetch(`${CORE_INTERNAL_URL}${path}`, {
		...init,
		method,
		headers: requestHeaders,
		body: body === undefined ? undefined : JSON.stringify(body),
	});

	if (!response.ok) {
		let message = response.statusText || "Request failed";
		let code: string | undefined;
		try {
			const payload = (await response.json()) as {
				message?: string;
				error?: string;
				code?: string;
			};
			message = payload.message ?? payload.error ?? message;
			code = payload.code;
		} catch {
			// ignore JSON parse errors
		}
		throw new ApiError(response.status, message, code);
	}

	return response;
}

export async function coreFetch<T>(
	path: string,
	options: CoreOptions = {},
): Promise<T> {
	const response = await coreRequest(path, options);
	return (await response.json()) as T;
}

export async function coreFetchWithHeaders<T>(
	path: string,
	options: CoreOptions = {},
): Promise<{ data: T; headers: Headers }> {
	const response = await coreRequest(path, options);
	return { data: (await response.json()) as T, headers: response.headers };
}
