import { CopyableCode, useCopyToClipboard } from "@/components/ui/copy-button";
import { GITHUB_REPO } from "@/config";
import { m } from "@/locale/paraglide/messages";

const INSTALL_COMMAND = `curl -fsSL https://raw.githubusercontent.com/${GITHUB_REPO}/main/install.sh | sh`;

export function CliInstallSnippet() {
	const { copied, copy } = useCopyToClipboard();

	return (
		<CopyableCode
			code={INSTALL_COMMAND}
			copied={copied}
			onCopy={() => copy(INSTALL_COMMAND)}
			label={m.runners_copy()}
		/>
	);
}
