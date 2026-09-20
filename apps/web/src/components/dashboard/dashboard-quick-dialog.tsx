import { Bell, Plus } from "lucide-react";
import { useState } from "react";
import { BeepQuickCreate } from "@/components/beeps/beep-quick-create";
import { Button } from "@/components/ui/button";
import {
	ResponsiveDialog,
	ResponsiveDialogBody,
	ResponsiveDialogContent,
	ResponsiveDialogDescription,
	ResponsiveDialogHeader,
	ResponsiveDialogTitle,
	ResponsiveDialogTrigger,
} from "@/components/ui/responsive-dialog";
import type { NotificationChannel } from "@/lib/notification-channels";

export function DashboardQuickDialog({
	slug,
	defaultChannels,
	onCreated,
}: {
	slug: string;
	defaultChannels?: NotificationChannel[];
	onCreated: () => Promise<void> | void;
}) {
	const [open, setOpen] = useState(false);

	async function handleCreated() {
		setOpen(false);
		await onCreated();
	}

	return (
		<ResponsiveDialog open={open} onOpenChange={setOpen}>
			<ResponsiveDialogTrigger
				render={<Button size="sm" className="gap-1.5" />}
			>
				<Plus className="size-4" />
				<span>New Beep</span>
			</ResponsiveDialogTrigger>
			<ResponsiveDialogContent className="sm:max-w-2xl">
				<ResponsiveDialogHeader>
					<ResponsiveDialogTitle className="flex items-center gap-2">
						<Bell className="size-4 text-primary" />
						Create New Beep
					</ResponsiveDialogTitle>
					<ResponsiveDialogDescription>
						Schedule a reminder, recurring notification, or webhook trigger.
					</ResponsiveDialogDescription>
				</ResponsiveDialogHeader>
				<ResponsiveDialogBody className="pb-6">
					<BeepQuickCreate
						slug={slug}
						defaultChannels={defaultChannels}
						onCreated={handleCreated}
					/>
				</ResponsiveDialogBody>
			</ResponsiveDialogContent>
		</ResponsiveDialog>
	);
}
