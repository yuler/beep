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
import { updateSettings } from "@/lib/api/settings";
import { useTranslation } from "@/lib/i18n";
import { channelLabel, translateError } from "@/lib/i18n-labels";
import {
	NOTIFICATION_CHANNELS,
	type NotificationChannel,
	toggleChannel,
} from "@/lib/notification-channels";

export function NotificationChannelSettings({
	slug,
	channels,
	onChanged,
}: {
	slug: string;
	channels: NotificationChannel[];
	onChanged: () => Promise<void> | void;
}) {
	const { t, dict } = useTranslation();
	const [pending, setPending] = useState(false);
	const [error, setError] = useState<string | null>(null);

	async function setChannel(channel: NotificationChannel, enabled: boolean) {
		setError(null);
		setPending(true);
		try {
			await updateSettings(slug, {
				notification_channels: toggleChannel(channels, channel, enabled),
			});
			await onChanged();
		} catch (err) {
			setError(
				err instanceof ApiError ? err.message : translateError(dict, t, err),
			);
		} finally {
			setPending(false);
		}
	}

	return (
		<Card className="max-w-lg">
			<CardHeader>
				<CardTitle>{t("settings.notifications_title")}</CardTitle>
				<CardDescription>
					{t("settings.notifications_description")}
				</CardDescription>
			</CardHeader>
			<CardContent className="flex flex-col gap-3">
				<fieldset className="flex flex-col gap-2" disabled={pending}>
					<legend className="sr-only">{t("beeps.channels")}</legend>
					{NOTIFICATION_CHANNELS.map((channel) => (
						<Label key={channel} className="font-normal">
							<input
								type="checkbox"
								className="size-4 accent-primary"
								checked={channels.includes(channel)}
								disabled={pending}
								onChange={(event) =>
									void setChannel(channel, event.target.checked)
								}
							/>
							{channelLabel(t, channel)}
						</Label>
					))}
				</fieldset>
				{error ? (
					<p className="text-sm text-destructive" role="alert">
						{error}
					</p>
				) : null}
			</CardContent>
		</Card>
	);
}
