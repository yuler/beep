import { type NavigateOptions, redirect } from "@tanstack/react-router";

import { ApiError } from "@/lib/api/client";
import { fetchMeOrNull, type MeResponse } from "@/lib/api/session";
import {
	type AccountSummary,
	type PostAuthTarget,
	resolveDashboardTarget,
} from "@/lib/auth/account";
import { safeReturnTo } from "@/lib/auth/return-to";

export function redirectForTarget(target: PostAuthTarget): never {
	switch (target.kind) {
		case "href":
			throw redirect({ href: target.href });
		case "account":
			throw redirect({
				to: "/$account_slug",
				params: { account_slug: target.slug },
			});
		case "picker":
			throw redirect({ to: "/accounts" });
		case "sign":
			throw redirect({ to: "/sign" });
	}
}

export function redirectToSign(returnTo?: string | null): never {
	const safe = safeReturnTo(returnTo);
	throw redirect({
		to: "/sign",
		search: safe ? { return_to: safe } : {},
	});
}

export type RootRouteContext = {
	me: MeResponse | null;
};

/**
 * Mode B: the `session_id` cookie lives on the web origin, so every SSR request
 * carries it (forwarded to core by the server fn). The root `beforeLoad` has
 * already resolved `me` from the incoming cookie; a null `me` is genuinely
 * signed-out, on server and client alike.
 */
async function probeSession(
	context: RootRouteContext,
): Promise<MeResponse | null> {
	return context.me ?? (await fetchMeOrNull());
}

/**
 * Like core `Authentication#require_authentication` after `resume_session`.
 * Runs first on the server during SSR (cookie on the web origin), then on the
 * client for subsequent navigations. Redirects to /sign when signed-out.
 */
export async function requireSession({
	context,
	location,
}: {
	context: RootRouteContext;
	location: { pathname: string; searchStr: string };
}): Promise<MeResponse> {
	const me = await probeSession(context);
	if (!me) {
		redirectToSign(`${location.pathname}${location.searchStr}`);
	}
	return me;
}

/**
 * Like core `redirect_authenticated_user`.
 * Signed-in: `return_to` if safe, otherwise `/accounts` (picker / single dashboard).
 */
export async function requireGuest({
	context,
	search,
}: {
	context: RootRouteContext;
	search: { return_to?: string };
}) {
	const me = await probeSession(context);
	if (!me) return;
	const safe = safeReturnTo(search.return_to);
	if (safe) throw redirect({ href: safe });
	throw redirect({ to: "/accounts" });
}

export function requireStaff(me: {
	identity: { staff: boolean };
	accounts: AccountSummary[];
}) {
	if (!me.identity.staff) {
		redirectForTarget(resolveDashboardTarget(me.accounts));
	}
}

type LoaderContext = {
	location: { pathname: string; searchStr: string };
	params?: { account_slug?: string; [key: string]: string | undefined };
};

/**
 * Wrap a route loader so a 401 (stale/revoked session) or 403 (role change)
 * redirects like the removed `useAdminResource` hook did, instead of dumping a
 * raw error page through the default error boundary.
 */
export function withAuthRedirects<C extends LoaderContext, R>(
	load: (ctx: C) => Promise<R>,
) {
	return async (ctx: C): Promise<R> => {
		try {
			return await load(ctx);
		} catch (err) {
			if (err instanceof ApiError && err.status === 401) {
				redirectToSign(`${ctx.location.pathname}${ctx.location.searchStr}`);
			}
			if (err instanceof ApiError && err.status === 403) {
				throw redirect({ to: "/accounts" });
			}
			throw err;
		}
	};
}

type NavigateFn = (opts: NavigateOptions) => Promise<void> | void;

export function navigateForTarget(
	navigate: NavigateFn,
	target: PostAuthTarget,
) {
	switch (target.kind) {
		case "href":
			return navigate({ href: target.href });
		case "account":
			return navigate({
				to: "/$account_slug",
				params: { account_slug: target.slug },
			});
		case "picker":
			return navigate({ to: "/accounts" });
		case "sign":
			return navigate({ to: "/sign" });
	}
}

export type { AccountSummary };
