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

func TestBuildFetchBodyIncludesConfiguredLang(t *testing.T) {
	t.Parallel()

	runtime := newFetchBodyTestRuntime(context.Background())
	runtime.Config.Lang = i18n.LangEnUS

	body := buildFetchBody(runtime)
	if got := body["lang"]; got != "en_us" {
		t.Fatalf("lang = %#v, want %q", got, "en_us")
	}
}

func TestBuildFetchBodyExplicitLangOverridesConfiguredLang(t *testing.T) {
	t.Parallel()

	runtime := newFetchBodyTestRuntime(context.Background())
	runtime.Config.Lang = i18n.LangZhCN
	if err := runtime.Cmd.Flags().Set("lang", "en_us"); err != nil {
		t.Fatalf("set lang flag: %v", err)
	}

	body := buildFetchBody(runtime)
	if got := body["lang"]; got != "en_us" {
		t.Fatalf("lang = %#v, want %q", got, "en_us")
	}
}

func TestBuildFetchBodyPassesExplicitUnsupportedLang(t *testing.T) {
	t.Parallel()

	runtime := newFetchBodyTestRuntime(context.Background())
	if err := runtime.Cmd.Flags().Set("lang", "unsupported_lang"); err != nil {
		t.Fatalf("set lang flag: %v", err)
	}

	body := buildFetchBody(runtime)
	if got := body["lang"]; got != "unsupported_lang" {
		t.Fatalf("lang = %#v, want %q", got, "unsupported_lang")
	}
}

func TestBuildFetchBodyOmitsUnrecognizedLang(t *testing.T) {
	t.Parallel()

	runtime := newFetchBodyTestRuntime(context.Background())
	runtime.Config.Lang = "bad_lang"

	body := buildFetchBody(runtime)
	if _, ok := body["lang"]; ok {
		t.Fatalf("did not expect invalid lang in fetch body: %#v", body)
	}
}

func newFetchBodyTestRuntime(ctx context.Context) *common.RuntimeContext {
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
	return common.TestNewRuntimeContextWithCtx(ctx, cmd, &core.CliConfig{})
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
