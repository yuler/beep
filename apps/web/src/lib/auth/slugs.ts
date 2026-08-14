/** Keep in sync with core AccountSlug::RESERVED_* (+ web-only globals). */
const RESERVED_FROM_CORE = [
	"admin",
	"api",
	"cable",
	"home",
	"hotwire-spark",
	"invitations",
	"join",
	"join_code",
	"landing",
	"letter_opener",
	"manifest",
	"my",
	"payment",
	"rails",
	"service-worker",
	"session",
	"settings",
	"subscription",
	"up",
	"users",
	"webhooks",
	"assets",
	"billing",
	"dev",
	"device",
	"help",
	"jobs",
	"landings",
	"login",
	"logout",
	"magic_link",
	"setup",
	"static",
	"support",
	"test",
	"web_push",
] as const;

const RESERVED_WEB = ["sign", "dashboard", "accounts"] as const;

export const RESERVED_SLUGS = new Set<string>([
	...RESERVED_FROM_CORE,
	...RESERVED_WEB,
]);

const SLUG_PATTERN = /^[a-zA-Z0-9_-]{4,16}$/;

export function isAccountSlug(value: string): boolean {
	return SLUG_PATTERN.test(value) && !RESERVED_SLUGS.has(value);
}
