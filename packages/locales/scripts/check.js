import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const inlangDir = path.resolve(__dirname, "../project.inlang");
const settingsPath = path.join(inlangDir, "settings.json");
const messagesDir = path.join(inlangDir, "messages");

const settings = JSON.parse(fs.readFileSync(settingsPath, "utf-8"));
const locales = settings.languageTags || ["en", "zh-CN"];

console.log("🔍 Checking translation keys across locales:", locales.join(", "));

const dictionaries = {};
for (const locale of locales) {
	const filePath = path.join(messagesDir, `${locale}.json`);
	if (!fs.existsSync(filePath)) {
		console.error(`❌ Missing message file: ${filePath}`);
		process.exit(1);
	}
	dictionaries[locale] = JSON.parse(fs.readFileSync(filePath, "utf-8"));
}

const baseLocale = settings.sourceLanguageTag || "en";
const baseKeys = Object.keys(dictionaries[baseLocale] || {});
let hasErrors = false;

for (const locale of locales) {
	if (locale === baseLocale) continue;
	const currentKeys = new Set(Object.keys(dictionaries[locale] || {}));

	const missingKeys = baseKeys.filter((k) => !currentKeys.has(k));
	const extraKeys = [...currentKeys].filter((k) => !baseKeys.includes(k));

	if (missingKeys.length > 0) {
		hasErrors = true;
		console.error(`❌ [${locale}] Missing keys compared to [${baseLocale}]:`);
		for (const k of missingKeys) {
			console.error(`   - ${k}`);
		}
	}

	if (extraKeys.length > 0) {
		console.warn(`⚠️ [${locale}] Extra keys not in [${baseLocale}]:`);
		for (const k of extraKeys) {
			console.warn(`   + ${k}`);
		}
	}
}

if (hasErrors) {
	process.exit(1);
} else {
	console.log(`✅ All ${baseKeys.length} keys match perfectly across locales!`);
}
