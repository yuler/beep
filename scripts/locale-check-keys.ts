import { readFile } from "node:fs/promises";

const JSON_MESSAGE_KEYS = [] as const;

async function readMessages(locale: "en" | "zh") {
	const raw = await readFile(`project.inlang/messages/${locale}.json`, "utf8");
	return JSON.parse(raw) as Record<string, string>;
}

const en = await readMessages("en");
const zh = await readMessages("zh");
const enKeys = Object.keys(en).sort();
const zhKeys = Object.keys(zh).sort();

const missingInZh = enKeys.filter((key) => !zhKeys.includes(key));
const missingInEn = zhKeys.filter((key) => !enKeys.includes(key));
const emptyValues = [...enKeys, ...zhKeys].filter((key, index, keys) => {
	if (keys.indexOf(key) !== index) return false;
	return en[key] === "" || zh[key] === "";
});

for (const key of JSON_MESSAGE_KEYS) {
	for (const [locale, messages] of [
		["en", en],
		["zh", zh],
	] as const) {
		try {
			JSON.parse(messages[key] ?? "");
		} catch {
			throw new Error(`${locale}.${key} is not valid JSON`);
		}
	}
}

if (missingInZh.length || missingInEn.length || emptyValues.length) {
	console.error(
		JSON.stringify({ missingInZh, missingInEn, emptyValues }, null, 2),
	);
	process.exit(1);
}

console.log(`Locale keys OK (${enKeys.length} keys)`);
