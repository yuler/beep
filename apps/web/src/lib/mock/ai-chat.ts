export type MockChatReply = {
	label: string;
	sub: string;
	time: string;
	body: string;
};

export const mockChatTabs = ["Beeps", "History"] as const;

export const mockStarterMessage = "What beeps do I have coming up this week?";

export const mockChatReplies: MockChatReply[] = [
	{
		label: "Beeps",
		sub: "Account data",
		time: "2s",
		body: "You have 3 upcoming reminders — call mom (tomorrow), standup notes (Friday), and renew license (next Monday).",
	},
	{
		label: "Schedule",
		sub: "Trend check",
		time: "1s",
		body: "Most beeps land on weekday mornings. Want me to batch new ones into a morning digest?",
	},
];

export const mockFineTuneTypes = ["Seasonal", "Classic", "Limited"] as const;
