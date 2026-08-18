export const BEEP_CHANNELS = ["email", "web_push"] as const;

export type BeepChannel = (typeof BEEP_CHANNELS)[number];

export function toggleChannel(
	channels: BeepChannel[],
	channel: BeepChannel,
	enabled: boolean,
): BeepChannel[] {
	if (enabled) {
		if (channels.includes(channel)) return channels;
		return [...channels, channel];
	}
	return channels.filter((item) => item !== channel);
}
