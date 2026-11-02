import { Monitor, Moon, Sun } from "lucide-react";
import { useEffect, useState } from "react";

import { Button } from "@/components/ui/button";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuRadioGroup,
	DropdownMenuRadioItem,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { m } from "@/lib/locale";
import { getStoredTheme, setTheme, type ThemePreference } from "@/lib/theme";

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

	function handleSelect(next: ThemePreference) {
		setPreference(next);
		setTheme(next);
	}

	return (
		<DropdownMenu>
			<DropdownMenuTrigger
				render={
					<Button
						type="button"
						variant="ghost"
						size="icon"
						aria-label={m.common_toggle_color_theme()}
						className={className}
					/>
				}
			>
				<Sun className="hidden size-4 dark:block" />
				<Moon className="block size-4 dark:hidden" />
			</DropdownMenuTrigger>
			<DropdownMenuContent align="end" className="min-w-36">
				<DropdownMenuRadioGroup
					value={preference}
					onValueChange={(val) => handleSelect(val as ThemePreference)}
				>
					<DropdownMenuRadioItem value="dark">
						<Moon className="size-4 text-gray" />
						<span>{m.theme_dark()}</span>
					</DropdownMenuRadioItem>
					<DropdownMenuRadioItem value="light">
						<Sun className="size-4 text-gray" />
						<span>{m.theme_light()}</span>
					</DropdownMenuRadioItem>
					<DropdownMenuRadioItem value="system">
						<Monitor className="size-4 text-gray" />
						<span>{m.theme_system()}</span>
					</DropdownMenuRadioItem>
				</DropdownMenuRadioGroup>
			</DropdownMenuContent>
		</DropdownMenu>
	);
}
