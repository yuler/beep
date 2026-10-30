import { Moon, Sun } from "lucide-react";

import { Button } from "@/components/ui/button";
import { toggleTheme } from "@/lib/theme";

export function ThemeToggle({ className }: { className?: string }) {
	return (
		<Button
			type="button"
			variant="ghost"
			size="icon"
			aria-label="Toggle theme"
			className={className}
			onClick={() => {
				toggleTheme();
			}}
		>
			<Sun className="hidden size-4 dark:block" />
			<Moon className="block size-4 dark:hidden" />
		</Button>
	);
}
