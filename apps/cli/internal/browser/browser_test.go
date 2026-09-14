package browser
 
import (
	"testing"
)

func TestIsLocalhost(t *testing.T) {
	cases := []struct {
		host  string
		valid bool
	}{
		{"localhost", true},
		{"127.0.0.1", true},
		{"::1", true},
		{"web.beep.localhost", true},
		{"core.localhost", true},
		{"foo.bar.localhost", true},
		{"example.com", false},
		{"notlocalhost", false},
		{"localhost.com", false},
		{"evil.com", false},
	}

	for _, tc := range cases {
		if got := isLocalhost(tc.host); got != tc.valid {
			t.Errorf("isLocalhost(%q) = %v, want %v", tc.host, got, tc.valid)
		}
	}
}
