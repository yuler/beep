import { useMemo, useState } from "react";
import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import {
	Combobox,
	ComboboxContent,
	ComboboxEmpty,
	ComboboxInput,
	ComboboxItem,
	ComboboxList,
	ComboboxTrigger,
} from "@/components/ui/combobox";
import { Label } from "@/components/ui/label";
import { ApiError } from "@/lib/api/client";
import { type TimezoneSource, updateSettings } from "@/lib/api/settings";
import {
	browserTimezone,
	type TimezoneOption,
	timezoneOption,
	timezoneOptions,
} from "@/lib/timezone";

function sourceHint(source: TimezoneSource | null, timezone: string | null) {
	if (!timezone) {
		return "Not set yet. Opening this workspace detects it from your browser.";
	}
	if (source === "manual") {
		return "Set for this workspace. Browser detection will not overwrite it.";
	}
	return "Detected from this browser. Choosing another zone locks it for this workspace.";
}

function TimezoneLabel({ option }: { option: TimezoneOption }) {
	return (
		<span className="flex min-w-0 flex-1 items-center gap-2">
			<span className="text-base leading-none" aria-hidden>
				{option.flag}
			</span>
			<span className="truncate">{option.value}</span>
			{option.countryName ? (
				<span className="truncate text-xs text-muted-foreground">
					{option.countryName}
				</span>
			) : null}
		</span>
	);
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
	const items = useMemo(() => {
		const options = timezoneOptions();
		if (timezone && !options.some((item) => item.value === timezone)) {
			return [timezoneOption(timezone), ...options];
		}
		return options;
	}, [timezone]);
	const selected = items.find((item) => item.value === timezone) ?? null;

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

	async function onReset() {
		const detected = browserTimezone();
		if (detected === timezone) return;

		setError(null);
		setPending(true);
		try {
			await updateSettings(slug, {
				timezone: detected,
				timezone_source: "manual",
			});
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
				<div className="flex flex-col gap-1.5">
					<Label htmlFor="account-timezone">IANA timezone</Label>
					<Combobox
						id="account-timezone"
						items={items}
						value={selected}
						disabled={pending}
						autoHighlight
						itemToStringLabel={(item) => item.search}
						isItemEqualToValue={(item, value) => item.value === value.value}
						onValueChange={(next) => {
							if (next && typeof next === "object") {
								void onSelect(next.value);
							}
						}}
					>
						<ComboboxTrigger className="w-full">
							{selected ? (
								<TimezoneLabel option={selected} />
							) : (
								<span className="text-muted-foreground">Not set</span>
							)}
						</ComboboxTrigger>
						<ComboboxContent className="flex flex-col">
							<div className="border-b border-border p-1.5">
								<ComboboxInput placeholder="Search city, country, or zone" />
							</div>
							<ComboboxEmpty>No timezone found.</ComboboxEmpty>
							<ComboboxList>
								{(item: TimezoneOption) => (
									<ComboboxItem key={item.value} value={item}>
										<TimezoneLabel option={item} />
									</ComboboxItem>
								)}
							</ComboboxList>
						</ComboboxContent>
					</Combobox>
				</div>
				<div className="flex items-center justify-between gap-3">
					<p className="text-xs text-muted-foreground">
						Browser timezone: {browserTimezone()}
					</p>
					<Button
						type="button"
						variant="outline"
						size="sm"
						disabled={pending || timezone === browserTimezone()}
						onClick={() => void onReset()}
					>
						Reset to browser
					</Button>
				</div>
				{error ? (
					<p className="text-sm text-destructive" role="alert">
						{error}
					</p>
				) : null}
			</CardContent>
		</Card>
	);
}
