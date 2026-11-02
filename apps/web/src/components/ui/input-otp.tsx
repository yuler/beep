import {
	type ChangeEvent,
	type KeyboardEvent,
	useEffect,
	useRef,
	useState,
} from "react";

import { cn } from "@/lib/utils";

function sanitizeCode(input: string, maxLength: number) {
	return input
		.toUpperCase()
		.replace(/O/g, "0")
		.replace(/[IL]/g, "1")
		.replace(/[^A-Z0-9]/g, "")
		.slice(0, maxLength);
}

export function OtpField({
	value,
	onChange,
	onComplete,
	length = 6,
	id,
	autoFocus,
	disabled,
	className,
	"aria-invalid": ariaInvalid,
}: {
	value: string;
	onChange: (value: string) => void;
	onComplete?: (code: string) => void;
	length?: number;
	id?: string;
	autoFocus?: boolean;
	disabled?: boolean;
	className?: string;
	"aria-invalid"?: boolean;
}) {
	const inputRef = useRef<HTMLInputElement | null>(null);
	const wasDisabledRef = useRef(Boolean(disabled));
	const [isFocused, setIsFocused] = useState(false);

	const code = sanitizeCode(value, length);

	useEffect(() => {
		if (wasDisabledRef.current && !disabled && autoFocus) {
			inputRef.current?.focus();
		}
		wasDisabledRef.current = Boolean(disabled);
	}, [disabled, autoFocus]);

	function moveCaretToEnd(el: HTMLInputElement | null) {
		if (!el) return;
		const len = el.value.length;
		try {
			el.setSelectionRange(len, len);
		} catch {
			// Ignore if input does not support selection range.
		}
	}

	function handleChange(e: ChangeEvent<HTMLInputElement>) {
		const next = sanitizeCode(e.target.value, length);
		onChange(next);
		if (next.length === length) {
			onComplete?.(next);
		}
	}

	function handleKeyDown(e: KeyboardEvent<HTMLInputElement>) {
		if (
			e.key === "ArrowLeft" ||
			e.key === "ArrowRight" ||
			e.key === "ArrowUp" ||
			e.key === "ArrowDown" ||
			e.key === "Home"
		) {
			e.preventDefault();
			moveCaretToEnd(e.currentTarget);
		}
	}

	const slots = Array.from({ length }, (_, index) => index);
	const activeIndex = Math.min(code.length, length - 1);

	return (
		<div
			className={cn("relative flex items-center justify-center", className)}
			onClick={() => {
				inputRef.current?.focus();
				moveCaretToEnd(inputRef.current);
			}}
		>
			{/* Real transparent input overlay */}
			<input
				ref={inputRef}
				id={id}
				type="text"
				inputMode="text"
				autoComplete="one-time-code"
				autoFocus={autoFocus}
				disabled={disabled}
				value={code}
				onChange={handleChange}
				onKeyDown={handleKeyDown}
				onFocus={(e) => {
					setIsFocused(true);
					moveCaretToEnd(e.currentTarget);
				}}
				onSelect={(e) => moveCaretToEnd(e.currentTarget)}
				onBlur={() => setIsFocused(false)}
				className="absolute inset-0 z-10 w-full h-full opacity-0 cursor-text disabled:cursor-not-allowed"
				aria-invalid={ariaInvalid}
			/>

			{/* 6 equal-width boxes */}
			<div className="flex items-center gap-2" aria-hidden="true">
				{slots.map((index) => {
					const char = code[index] ?? "";
					const isFilled = char.length > 0;
					const isActive = isFocused && index === activeIndex;

					let borderClass = "border-line";
					if (ariaInvalid) {
						borderClass = "border-danger text-danger";
					} else if (isActive) {
						borderClass = "border-ink";
					} else if (isFilled) {
						borderClass = "border-gray";
					}

					return (
						<div
							key={index}
							className={cn(
								"flex h-12 w-11 sm:h-13 sm:w-12 items-center justify-center rounded-sm border bg-transparent font-mono text-xl text-ink transition-colors select-none",
								borderClass,
							)}
						>
							{char}
							{isActive && !isFilled && (
								<span className="inline-block w-0.5 h-5 bg-ink animate-pulse" />
							)}
						</div>
					);
				})}
			</div>
		</div>
	);
}

export const InputOTP = OtpField;
