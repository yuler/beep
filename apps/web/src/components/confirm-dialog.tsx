import { useCallback, useEffect, useRef, useState } from "react";
import {
	AlertDialog,
	AlertDialogAction,
	AlertDialogCancel,
	AlertDialogContent,
	AlertDialogDescription,
	AlertDialogFooter,
	AlertDialogHeader,
	AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { m } from "@/locale/paraglide/messages";

export type ConfirmOptions = {
	title?: string;
	description: string;
	confirmLabel?: string;
	cancelLabel?: string;
	variant?: "default" | "destructive";
};

type ConfirmRequest = ConfirmOptions & {
	id: number;
};

type ConfirmFn = (options: string | ConfirmOptions) => Promise<boolean>;

let confirmImpl: ConfirmFn | null = null;
let requestId = 0;

function normalizeOptions(options: string | ConfirmOptions): ConfirmOptions {
	if (typeof options === "string") {
		return { description: options, variant: "destructive" };
	}
	return { variant: "destructive", ...options };
}

/**
 * Promise-based confirm that renders a shadcn AlertDialog via ConfirmDialogHost.
 * Drop-in replacement for `window.confirm` / `confirm`.
 */
export function confirm(options: string | ConfirmOptions): Promise<boolean> {
	if (!confirmImpl) {
		throw new Error(
			"confirm() called before ConfirmDialogHost mounted. Add <ConfirmDialogHost /> near the app root.",
		);
	}
	return confirmImpl(options);
}

export function ConfirmDialogHost() {
	const [request, setRequest] = useState<ConfirmRequest | null>(null);
	const resolveRef = useRef<((value: boolean) => void) | null>(null);
	const resultRef = useRef<boolean | null>(null);

	const settle = useCallback((value: boolean) => {
		const resolve = resolveRef.current;
		resolveRef.current = null;
		resultRef.current = null;
		setRequest(null);
		resolve?.(value);
	}, []);

	useEffect(() => {
		confirmImpl = (options) =>
			new Promise<boolean>((resolve) => {
				resolveRef.current?.(false);
				resolveRef.current = resolve;
				resultRef.current = null;
				requestId += 1;
				setRequest({ id: requestId, ...normalizeOptions(options) });
			});

		return () => {
			confirmImpl = null;
			resolveRef.current?.(false);
			resolveRef.current = null;
		};
	}, []);

	const open = request !== null;

	return (
		<AlertDialog
			open={open}
			onOpenChange={(nextOpen) => {
				if (!nextOpen) {
					settle(resultRef.current ?? false);
				}
			}}
		>
			<AlertDialogContent size="sm">
				<AlertDialogHeader>
					<AlertDialogTitle>
						{request?.title ?? m.common_confirm()}
					</AlertDialogTitle>
					<AlertDialogDescription>
						{request?.description}
					</AlertDialogDescription>
				</AlertDialogHeader>
				<AlertDialogFooter>
					<AlertDialogCancel
						onClick={() => {
							resultRef.current = false;
						}}
					>
						{request?.cancelLabel ?? m.common_cancel()}
					</AlertDialogCancel>
					<AlertDialogAction
						variant={request?.variant ?? "destructive"}
						onClick={() => {
							resultRef.current = true;
						}}
					>
						{request?.confirmLabel ?? m.common_confirm()}
					</AlertDialogAction>
				</AlertDialogFooter>
			</AlertDialogContent>
		</AlertDialog>
	);
}
