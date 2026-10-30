import type * as React from "react";
import { Drawer as DrawerPrimitive } from "vaul";
import { cn } from "@/lib/utils";

function Drawer({
	shouldScaleBackground = false,
	...props
}: React.ComponentProps<typeof DrawerPrimitive.Root>) {
	return (
		<DrawerPrimitive.Root
			shouldScaleBackground={shouldScaleBackground}
			{...props}
		/>
	);
}

function DrawerTrigger({
	...props
}: React.ComponentProps<typeof DrawerPrimitive.Trigger>) {
	return <DrawerPrimitive.Trigger data-slot="drawer-trigger" {...props} />;
}

function DrawerPortal({
	...props
}: React.ComponentProps<typeof DrawerPrimitive.Portal>) {
	return <DrawerPrimitive.Portal data-slot="drawer-portal" {...props} />;
}

function DrawerClose({
	...props
}: React.ComponentProps<typeof DrawerPrimitive.Close>) {
	return <DrawerPrimitive.Close data-slot="drawer-close" {...props} />;
}

function DrawerOverlay({
	className,
	...props
}: React.ComponentProps<typeof DrawerPrimitive.Overlay>) {
	return (
		<DrawerPrimitive.Overlay
			data-slot="drawer-overlay"
			className={cn("fixed inset-0 z-50 bg-ground/80 duration-150", className)}
			{...props}
		/>
	);
}

function DrawerContent({
	className,
	children,
	...props
}: React.ComponentProps<typeof DrawerPrimitive.Content>) {
	return (
		<DrawerPortal>
			<DrawerOverlay />
			<DrawerPrimitive.Content
				data-slot="drawer-content"
				className={cn(
					"fixed inset-x-0 bottom-0 z-50 flex max-h-[90dvh] flex-col rounded-none border-t border-line bg-ground focus-visible:outline-none",
					className,
				)}
				{...props}
			>
				{children}
			</DrawerPrimitive.Content>
		</DrawerPortal>
	);
}

function DrawerHeader({
	className,
	...props
}: React.HTMLAttributes<HTMLDivElement>) {
	return (
		<div
			data-slot="drawer-header"
			className={cn(
				"grid gap-1.5 p-4 text-left border-b border-line",
				className,
			)}
			{...props}
		/>
	);
}

function DrawerFooter({
	className,
	...props
}: React.HTMLAttributes<HTMLDivElement>) {
	return (
		<div
			data-slot="drawer-footer"
			className={cn(
				"mt-auto flex flex-col gap-2 p-4 border-t border-line bg-transparent",
				className,
			)}
			{...props}
		/>
	);
}

function DrawerTitle({
	className,
	...props
}: React.ComponentProps<typeof DrawerPrimitive.Title>) {
	return (
		<DrawerPrimitive.Title
			data-slot="drawer-title"
			className={cn(
				"type-app-title font-medium leading-none tracking-tight text-ink",
				className,
			)}
			{...props}
		/>
	);
}

function DrawerDescription({
	className,
	...props
}: React.ComponentProps<typeof DrawerPrimitive.Description>) {
	return (
		<DrawerPrimitive.Description
			data-slot="drawer-description"
			className={cn("type-meta text-gray", className)}
			{...props}
		/>
	);
}

export {
	Drawer,
	DrawerPortal,
	DrawerOverlay,
	DrawerTrigger,
	DrawerClose,
	DrawerContent,
	DrawerHeader,
	DrawerFooter,
	DrawerTitle,
	DrawerDescription,
};
