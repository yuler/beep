import fs from "node:fs";
import path from "node:path";
import { fileURLToPath } from "node:url";

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const messagesDir = path.resolve(__dirname, "../../../project.inlang/messages");

const files = fs.readdirSync(messagesDir).filter((f) => f.endsWith(".json"));

for (const file of files) {
	const filePath = path.join(messagesDir, file);
	const content = JSON.parse(fs.readFileSync(filePath, "utf-8"));
	const sorted = Object.keys(content)
		.sort()
		.reduce<Record<string, string>>((acc, key) => {
			acc[key] = content[key];
			return acc;
		}, {});

	fs.writeFileSync(filePath, `${JSON.stringify(sorted, null, 2)}\n`, "utf-8");
	console.log(`✨ Sorted ${file}`);
}
