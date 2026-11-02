const STORAGE_KEY = "theme";

export type Theme = "light" | "dark";
export type ThemePreference = "light" | "dark" | "system";
export type ThemeAnchor = Element | { clientX: number; clientY: number };

const THEME_COLOR_DARK = "#0a0a0a";
const THEME_COLOR_LIGHT = "#ffffff";
const PREFERENCE_ORDER: ThemePreference[] = ["dark", "light", "system"];

let mediaListenerAttached = false;

export function getStoredTheme(): ThemePreference {
	if (typeof window === "undefined") return "dark";
	const value = localStorage.getItem(STORAGE_KEY);
	if (value === "light" || value === "dark" || value === "system") {
		return value;
	}
	return "dark";
}

function getSystemTheme(): Theme {
	if (typeof window === "undefined" || !window.matchMedia) return "dark";
	return window.matchMedia("(prefers-color-scheme: light)").matches
		? "light"
		: "dark";
}

export function getResolvedTheme(
	pref: ThemePreference = getStoredTheme(),
): Theme {
	if (pref === "system") {
		return getSystemTheme();
	}
	return pref;
}

export function nextThemePreference(
	current: ThemePreference = getStoredTheme(),
): ThemePreference {
	const index = PREFERENCE_ORDER.indexOf(current);
	const from = index < 0 ? 0 : index;
	return PREFERENCE_ORDER[(from + 1) % PREFERENCE_ORDER.length];
}

function syncThemeColorMeta(theme: Theme) {
	let meta = document.querySelector(
		'meta[name="theme-color"]',
	) as HTMLMetaElement | null;
	if (!meta) {
		meta = document.createElement("meta");
		meta.name = "theme-color";
		document.head.appendChild(meta);
	}
	meta.content = theme === "dark" ? THEME_COLOR_DARK : THEME_COLOR_LIGHT;
}

function syncFavicon(theme: Theme) {
	const link = document.getElementById(
		"favicon-theme",
	) as HTMLLinkElement | null;
	if (link) {
		link.href = theme === "dark" ? "/icon-dark.svg" : "/icon-light.svg";
	}
}

function ensureSystemMediaListener() {
	if (
		mediaListenerAttached ||
		typeof window === "undefined" ||
		!window.matchMedia
	) {
		return;
	}
	mediaListenerAttached = true;
	const mq = window.matchMedia("(prefers-color-scheme: light)");
	const handler = () => {
		if (getStoredTheme() === "system") {
			applyTheme(getSystemTheme());
		}
	};
	mq.addEventListener?.("change", handler);
}

export function applyTheme(theme: Theme, syncColorScheme = true) {
	if (typeof document === "undefined") return;
	const root = document.documentElement;
	root.classList.toggle("dark", theme === "dark");
	root.dataset.theme = theme;
	if (syncColorScheme) root.style.colorScheme = theme;
	syncFavicon(theme);
	syncThemeColorMeta(theme);
}

function persistPreference(pref: ThemePreference, resolved: Theme) {
	if (typeof window === "undefined") return;
	localStorage.setItem(STORAGE_KEY, pref);
	ensureSystemMediaListener();
	window.dispatchEvent(
		new CustomEvent("beep:theme-change", {
			detail: { preference: pref, resolved },
		}),
	);
}

export function setTheme(pref: ThemePreference, syncColorScheme = true): Theme {
	const resolved = getResolvedTheme(pref);
	persistPreference(pref, resolved);
	applyTheme(resolved, syncColorScheme);
	return resolved;
}

type ThemeViewTransition = {
	finished: Promise<void>;
	skipTransition: () => void;
};

function startViewTransition(
	update: () => void,
): ThemeViewTransition | undefined {
	const start = document.startViewTransition?.bind(document);
	if (typeof start !== "function") return undefined;
	return start(update) as ThemeViewTransition;
}

function anchorCenter(el: Element | null | undefined) {
	if (!el || typeof el.getBoundingClientRect !== "function") return null;
	const rect = el.getBoundingClientRect();
	if (rect.width === 0 && rect.height === 0) return null;
	return { x: rect.left + rect.width / 2, y: rect.top + rect.height / 2 };
}

// Prefer the toggle's own rect. A raw mouse point, or a keyboard click at
// (0, 0), would start the wipe off the button.
function clickPoint(anchor?: ThemeAnchor) {
	const fallback = anchorCenter(
		document.querySelector("[data-theme-toggle]"),
	) ?? {
		x: window.innerWidth - 36,
		y: 28,
	};
	if (anchor instanceof Element) return anchorCenter(anchor) ?? fallback;
	if (
		anchor &&
		typeof anchor.clientX === "number" &&
		typeof anchor.clientY === "number" &&
		!(anchor.clientX === 0 && anchor.clientY === 0)
	) {
		return { x: anchor.clientX, y: anchor.clientY };
	}
	return fallback;
}

function prefersReducedMotion() {
	return window.matchMedia("(prefers-reduced-motion: reduce)").matches;
}

function animateThemeChange(resolvedNext: Theme, anchor?: ThemeAnchor) {
	const root = document.documentElement;
	if (root.classList.contains("theme-transitioning")) return;

	const { x, y } = clickPoint(anchor);
	const endRadius = Math.hypot(
		Math.max(x, window.innerWidth - x),
		Math.max(y, window.innerHeight - y),
	);
	const w = Math.max(window.innerWidth, 1);
	const h = Math.max(window.innerHeight, 1);
	const goingDark = resolvedNext === "dark";

	root.classList.add("theme-transitioning");
	if (goingDark) root.classList.add("theme-reveal-old");
	// Percentages, not px: some Chrome setups size the view-transition snapshot
	// box at 2x and scale it down, so px origins land at half (top center).
	root.style.setProperty("--vt-x", `${(x / w) * 100}%`);
	root.style.setProperty("--vt-y", `${(y / h) * 100}%`);
	root.style.setProperty(
		"--vt-r",
		`${(endRadius / (Math.hypot(w, h) / Math.SQRT2)) * 100}%`,
	);
	root.style.setProperty(
		"--vt-bg",
		goingDark ? THEME_COLOR_LIGHT : THEME_COLOR_DARK,
	);
	root.style.colorScheme = goingDark ? "light" : "dark";

	let timeout = 0;
	let finished = false;
	let transition: ThemeViewTransition | undefined;
	const finish = (skip = false) => {
		if (finished) return;
		finished = true;
		window.clearTimeout(timeout);
		if (skip) {
			try {
				transition?.skipTransition();
			} catch {
				// already skipped
			}
		}
		root.classList.remove("theme-transitioning", "theme-reveal-old");
		root.style.removeProperty("--vt-x");
		root.style.removeProperty("--vt-y");
		root.style.removeProperty("--vt-r");
		root.style.removeProperty("--vt-bg");
		applyTheme(resolvedNext);
	};
	timeout = window.setTimeout(() => finish(true), 800);

	try {
		transition = startViewTransition(() => {
			applyTheme(resolvedNext, false);
			// Native controls follow color-scheme, not author classes. The
			// incoming snapshot has to be captured with the new scheme.
			root.style.colorScheme = resolvedNext;
		});
	} catch {
		finish(true);
		return;
	}

	if (!transition) {
		finish(true);
		return;
	}

	transition.finished.then(() => finish(false)).catch(() => finish(true));
}

export function cycleTheme(anchor?: ThemeAnchor): ThemePreference {
	if (
		typeof document !== "undefined" &&
		document.documentElement.classList.contains("theme-transitioning")
	) {
		return getStoredTheme();
	}

	const current = getStoredTheme();
	const next = nextThemePreference(current);
	const from = getResolvedTheme(current);
	const to = getResolvedTheme(next);

	if (
		from === to ||
		prefersReducedMotion() ||
		typeof document.startViewTransition !== "function"
	) {
		setTheme(next);
		return next;
	}

	persistPreference(next, to);
	animateThemeChange(to, anchor);
	return next;
}

export function toggleTheme(anchor?: ThemeAnchor): ThemePreference {
	return cycleTheme(anchor);
}
