package workspace

import (
	"errors"
	"log"
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
		if !errors.Is(err, os.ErrNotExist) {
			log.Printf("workspace: ignoring %s: %v", path, err)
		}
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
// Lines with invalid keys ([A-Za-z_][A-Za-z0-9_]* required), unterminated
// quotes, or trailing content after a closing quote (other than whitespace
// or a # comment) are skipped.
func parseEnv(content string) []string {
	o := envx.New()
	parseInto(o, content)
	return o.Slice()
}

func parseInto(dst *envx.Ordered, content string) {
	// strings.Split (not bufio.Scanner) so a single very long line can't
	// silently truncate the rest of the file (Scanner caps tokens at 64KB).
	for _, rawLine := range strings.Split(content, "\n") {
		line := strings.TrimSpace(rawLine)
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
		if !isValidEnvKey(key) {
			continue
		}

		parsed, ok := parseEnvValue(strings.TrimSpace(val))
		if !ok {
			continue
		}
		dst.Set(key, parsed)
	}
}

func isValidEnvKey(key string) bool {
	if key == "" {
		return false
	}
	for i := 0; i < len(key); i++ {
		c := key[i]
		valid := c == '_' || c >= 'A' && c <= 'Z' || c >= 'a' && c <= 'z' || i > 0 && c >= '0' && c <= '9'
		if !valid {
			return false
		}
	}
	return true
}

func parseEnvValue(raw string) (string, bool) {
	if raw == "" {
		return "", true
	}

	// Double-quoted string: "hello world"
	if strings.HasPrefix(raw, "\"") {
		var sb strings.Builder
		escaped := false
		closed := false
		end := -1

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
				end = i
				break
			} else {
				sb.WriteByte(ch)
			}
		}
		if !closed {
			return "", false
		}
		if !validQuotedRemainder(raw[end+1:]) {
			return "", false
		}
		return sb.String(), true
	}

	// Single-quoted string: 'hello world' (literal)
	if strings.HasPrefix(raw, "'") {
		idx := strings.Index(raw[1:], "'")
		if idx == -1 {
			return "", false
		}
		if !validQuotedRemainder(raw[1+idx+1:]) {
			return "", false
		}
		return raw[1 : 1+idx], true
	}

	// Unquoted: strip trailing comments (space followed by #)
	for i := 0; i < len(raw); i++ {
		if raw[i] == '#' && i > 0 && unicode.IsSpace(rune(raw[i-1])) {
			raw = raw[:i]
			break
		}
	}

	return strings.TrimSpace(raw), true
}

// validQuotedRemainder reports whether the text after a closing quote is
// empty or just a trailing comment.
func validQuotedRemainder(rest string) bool {
	rest = strings.TrimSpace(rest)
	return rest == "" || strings.HasPrefix(rest, "#")
}
