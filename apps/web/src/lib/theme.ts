const STORAGE_KEY = "theme";

export type Theme = "light" | "dark";
export type ThemePreference = Theme | "system";

export function getStoredTheme(): Theme | null {
	const value = localStorage.getItem(STORAGE_KEY);
	return value === "light" || value === "dark" ? value : null;
}

export function getThemePreference(): ThemePreference {
	return getStoredTheme() ?? "system";
}

export function getSystemTheme(): Theme {
	return window.matchMedia("(prefers-color-scheme: dark)").matches
		? "dark"
		: "light";
}

export function getResolvedTheme(): Theme {
	return getStoredTheme() ?? getSystemTheme();
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
	document.documentElement.classList.toggle("dark", theme === "dark");
	document.documentElement.dataset.theme = theme;
	syncFavicon(theme);
}

export function setTheme(theme: Theme) {
	localStorage.setItem(STORAGE_KEY, theme);
	applyTheme(theme);
}

export function setThemePreference(preference: ThemePreference) {
	if (preference === "system") {
		localStorage.removeItem(STORAGE_KEY);
		applyTheme(getSystemTheme());
		return;
	}
	setTheme(preference);
}

export function toggleTheme(): Theme {
	const next = getResolvedTheme() === "dark" ? "light" : "dark";
	setTheme(next);
	return next;
}
