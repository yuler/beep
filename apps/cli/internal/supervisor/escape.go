package supervisor

import (
	"fmt"
	"html"
	"os"
	"strings"
)

func writePrivateFile(path string, data []byte) error {
	return os.WriteFile(path, data, 0o600)
}

func ServiceTitle(service string) string {
	if service == "" {
		return service
	}
	return strings.ToUpper(service[:1]) + service[1:]
}

func systemdQuote(s string) string {
	if strings.ContainsAny(s, " \t\n\"\\") {
		escaped := strings.ReplaceAll(s, `\`, `\\`)
		escaped = strings.ReplaceAll(escaped, `"`, `\"`)
		return `"` + escaped + `"`
	}
	return s
}

func systemdEnvironmentLine(key, value string) (string, bool) {
	if key == "" || strings.ContainsAny(key, "=\n") || strings.Contains(value, "\n") {
		return "", false
	}
	escaped := strings.ReplaceAll(value, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	escaped = strings.ReplaceAll(escaped, `$`, `$$`)
	return fmt.Sprintf(`"%s=%s"`, key, escaped), true
}

func xmlEscape(s string) string {
	return html.EscapeString(s)
}
