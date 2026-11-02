import { Input as InputPrimitive } from "@base-ui/react/input";
import type * as React from "react";

import { cn } from "@/lib/utils";

function Input({ className, type, ...props }: React.ComponentProps<"input">) {
	return (
		<InputPrimitive
			type={type}
			data-slot="input"
			className={cn(
				"w-full min-w-0 rounded-sm border border-line bg-transparent px-3 py-2 text-base text-ink placeholder:text-gray transition-colors outline-none disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-40 sm:text-sm focus-visible:border-ink focus-visible:outline focus-visible:outline-1 focus-visible:outline-ink focus-visible:outline-offset-2 aria-[invalid=true]:border-danger aria-[invalid=true]:focus-visible:outline-danger",
				className,
			)}
			{...props}
		/>
	);
}

function TextField({
	className,
	type,
	...props
}: React.ComponentProps<"input">) {
	return <Input className={className} type={type} {...props} />;
}

export { Input, TextField };
