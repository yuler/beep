package workspace

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
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
`

	got := parseEnv(input)
	expected := []string{
		"FOO=bar",
		"BAZ=hello world",
		"QUOTED_SINGLE=single quoted # not comment",
		"WITH_COMMENT=simple_value",
		"ESCAPED=line1\nline2\t\"quoted\"",
		"EMPTY=",
		"SPACED=trimmed",
	}

	if !reflect.DeepEqual(got, expected) {
		t.Fatalf("expected:\n%#v\ngot:\n%#v", expected, got)
	}
}

func TestLoadEnv(t *testing.T) {
	root := t.TempDir()
	ws, err := Open(root)
	if err != nil {
		t.Fatal(err)
	}

	// Empty workspace returns nil
	if env := ws.LoadEnv(); env != nil {
		t.Fatalf("expected nil env for empty workspace, got %v", env)
	}

	// Create .env
	envContent := `
GLOBAL_KEY=from_env
OVERRIDE_KEY=old_value
`
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte(envContent), 0o600); err != nil {
		t.Fatal(err)
	}

	env1 := ws.LoadEnv()
	expected1 := []string{
		"GLOBAL_KEY=from_env",
		"OVERRIDE_KEY=old_value",
	}
	if !reflect.DeepEqual(env1, expected1) {
		t.Fatalf("expected:\n%#v\ngot:\n%#v", expected1, env1)
	}

	// Create .env.local to test overrides
	envLocalContent := `
OVERRIDE_KEY=new_local_value
LOCAL_ONLY=secret
`
	if err := os.WriteFile(filepath.Join(root, ".env.local"), []byte(envLocalContent), 0o600); err != nil {
		t.Fatal(err)
	}

	env2 := ws.LoadEnv()
	expected2 := []string{
		"GLOBAL_KEY=from_env",
		"OVERRIDE_KEY=new_local_value",
		"LOCAL_ONLY=secret",
	}
	if !reflect.DeepEqual(env2, expected2) {
		t.Fatalf("expected:\n%#v\ngot:\n%#v", expected2, env2)
	}
}
