package workspace

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectTimezoneOKFromLocaltime(t *testing.T) {
	tz, ok := DetectTimezoneOK()
	if !ok {
		// Environments without zoneinfo / localtime symlink may fail; skip.
		if _, err := os.Lstat("/etc/localtime"); err != nil {
			t.Skip("no /etc/localtime")
		}
		t.Fatalf("DetectTimezoneOK() failed on host with /etc/localtime")
	}
	if tz == "" || tz == "Local" {
		t.Fatalf("unexpected timezone %q", tz)
	}
	if !ValidIANATimezone(tz) {
		t.Fatalf("detected timezone %q is not loadable", tz)
	}
	t.Logf("detected timezone: %s", tz)
}

func TestListIANATimezones(t *testing.T) {
	zones := ListIANATimezones()
	if len(zones) < 2 {
		t.Fatalf("expected multiple zones, got %v", zones)
	}
	foundUTC := false
	for _, z := range zones {
		if z == "UTC" {
			foundUTC = true
		}
		if !ValidIANATimezone(z) {
			t.Fatalf("listed invalid zone %q", z)
		}
	}
	if !foundUTC {
		t.Fatal("UTC missing from list")
	}
}

func TestTimezoneFromLocaltimeSymlink(t *testing.T) {
	target, err := filepath.EvalSymlinks("/etc/localtime")
	if err != nil {
		t.Skip("no resolvable /etc/localtime symlink")
	}
	t.Logf("/etc/localtime -> %s", target)
	tz, ok := timezoneFromLocaltime()
	if !ok {
		t.Fatalf("timezoneFromLocaltime failed for %s", target)
	}
	t.Logf("resolved: %s", tz)
}
