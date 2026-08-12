import { Moon, Sun } from "lucide-react";
import { useEffect } from "react";

import { Button } from "@/components/ui/button";
import { getStoredTheme, toggleTheme } from "@/lib/theme";

export function ThemeToggle() {
	useEffect(() => {
		const media = window.matchMedia("(prefers-color-scheme: dark)");
		const onChange = () => {
			if (!getStoredTheme()) {
				document.documentElement.classList.toggle("dark", media.matches);
				document.documentElement.dataset.theme = media.matches
					? "dark"
					: "light";
			}
		};

		media.addEventListener("change", onChange);
		return () => media.removeEventListener("change", onChange);
	}, []);

	return (
		<Button
			type="button"
			variant="outline"
			size="icon"
			aria-label="Toggle color theme"
			onClick={() => toggleTheme()}
		>
			<Sun className="size-4 dark:hidden" />
			<Moon className="hidden size-4 dark:inline" />
		</Button>
	);
}
