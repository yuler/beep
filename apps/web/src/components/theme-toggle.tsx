import { Monitor, Moon, Sun } from "lucide-react";
import { useEffect, useState } from "react";

import { Button } from "@/components/ui/button";
import { m } from "@/lib/locale";
import { cycleTheme, getStoredTheme, type ThemePreference } from "@/lib/theme";

function preferenceLabel(preference: ThemePreference) {
	if (preference === "light") return m.theme_light();
	if (preference === "system") return m.theme_system();
	return m.theme_dark();
}

function PreferenceIcon({ preference }: { preference: ThemePreference }) {
	if (preference === "light") {
		return <Sun className="size-4" aria-hidden="true" />;
	}
	if (preference === "system") {
		return <Monitor className="size-4" aria-hidden="true" />;
	}
	return <Moon className="size-4" aria-hidden="true" />;
}

export function ThemeToggle({ className }: { className?: string }) {
	const [preference, setPreference] = useState<ThemePreference>("dark");

	useEffect(() => {
		setPreference(getStoredTheme());
		const handler = (event: Event) => {
			const detail = (event as CustomEvent<{ preference: ThemePreference }>)
				.detail;
			if (detail?.preference) {
				setPreference(detail.preference);
			}
		};
		window.addEventListener("beep:theme-change", handler);
		return () => window.removeEventListener("beep:theme-change", handler);
	}, []);

	const label = `${m.common_toggle_color_theme()}: ${preferenceLabel(preference)}`;

	return (
		<Button
			type="button"
			variant="ghost"
			size="icon"
			data-theme-toggle
			aria-label={label}
			title={label}
			className={className}
			onClick={(event) => {
				setPreference(cycleTheme(event.currentTarget));
			}}
		>
			<PreferenceIcon preference={preference} />
		</Button>
	);
}
