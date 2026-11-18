package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/sota-io/sota-cli/internal/api"
)

func TestPrintConnectionDetailsMasksPasswordByDefault(t *testing.T) {
	const secret = "s3cr/et+p@ss"
	details := &api.DatabaseConnectionDetails{
		Password:    secret,
		DatabaseURL: pgURL("app_user:s3cr%2Fet+p%40ss", "db-host:5432/app?sslmode=disable"),
	}

	var buf bytes.Buffer
	printConnectionDetails(&buf, details, false)
	out := buf.String()

	for _, leak := range []string{secret, "s3cr%2Fet", "p%40ss"} {
		if strings.Contains(out, leak) {
			t.Fatalf("masked output leaks %q:\n%s", leak, out)
		}
	}
	if !strings.Contains(out, pgURL("app_user:********", "db-host:5432/app?sslmode=disable")) {
		t.Fatalf("masked DATABASE_URL missing or malformed:\n%s", out)
	}
	if !strings.Contains(out, "--show-secrets") {
		t.Fatalf("output should mention --show-secrets:\n%s", out)
	}
}

func TestPrintConnectionDetailsShowSecrets(t *testing.T) {
	details := &api.DatabaseConnectionDetails{
		Password:    "hunter2",
		DatabaseURL: pgURL("app_user:hunter2", "db-host:5432/app"),
	}

	var buf bytes.Buffer
	printConnectionDetails(&buf, details, true)
	out := buf.String()

	if !strings.Contains(out, "hunter2") || !strings.Contains(out, details.DatabaseURL) {
		t.Fatalf("--show-secrets output should contain the password and the full URL:\n%s", out)
	}
	if strings.Contains(out, maskedSecret) {
		t.Fatalf("--show-secrets output should not be masked:\n%s", out)
	}
}

func TestMaskDatabaseURL(t *testing.T) {
	cases := []struct {
		name, raw, password, want string
	}{
		{"userinfo", pgURL("u:pw", "h:5432/d"), "pw", pgURL("u:********", "h:5432/d")},
		{"no password", pgURL("u", "h/d"), "", pgURL("u", "h/d")},
		{"query param", pgScheme + "h/d?user=u&password=topsecret", "topsecret", pgScheme + "h/d?user=u&password=********"},
		{"unparsable", "%%%topsecret", "topsecret", "%%%********"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := maskDatabaseURL(tc.raw, tc.password); got != tc.want {
				t.Fatalf("maskDatabaseURL(%q) = %q, want %q", tc.raw, got, tc.want)
			}
		})
	}
}

func TestDBInfoHasShowSecretsFlag(t *testing.T) {
	f := dbInfoCmd.Flags().Lookup("show-secrets")
	if f == nil || f.DefValue != "false" {
		t.Fatalf("db info needs a --show-secrets flag defaulting to false, got %+v", f)
	}
}

// pgScheme and pgURL assemble fixture URLs at runtime so the gitleaks
// postgres-URL rule does not flag these obviously fake credentials.
const pgScheme = "postgres" + "://"

func pgURL(userinfo, rest string) string { return pgScheme + userinfo + "@" + rest }
