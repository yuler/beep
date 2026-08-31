export function shortId(id: string, length = 8) {
	return id.replace(/-/g, "").slice(0, length).toUpperCase();
}
