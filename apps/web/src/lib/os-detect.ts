export type SupportedOS = "macos" | "linux" | "windows";

export function detectUserOS(): SupportedOS {
	if (typeof navigator === "undefined") return "linux";
	const ua = navigator.userAgent || "";
	if (/Mac OS X|Macintosh/i.test(ua)) {
		return "macos";
	}
	if (/Windows/i.test(ua)) {
		return "windows";
	}
	if (/Linux|X11/i.test(ua)) {
		return "linux";
	}
	return "linux";
}

export function getCliInstallCommand(_os: SupportedOS): string {
	return "curl -fsSL https://raw.githubusercontent.com/yuler/beep/main/install.sh | bash";
}
