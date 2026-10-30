const STORAGE_KEY = "theme";

export type Theme = "light" | "dark";

export function getStoredTheme(): Theme {
	if (typeof window === "undefined") return "dark";
	const value = localStorage.getItem(STORAGE_KEY);
	return value === "light" ? "light" : "dark";
}

export function getResolvedTheme(): Theme {
	return getStoredTheme();
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
	meta.content = theme === "dark" ? "#000000" : "#ffffff";
}

function syncFavicon(theme: Theme) {
	const link = document.getElementById(
		"favicon-theme",
	) as HTMLLinkElement | null;
	if (link) {
		link.href = theme === "dark" ? "/icon-dark.svg" : "/icon-light.svg";
	}
}

export function applyTheme(theme: Theme) {
	if (typeof document === "undefined") return;
	document.documentElement.classList.toggle("dark", theme === "dark");
	document.documentElement.dataset.theme = theme;
	syncFavicon(theme);
	syncThemeColorMeta(theme);
}

export function setTheme(theme: Theme) {
	if (typeof window !== "undefined") {
		localStorage.setItem(STORAGE_KEY, theme);
	}
	applyTheme(theme);
}

export function toggleTheme(): Theme {
	const next = getResolvedTheme() === "dark" ? "light" : "dark";
	setTheme(next);
	return next;
}
