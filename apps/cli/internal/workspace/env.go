package workspace

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// LoadEnv loads environment variables from <Root>/.env and <Root>/.env.local (if present).
// Later files (.env.local) override earlier files (.env).
func (w *Workspace) LoadEnv() []string {
	if w == nil || w.Root == "" {
		return nil
	}

	envFile := filepath.Join(w.Root, ".env")
	envLocalFile := filepath.Join(w.Root, ".env.local")

	merged := make(map[string]string)
	var order []string

	applyFile := func(path string) {
		data, err := os.ReadFile(path)
		if err != nil {
			return
		}
		pairs := ParseEnv(string(data))
		for _, pair := range pairs {
			key, val, ok := strings.Cut(pair, "=")
			if !ok {
				continue
			}
			if _, exists := merged[key]; !exists {
				order = append(order, key)
			}
			merged[key] = val
		}
	}

	applyFile(envFile)
	applyFile(envLocalFile)

	if len(order) == 0 {
		return nil
	}

	out := make([]string, 0, len(order))
	for _, k := range order {
		out = append(out, fmt.Sprintf("%s=%s", k, merged[k]))
	}
	return out
}

// ParseEnv parses standard .env file content into KEY=VALUE pairs.
// Supports:
//   - KEY=VALUE and export KEY=VALUE
//   - Single and double quotes (with escape sequences in double quotes)
//   - Comments (#) and inline comments (e.g. KEY=VAL # comment)
//   - Empty lines and leading/trailing whitespace
func ParseEnv(content string) []string {
	var results []string
	seen := make(map[string]int) // key -> index in results

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

		parsedVal := parseEnvValue(strings.TrimSpace(val))

		item := fmt.Sprintf("%s=%s", key, parsedVal)
		if idx, exists := seen[key]; exists {
			results[idx] = item
		} else {
			seen[key] = len(results)
			results = append(results, item)
		}
	}

	return results
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
