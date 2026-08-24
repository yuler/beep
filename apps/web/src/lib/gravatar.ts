/**
 * Synchronously computes SHA-256 hex string using standard SubtleCrypto.
 */
async function sha256(message: string): Promise<string> {
	const msgBuffer = new TextEncoder().encode(message.trim().toLowerCase());
	const hashBuffer = await crypto.subtle.digest("SHA-256", msgBuffer);
	const hashArray = Array.from(new Uint8Array(hashBuffer));
	return hashArray.map((b) => b.toString(16).padStart(2, "0")).join("");
}

export async function getGravatarUrl(
	email: string,
	size = 160,
): Promise<string> {
	const hash = await sha256(email);
	return `https://www.gravatar.com/avatar/${hash}?s=${size}&d=mp`;
}
