const STORAGE_KEY = "theme";

export type Theme = "light" | "dark";
export type ThemePreference = "light" | "dark" | "system";

const THEME_COLOR_DARK = "#0a0a0a";
const THEME_COLOR_LIGHT = "#ffffff";

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

export function applyTheme(theme: Theme) {
	if (typeof document === "undefined") return;
	document.documentElement.classList.toggle("dark", theme === "dark");
	document.documentElement.dataset.theme = theme;
	syncFavicon(theme);
	syncThemeColorMeta(theme);
}

export function setTheme(pref: ThemePreference): Theme {
	if (typeof window !== "undefined") {
		localStorage.setItem(STORAGE_KEY, pref);
	}
	ensureSystemMediaListener();
	const resolved = getResolvedTheme(pref);
	applyTheme(resolved);
	if (typeof window !== "undefined") {
		window.dispatchEvent(
			new CustomEvent("beep:theme-change", {
				detail: { preference: pref, resolved },
			}),
		);
	}
	return resolved;
}

export function toggleTheme(): Theme {
	const next = getResolvedTheme() === "dark" ? "light" : "dark";
	return setTheme(next);
}
