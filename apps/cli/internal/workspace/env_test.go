package workspace

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"beep/internal/envx"
)

func TestParseEnv(t *testing.T) {
	input := `
# Comment line
FOO=bar
export BAZ="hello world"
QUOTED_SINGLE='single quoted # not comment'
WITH_COMMENT=simple_value # inline comment
ESCAPED="line1\nline2\t\"quoted\""
EMPTY=
SPACED = trimmed 
UNCLOSED="oops
TRAILING="ok" junk
AFTER=kept
`

	got := envx.New()
	parseInto(got, input)
	expected := []string{
		"FOO=bar",
		"BAZ=hello world",
		"QUOTED_SINGLE=single quoted # not comment",
		"WITH_COMMENT=simple_value",
		"ESCAPED=line1\nline2\t\"quoted\"",
		"EMPTY=",
		"SPACED=trimmed",
		"AFTER=kept",
	}

	if !reflect.DeepEqual(got.Slice(), expected) {
		t.Fatalf("expected:\n%#v\ngot:\n%#v", expected, got.Slice())
	}
}

func TestLoadEnv(t *testing.T) {
	root := t.TempDir()
	ws, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}

	env, err := ws.LoadEnv()
	if err != nil {
		t.Fatal(err)
	}
	if env != nil {
		t.Fatalf("expected nil env for empty workspace, got %v", env)
	}

	envContent := `
GLOBAL_KEY=from_env
OVERRIDE_KEY=old_value
`
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte(envContent), 0o600); err != nil {
		t.Fatal(err)
	}

	env1, err := ws.LoadEnv()
	if err != nil {
		t.Fatal(err)
	}
	expected1 := []string{
		"GLOBAL_KEY=from_env",
		"OVERRIDE_KEY=old_value",
	}
	if !reflect.DeepEqual(env1, expected1) {
		t.Fatalf("expected:\n%#v\ngot:\n%#v", expected1, env1)
	}

	envLocalContent := `
OVERRIDE_KEY=new_local_value
LOCAL_ONLY=secret
`
	if err := os.WriteFile(filepath.Join(root, ".env.local"), []byte(envLocalContent), 0o600); err != nil {
		t.Fatal(err)
	}

	env2, err := ws.LoadEnv()
	if err != nil {
		t.Fatal(err)
	}
	expected2 := []string{
		"GLOBAL_KEY=from_env",
		"OVERRIDE_KEY=new_local_value",
		"LOCAL_ONLY=secret",
	}
	if !reflect.DeepEqual(env2, expected2) {
		t.Fatalf("expected:\n%#v\ngot:\n%#v", expected2, env2)
	}
}

func TestLoadEnvUnreadable(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root can read chmod 0 files")
	}

	root := t.TempDir()
	ws, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(root, ".env")
	if err := os.WriteFile(path, []byte("SECRET=should-not-load\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o600) })

	env, err := ws.LoadEnv()
	if err == nil {
		t.Fatalf("expected error for unreadable .env, got env %v", env)
	}
	if env != nil {
		t.Fatalf("expected nil env on error, got %v", env)
	}
}
