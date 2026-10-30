import { type ChangeEvent, type KeyboardEvent, useRef, useState } from "react";

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
	const [isFocused, setIsFocused] = useState(false);

	const code = sanitizeCode(value, length);

	function handleChange(e: ChangeEvent<HTMLInputElement>) {
		const next = sanitizeCode(e.target.value, length);
		onChange(next);
		if (next.length === length) {
			onComplete?.(next);
		}
	}

	function handleKeyDown(e: KeyboardEvent<HTMLInputElement>) {
		if (e.key === "Backspace" && code.length > 0) {
			// standard backspace handled by onChange
		}
	}

	const slots = Array.from({ length }, (_, index) => index);
	const activeIndex = Math.min(code.length, length - 1);

	return (
		<div
			className={cn("relative flex items-center justify-center", className)}
			onClick={() => inputRef.current?.focus()}
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
				maxLength={length}
				onChange={handleChange}
				onKeyDown={handleKeyDown}
				onFocus={() => setIsFocused(true)}
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
						borderClass = "border-ink";
					} else if (isActive) {
						borderClass = "border-ink";
					} else if (isFilled) {
						borderClass = "border-gray";
					}

					return (
						<div
							key={index}
							className={cn(
								"flex h-12 w-11 sm:h-13 sm:w-12 items-center justify-center border bg-transparent font-mono text-xl text-ink transition-colors select-none",
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
