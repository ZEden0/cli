// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package docs

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	clie2e "github.com/larksuite/cli/tests/cli_e2e"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// TestDocs_UpdateDryRunSuppressesSemanticWarnings asserts the contract that
// docsUpdateWarnings is NOT invoked on the --dry-run path. The unit tests in
// shortcuts/doc/docs_update_check_test.go prove the helper emits warnings for
// replace_range + blank-line and for combined-emphasis markers; this E2E
// locks in that they never reach the user during dry-run planning, so a
// future refactor that moves warning emission into a shared code path can't
// silently regress.
//
// Input is intentionally crafted to trigger BOTH warnings the helper emits:
//   - mode=replace_range + markdown containing "\n\n" (blank-line warning)
//   - markdown containing `***combined***` (combined bold+italic warning)
//
// Neither string may appear in dry-run output.
func TestDocs_UpdateDryRunSuppressesSemanticWarnings(t *testing.T) {
	// Fake creds are enough — dry-run short-circuits before any real API call.
	t.Setenv("LARKSUITE_CLI_APP_ID", "app")
	t.Setenv("LARKSUITE_CLI_APP_SECRET", "secret")
	t.Setenv("LARKSUITE_CLI_BRAND", "feishu")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)

	// "***combined***" is a triple-asterisk combined-emphasis shape; "\n\n"
	// is a paragraph break. Both would normally produce warnings when
	// Execute runs under --mode=replace_range; both must be absent here.
	markdown := "***combined***\n\nsecond paragraph"

	result, err := clie2e.RunCmd(ctx, clie2e.Request{
		Args: []string{
			"docs", "+update",
			"--doc", "doxcnDryRunE2E",
			"--mode", "replace_range",
			"--selection-with-ellipsis", "placeholder",
			"--markdown", markdown,
			"--dry-run",
		},
		DefaultAs: "bot",
	})
	require.NoError(t, err)
	result.AssertExitCode(t, 0)

	// Neither warning prefix ("warning:") nor either specific warning body
	// may appear in dry-run output (stdout OR stderr).
	combined := result.Stdout + "\n" + result.Stderr
	for _, needle := range []string{
		"warning:",
		"does not split a block into multiple paragraphs",
		"combined bold+italic markers",
	} {
		if strings.Contains(combined, needle) {
			t.Errorf("dry-run output must not surface pre-write warning %q\nstdout:\n%s\nstderr:\n%s",
				needle, result.Stdout, result.Stderr)
		}
	}
}

func TestDocs_FetchDryRunIncludesConfiguredLang(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", configDir)
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "config.json"), []byte(`{
  "apps": [
    {
      "appId": "app",
      "appSecret": "secret",
      "brand": "feishu",
      "lang": "en_us"
    }
  ]
}
`), 0o600))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)

	result, err := clie2e.RunCmd(ctx, clie2e.Request{
		Args: []string{
			"docs", "+fetch",
			"--api-version", "v2",
			"--doc", "doxcnDryRunLangE2E",
			"--dry-run",
		},
		DefaultAs: "bot",
		Env: map[string]string{
			"LARKSUITE_CLI_APP_ID":              "",
			"LARKSUITE_CLI_APP_SECRET":          "",
			"LARKSUITE_CLI_USER_ACCESS_TOKEN":   "",
			"LARKSUITE_CLI_TENANT_ACCESS_TOKEN": "",
			"LARKSUITE_CLI_DEFAULT_AS":          "",
			"LARKSUITE_CLI_STRICT_MODE":         "",
			"OPENCLAW_CLI":                      "",
			"OPENCLAW_HOME":                     "",
			"OPENCLAW_STATE_DIR":                "",
			"OPENCLAW_CONFIG_PATH":              "",
			"OPENCLAW_SERVICE_MARKER":           "",
			"OPENCLAW_SERVICE_VERSION":          "",
			"OPENCLAW_GATEWAY_PORT":             "",
			"OPENCLAW_SHELL":                    "",
			"HERMES_HOME":                       "",
			"HERMES_QUIET":                      "",
			"HERMES_EXEC_ASK":                   "",
			"HERMES_GATEWAY_TOKEN":              "",
			"HERMES_SESSION_KEY":                "",
			"LARK_CHANNEL":                      "",
		},
	})
	require.NoError(t, err)
	result.AssertExitCode(t, 0)

	if got := gjson.Get(result.Stdout, "api.0.method").String(); got != "POST" {
		t.Fatalf("method = %q, want POST\nstdout:\n%s", got, result.Stdout)
	}
	if got := gjson.Get(result.Stdout, "api.0.url").String(); got != "/open-apis/docs_ai/v1/documents/doxcnDryRunLangE2E/fetch" {
		t.Fatalf("url = %q, want docs_ai fetch URL\nstdout:\n%s", got, result.Stdout)
	}
	if got := gjson.Get(result.Stdout, "api.0.body.lang").String(); got != "en_us" {
		t.Fatalf("body.lang = %q, want en_us\nstdout:\n%s", got, result.Stdout)
	}
}
