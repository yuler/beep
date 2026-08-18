import { Label } from "@/components/ui/label";
import { type BeepChannel, toggleChannel } from "@/lib/beep-channels";

const CHANNEL_LABELS: Record<BeepChannel, string> = {
	email: "Email",
	web_push: "Browser notifications",
};

export function BeepChannelFields({
	channels,
	onChange,
	disabled,
}: {
	channels: BeepChannel[];
	onChange: (channels: BeepChannel[]) => void;
	disabled?: boolean;
}) {
	return (
		<fieldset className="flex flex-col gap-2">
			<legend className="text-sm font-medium">Channels</legend>
			<p className="text-sm text-muted-foreground">
				Email is the reliable default. Browser notifications still fire when
				this device is subscribed.
			</p>
			{(["email", "web_push"] as const).map((channel) => (
				<Label key={channel} className="font-normal">
					<input
						type="checkbox"
						className="size-4 accent-primary"
						checked={channels.includes(channel)}
						disabled={disabled}
						onChange={(event) =>
							onChange(toggleChannel(channels, channel, event.target.checked))
						}
					/>
					{CHANNEL_LABELS[channel]}
				</Label>
			))}
		</fieldset>
	);
}
