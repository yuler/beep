import { Input as InputPrimitive } from "@base-ui/react/input";
import type * as React from "react";

import { cn } from "@/lib/utils";

function Input({ className, type, ...props }: React.ComponentProps<"input">) {
	return (
		<InputPrimitive
			type={type}
			data-slot="input"
			className={cn(
				"w-full min-w-0 rounded-none border-b border-line border-t-0 border-x-0 bg-transparent px-2 py-2 text-base text-ink placeholder:text-gray transition-colors outline-none disabled:pointer-events-none disabled:cursor-not-allowed disabled:opacity-40 sm:text-sm focus-visible:outline focus-visible:outline-1 focus-visible:outline-ink focus-visible:outline-offset-2",
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
