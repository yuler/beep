package workspace

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"beep/internal/envx"
)

// LoadEnv loads environment variables from <Root>/.env and <Root>/.env.local
// (if present). Later files (.env.local) override earlier files (.env).
func (w *Workspace) LoadEnv() []string {
	if w == nil || w.Root == "" {
		return nil
	}

	merged := envx.New()
	applyFile(merged, filepath.Join(w.Root, ".env"))
	applyFile(merged, filepath.Join(w.Root, ".env.local"))
	if merged.Len() == 0 {
		return nil
	}
	return merged.Slice()
}

func applyFile(dst *envx.Ordered, path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	parseInto(dst, string(data))
}

// parseEnv parses .env content into KEY=VALUE pairs in file order, where a
// later assignment to the same key wins.
//
// Supports:
//   - KEY=VALUE and export KEY=VALUE
//   - Single and double quotes (with escape sequences in double quotes)
//   - Comments (#) and inline comments (e.g. KEY=VAL # comment)
//   - Empty lines and leading/trailing whitespace
//
// Multi-line values are not supported; each KEY=VALUE must fit on one line.
func parseEnv(content string) []string {
	o := envx.New()
	parseInto(o, content)
	return o.Slice()
}

func parseInto(dst *envx.Ordered, content string) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "export") && len(line) > 6 && unicode.IsSpace(rune(line[6])) {
			line = strings.TrimSpace(line[6:])
		}

		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		key = strings.TrimSpace(key)
		if key == "" || strings.HasPrefix(key, "#") {
			continue
		}

		dst.Set(key, parseEnvValue(strings.TrimSpace(val)))
	}
}

func parseEnvValue(raw string) string {
	if raw == "" {
		return ""
	}

	// Double-quoted string: "hello world"
	if strings.HasPrefix(raw, "\"") {
		var sb strings.Builder
		escaped := false
		closed := false

		for i := 1; i < len(raw); i++ {
			ch := raw[i]
			if escaped {
				switch ch {
				case 'n':
					sb.WriteByte('\n')
				case 'r':
					sb.WriteByte('\r')
				case 't':
					sb.WriteByte('\t')
				case '\\':
					sb.WriteByte('\\')
				case '"':
					sb.WriteByte('"')
				default:
					sb.WriteByte('\\')
					sb.WriteByte(ch)
				}
				escaped = false
			} else if ch == '\\' {
				escaped = true
			} else if ch == '"' {
				closed = true
				break
			} else {
				sb.WriteByte(ch)
			}
		}
		if closed {
			return sb.String()
		}
	}

	// Single-quoted string: 'hello world' (literal)
	if strings.HasPrefix(raw, "'") {
		idx := strings.Index(raw[1:], "'")
		if idx != -1 {
			return raw[1 : 1+idx]
		}
	}

	// Unquoted: strip trailing comments (space followed by #)
	for i := 0; i < len(raw); i++ {
		if raw[i] == '#' && i > 0 && unicode.IsSpace(rune(raw[i-1])) {
			raw = raw[:i]
			break
		}
	}

	return strings.TrimSpace(raw)
}
