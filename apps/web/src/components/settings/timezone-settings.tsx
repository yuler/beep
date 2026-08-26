import { useState } from "react";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { ApiError } from "@/lib/api/client";
import { type TimezoneSource, updateSettings } from "@/lib/api/settings";
import { ianaTimezones } from "@/lib/timezone";

function sourceHint(source: TimezoneSource | null, timezone: string | null) {
	if (!timezone) {
		return "Not set yet. Opening this workspace detects it from your browser.";
	}
	if (source === "manual") {
		return "Set for this workspace. Browser detection will not overwrite it.";
	}
	return "Detected from this browser. Choosing another zone locks it for this workspace.";
}

export function TimezoneSettings({
	slug,
	timezone,
	timezoneSource,
	onChanged,
}: {
	slug: string;
	timezone: string | null;
	timezoneSource: TimezoneSource | null;
	onChanged: () => Promise<void> | void;
}) {
	const [pending, setPending] = useState(false);
	const [error, setError] = useState<string | null>(null);
	const zones = ianaTimezones();
	const options =
		timezone && !zones.includes(timezone) ? [timezone, ...zones] : zones;

	async function onSelect(next: string) {
		if (!next || next === timezone) return;

		setError(null);
		setPending(true);
		try {
			await updateSettings(slug, { timezone: next });
			await onChanged();
		} catch (err) {
			setError(err instanceof ApiError ? err.message : "Something went wrong.");
		} finally {
			setPending(false);
		}
	}

	return (
		<Card className="max-w-lg">
			<CardHeader>
				<CardTitle>Timezone</CardTitle>
				<CardDescription>
					{sourceHint(timezoneSource, timezone)}
				</CardDescription>
			</CardHeader>
			<CardContent className="flex flex-col gap-3">
				<Label className="flex flex-col items-start gap-1.5 font-normal">
					<span className="text-sm font-medium">IANA timezone</span>
					<select
						className="h-8 w-full rounded-lg border border-input bg-transparent px-2.5 text-sm outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 disabled:opacity-50"
						value={timezone ?? ""}
						disabled={pending}
						onChange={(event) => void onSelect(event.target.value)}
					>
						{timezone ? null : (
							<option value="" disabled>
								Not set
							</option>
						)}
						{options.map((zone) => (
							<option key={zone} value={zone}>
								{zone}
							</option>
						))}
					</select>
				</Label>
				{error ? (
					<p className="text-sm text-destructive" role="alert">
						{error}
					</p>
				) : null}
			</CardContent>
		</Card>
	);
}
