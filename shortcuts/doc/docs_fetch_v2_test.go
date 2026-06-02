// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package doc

import (
	"context"
	"testing"

	"github.com/larksuite/cli/internal/core"
	"github.com/larksuite/cli/internal/i18n"
	"github.com/larksuite/cli/shortcuts/common"
	"github.com/spf13/cobra"
)

func TestBuildFetchBodyIncludesSceneFromContext(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(), docsSceneContextKey, " DoubaoCLI ")
	runtime := newFetchBodyTestRuntime(ctx)

	body := buildFetchBody(runtime)
	if got := body["scene"]; got != "DoubaoCLI" {
		t.Fatalf("scene = %#v, want %q", got, "DoubaoCLI")
	}
}

func TestBuildCreateBodyIncludesSceneFromContext(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(), docsSceneContextKey, "DoubaoCLI")
	runtime := newCreateBodyTestRuntime(ctx)

	body := buildCreateBody(runtime)
	if got := body["scene"]; got != "DoubaoCLI" {
		t.Fatalf("scene = %#v, want %q", got, "DoubaoCLI")
	}
}

func TestBuildUpdateBodyIncludesSceneFromContext(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(), docsSceneContextKey, "DoubaoCLI")
	runtime := newUpdateBodyTestRuntime(ctx)

	body := buildUpdateBody(runtime)
	if got := body["scene"]; got != "DoubaoCLI" {
		t.Fatalf("scene = %#v, want %q", got, "DoubaoCLI")
	}
}

func TestBuildFetchBodyOmitsEmptyScene(t *testing.T) {
	t.Parallel()

	runtime := newFetchBodyTestRuntime(context.Background())

	body := buildFetchBody(runtime)
	if _, ok := body["scene"]; ok {
		t.Fatalf("did not expect empty scene in fetch body: %#v", body)
	}
}

func TestBuildFetchBodyLangExplicitCanonicalizes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want string
	}{
		{name: "upper short english", raw: "EN", want: "en_us"},
		{name: "hyphenated chinese", raw: "zh-CN", want: "zh_cn"},
		{name: "underscore japanese", raw: "ja_jp", want: "ja_jp"},
		{name: "common japanese country shorthand", raw: "JP", want: "ja_jp"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runtime := newFetchBodyTestRuntime(context.Background())
			if err := runtime.Cmd.Flags().Set("lang", tt.raw); err != nil {
				t.Fatalf("set lang: %v", err)
			}

			body := buildFetchBody(runtime)
			if got := body["lang"]; got != tt.want {
				t.Fatalf("lang = %#v, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildFetchBodyLangDefaultsToRuntimePreference(t *testing.T) {
	t.Parallel()

	runtime := newFetchBodyTestRuntimeWithConfig(context.Background(), &core.CliConfig{Lang: i18n.LangJaJP})

	body := buildFetchBody(runtime)
	if got := body["lang"]; got != "ja_jp" {
		t.Fatalf("lang = %#v, want ja_jp", got)
	}
	query := buildFetchQueryParams(runtime)
	if got := query.Get("lang"); got != "ja_jp" {
		t.Fatalf("query lang = %q, want ja_jp", got)
	}
}

func TestResolveFetchLangInvalidWarnsWithoutBodyLang(t *testing.T) {
	t.Parallel()

	runtime := newFetchBodyTestRuntime(context.Background())
	if err := runtime.Cmd.Flags().Set("lang", "xx-YY"); err != nil {
		t.Fatalf("set lang: %v", err)
	}

	body := buildFetchBody(runtime)
	if _, ok := body["lang"]; ok {
		t.Fatalf("invalid lang should not be sent in request body: %#v", body)
	}
	lang, warning := resolveFetchLang(runtime)
	if lang != "" {
		t.Fatalf("invalid lang resolved to %q", lang)
	}
	if warning == "" {
		t.Fatal("expected fallback warning")
	}
}

func TestAppendFetchWarningsMergesExistingWarnings(t *testing.T) {
	t.Parallel()

	data := map[string]interface{}{"warnings": []interface{}{"server_warning"}}
	appendFetchWarnings(data, "client_warning")

	got, ok := data["warnings"].([]interface{})
	if !ok {
		t.Fatalf("warnings type = %T, want []interface{}", data["warnings"])
	}
	if len(got) != 2 || got[0] != "server_warning" || got[1] != "client_warning" {
		t.Fatalf("warnings = %#v", got)
	}
}

func newFetchBodyTestRuntime(ctx context.Context) *common.RuntimeContext {
	return newFetchBodyTestRuntimeWithConfig(ctx, nil)
}

func newFetchBodyTestRuntimeWithConfig(ctx context.Context, cfg *core.CliConfig) *common.RuntimeContext {
	cmd := &cobra.Command{Use: "+fetch"}
	cmd.Flags().String("doc-format", "xml", "")
	cmd.Flags().String("detail", "simple", "")
	cmd.Flags().String("lang", "", "")
	cmd.Flags().Int("revision-id", -1, "")
	cmd.Flags().String("scope", "full", "")
	cmd.Flags().String("start-block-id", "", "")
	cmd.Flags().String("end-block-id", "", "")
	cmd.Flags().String("keyword", "", "")
	cmd.Flags().Int("context-before", 0, "")
	cmd.Flags().Int("context-after", 0, "")
	cmd.Flags().Int("max-depth", -1, "")
	return common.TestNewRuntimeContextWithCtx(ctx, cmd, cfg)
}

func newCreateBodyTestRuntime(ctx context.Context) *common.RuntimeContext {
	cmd := &cobra.Command{Use: "+create"}
	cmd.Flags().String("doc-format", "xml", "")
	cmd.Flags().String("content", "<title>hello</title>", "")
	cmd.Flags().String("parent-token", "", "")
	cmd.Flags().String("parent-position", "", "")
	return common.TestNewRuntimeContextWithCtx(ctx, cmd, nil)
}

func newUpdateBodyTestRuntime(ctx context.Context) *common.RuntimeContext {
	cmd := &cobra.Command{Use: "+update"}
	cmd.Flags().String("doc-format", "xml", "")
	cmd.Flags().String("command", "append", "")
	cmd.Flags().Int("revision-id", 0, "")
	cmd.Flags().String("content", "<p>hello</p>", "")
	cmd.Flags().String("pattern", "", "")
	cmd.Flags().String("block-id", "", "")
	cmd.Flags().String("src-block-ids", "", "")
	return common.TestNewRuntimeContextWithCtx(ctx, cmd, nil)
}
