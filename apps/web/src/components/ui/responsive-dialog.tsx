import * as React from "react";
import {
	Dialog,
	DialogClose,
	DialogContent,
	DialogDescription,
	DialogFooter,
	DialogHeader,
	DialogTitle,
	DialogTrigger,
} from "@/components/ui/dialog";
import {
	Drawer,
	DrawerClose,
	DrawerContent,
	DrawerDescription,
	DrawerFooter,
	DrawerHeader,
	DrawerTitle,
	DrawerTrigger,
} from "@/components/ui/drawer";
import { useIsMobile } from "@/hooks/use-mobile";
import { cn } from "@/lib/utils";

function resolveClassName(className: unknown): string | undefined {
	return typeof className === "string" ? className : undefined;
}

function resolveStyle(style: unknown): React.CSSProperties | undefined {
	return style && typeof style !== "function"
		? (style as React.CSSProperties)
		: undefined;
}

const ResponsiveDialogContext = React.createContext<boolean | null>(null);

function useResponsiveDialogIsMobile() {
	const isMobile = React.useContext(ResponsiveDialogContext);
	if (isMobile === null) {
		throw new Error(
			"ResponsiveDialog components must be used within ResponsiveDialog.",
		);
	}
	return isMobile;
}

interface ResponsiveDialogProps {
	children: React.ReactNode;
	open?: boolean;
	onOpenChange?: (open: boolean) => void;
	shouldScaleBackground?: boolean;
}

function ResponsiveDialog({
	children,
	open,
	onOpenChange,
	shouldScaleBackground = true,
}: ResponsiveDialogProps) {
	const isMobile = useIsMobile();

	return (
		<ResponsiveDialogContext.Provider value={isMobile}>
			{isMobile ? (
				<Drawer
					open={open}
					onOpenChange={onOpenChange}
					shouldScaleBackground={shouldScaleBackground}
				>
					{children}
				</Drawer>
			) : (
				<Dialog open={open} onOpenChange={onOpenChange}>
					{children}
				</Dialog>
			)}
		</ResponsiveDialogContext.Provider>
	);
}

function ResponsiveDialogTrigger({
	children,
	...props
}: React.ComponentProps<typeof DialogTrigger>) {
	const isMobile = useResponsiveDialogIsMobile();
	if (isMobile) {
		return (
			<DrawerTrigger {...(props as React.ComponentProps<typeof DrawerTrigger>)}>
				{children}
			</DrawerTrigger>
		);
	}
	return <DialogTrigger {...props}>{children}</DialogTrigger>;
}

function ResponsiveDialogClose({
	children,
	...props
}: React.ComponentProps<typeof DialogClose>) {
	const isMobile = useResponsiveDialogIsMobile();
	if (isMobile) {
		return (
			<DrawerClose {...(props as React.ComponentProps<typeof DrawerClose>)}>
				{children}
			</DrawerClose>
		);
	}
	return <DialogClose {...props}>{children}</DialogClose>;
}

function ResponsiveDialogContent({
	className,
	children,
	showCloseButton = true,
	style,
	...props
}: React.ComponentProps<typeof DialogContent>) {
	const isMobile = useResponsiveDialogIsMobile();

	if (isMobile) {
		return (
			<DrawerContent
				className={cn(
					"max-h-[90dvh] flex flex-col focus-visible:outline-none",
					resolveClassName(className),
				)}
				style={resolveStyle(style)}
				{...(props as React.ComponentProps<typeof DrawerContent>)}
			>
				{children}
			</DrawerContent>
		);
	}

	return (
		<DialogContent
			className={cn(
				"max-h-[min(90dvh,calc(100vh-3rem))] flex flex-col overflow-hidden",
				className,
			)}
			showCloseButton={showCloseButton}
			{...props}
		>
			{children}
		</DialogContent>
	);
}

function ResponsiveDialogHeader({
	className,
	...props
}: React.HTMLAttributes<HTMLDivElement>) {
	const isMobile = useResponsiveDialogIsMobile();
	if (isMobile) {
		return (
			<DrawerHeader
				className={cn("shrink-0 px-4 pt-4 pb-2", className)}
				{...props}
			/>
		);
	}
	return <DialogHeader className={cn("shrink-0", className)} {...props} />;
}

function ResponsiveDialogFooter({
	className,
	...props
}: React.HTMLAttributes<HTMLDivElement>) {
	const isMobile = useResponsiveDialogIsMobile();
	if (isMobile) {
		return (
			<DrawerFooter
				className={cn("shrink-0 border-t bg-muted/30 p-4", className)}
				{...props}
			/>
		);
	}
	return (
		<DialogFooter
			className={cn(
				"shrink-0 border-t bg-muted/40 p-4 sm:justify-end",
				className,
			)}
			{...props}
		/>
	);
}

function ResponsiveDialogTitle({
	className,
	...props
}: React.ComponentProps<typeof DialogTitle>) {
	const isMobile = useResponsiveDialogIsMobile();
	if (isMobile) {
		return (
			<DrawerTitle
				className={resolveClassName(className)}
				{...(props as React.ComponentProps<typeof DrawerTitle>)}
			/>
		);
	}
	return <DialogTitle className={className} {...props} />;
}

function ResponsiveDialogDescription({
	className,
	...props
}: React.ComponentProps<typeof DialogDescription>) {
	const isMobile = useResponsiveDialogIsMobile();
	if (isMobile) {
		return (
			<DrawerDescription
				className={resolveClassName(className)}
				{...(props as React.ComponentProps<typeof DrawerDescription>)}
			/>
		);
	}
	return <DialogDescription className={className} {...props} />;
}

function ResponsiveDialogBody({
	className,
	children,
	...props
}: React.HTMLAttributes<HTMLDivElement>) {
	return (
		<div
			className={cn(
				"flex-1 min-h-0 overflow-y-auto px-4 py-2 sm:px-1 space-y-4",
				className,
			)}
			{...props}
		>
			{children}
		</div>
	);
}

export {
	ResponsiveDialog,
	ResponsiveDialogTrigger,
	ResponsiveDialogClose,
	ResponsiveDialogContent,
	ResponsiveDialogHeader,
	ResponsiveDialogFooter,
	ResponsiveDialogTitle,
	ResponsiveDialogDescription,
	ResponsiveDialogBody,
};
