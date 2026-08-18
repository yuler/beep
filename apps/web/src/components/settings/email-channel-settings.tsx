import { useState } from "react";

import { Button } from "@/components/ui/button";
import {
	Card,
	CardContent,
	CardDescription,
	CardHeader,
	CardTitle,
} from "@/components/ui/card";
import { ApiError } from "@/lib/api/client";
import { updateSettings } from "@/lib/api/settings";

export function EmailChannelSettings({
	slug,
	enabled,
	onChanged,
}: {
	slug: string;
	enabled: boolean;
	onChanged: () => Promise<void> | void;
}) {
	const [pending, setPending] = useState(false);
	const [error, setError] = useState<string | null>(null);

	async function setEnabled(next: boolean) {
		setError(null);
		setPending(true);
		try {
			await updateSettings(slug, { email_channel_enabled: next });
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
				<CardTitle>Email reminders</CardTitle>
				<CardDescription>
					Send due beeps to the email you use to sign in. You can also turn this
					off from the link in each reminder.
				</CardDescription>
			</CardHeader>
			<CardContent className="flex flex-col gap-3">
				<p className="text-sm text-muted-foreground">
					{enabled
						? "Email is on for this personal workspace."
						: "Email is off. New and existing beeps that selected email will not be sent."}
				</p>
				<Button
					type="button"
					className="w-fit"
					variant={enabled ? "outline" : "default"}
					disabled={pending}
					onClick={() => void setEnabled(!enabled)}
				>
					{pending ? "Saving…" : enabled ? "Turn off email" : "Turn on email"}
				</Button>
				{error ? (
					<p className="text-sm text-destructive" role="alert">
						{error}
					</p>
				) : null}
			</CardContent>
		</Card>
	);
}
