import Markdown from "react-markdown";

const ALLOWED_ELEMENTS = ["p", "em", "strong", "a", "ul", "ol", "li", "br"];

function withHardBreaks(source: string) {
	return source.replace(/([^\n])\n(?!\n)/g, "$1  \n");
}

export function BeepMarkdown({ source }: { source: string }) {
	return (
		<div className="text-sm leading-relaxed [&_a]:text-primary [&_a]:underline [&_ol]:mt-2 [&_ol]:list-decimal [&_ol]:pl-5 [&_p]:mt-0 [&_p+_p]:mt-2 [&_ul]:mt-2 [&_ul]:list-disc [&_ul]:pl-5">
			<Markdown allowedElements={ALLOWED_ELEMENTS} unwrapDisallowed>
				{withHardBreaks(source)}
			</Markdown>
		</div>
	);
}
