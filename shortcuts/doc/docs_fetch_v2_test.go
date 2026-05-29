// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package doc

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/larksuite/cli/internal/cmdutil"
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

func TestBuildFetchBodyUsesExplicitLang(t *testing.T) {
	t.Parallel()

	runtime := newFetchBodyTestRuntime(context.Background())
	if err := runtime.Cmd.Flags().Set("lang", "EN-US"); err != nil {
		t.Fatalf("set lang: %v", err)
	}

	body := buildFetchBody(runtime)
	if got := body["lang"]; got != "en_us" {
		t.Fatalf("lang = %#v, want %q", got, "en_us")
	}
}

func TestBuildFetchBodyUsesRuntimeLangWhenFlagOmitted(t *testing.T) {
	t.Parallel()

	runtime := newFetchBodyTestRuntime(context.Background())
	runtime.Config = &core.CliConfig{Lang: i18n.LangJaJP}

	body := buildFetchBody(runtime)
	if got := body["lang"]; got != "ja_jp" {
		t.Fatalf("lang = %#v, want %q", got, "ja_jp")
	}
}

func TestBuildFetchBodyUnsupportedExplicitLangFallsBack(t *testing.T) {
	t.Parallel()

	runtime := newFetchBodyTestRuntime(context.Background())
	runtime.Config = &core.CliConfig{Lang: i18n.LangEnUS}
	if err := runtime.Cmd.Flags().Set("lang", "unknown"); err != nil {
		t.Fatalf("set lang: %v", err)
	}

	body := buildFetchBody(runtime)
	if _, ok := body["lang"]; ok {
		t.Fatalf("unsupported explicit lang should be omitted: %#v", body)
	}
}

func TestWarnInvalidFetchLang(t *testing.T) {
	t.Parallel()

	var stderr bytes.Buffer
	runtime := newFetchBodyTestRuntime(context.Background())
	runtime.Factory = &cmdutil.Factory{IOStreams: &cmdutil.IOStreams{ErrOut: &stderr}}
	if err := runtime.Cmd.Flags().Set("lang", "unknown"); err != nil {
		t.Fatalf("set lang: %v", err)
	}

	warnInvalidFetchLang(runtime)
	if got := stderr.String(); !strings.Contains(got, `unsupported --lang "unknown"`) {
		t.Fatalf("stderr = %q, want unsupported lang warning", got)
	}
}

func TestUseV2FetchWhenLangChanged(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "+fetch"}
	cmd.Flags().String("api-version", "v1", "")
	cmd.Flags().String("lang", "", "")
	if err := cmd.Flags().Set("lang", "en"); err != nil {
		t.Fatalf("set lang: %v", err)
	}

	runtime := common.TestNewRuntimeContext(cmd, nil)
	if !useV2Fetch(runtime) {
		t.Fatal("explicit --lang should route docs +fetch to v2")
	}
}

func newFetchBodyTestRuntime(ctx context.Context) *common.RuntimeContext {
	cmd := &cobra.Command{Use: "+fetch"}
	cmd.Flags().String("doc-format", "xml", "")
	cmd.Flags().String("detail", "simple", "")
	cmd.Flags().Int("revision-id", -1, "")
	cmd.Flags().String("lang", "", "")
	cmd.Flags().String("scope", "full", "")
	cmd.Flags().String("start-block-id", "", "")
	cmd.Flags().String("end-block-id", "", "")
	cmd.Flags().String("keyword", "", "")
	cmd.Flags().Int("context-before", 0, "")
	cmd.Flags().Int("context-after", 0, "")
	cmd.Flags().Int("max-depth", -1, "")
	return common.TestNewRuntimeContextWithCtx(ctx, cmd, nil)
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
