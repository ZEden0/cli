// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package doc

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/larksuite/cli/internal/cmdutil"
	"github.com/larksuite/cli/internal/httpmock"
	"github.com/larksuite/cli/shortcuts/common"
)

func TestDocCoverUpdateDryRunOmitsOffsetsWhenNotPassed(t *testing.T) {
	cmd := newDocCoverUpdateTestCmd(t, "doxcnCoverDryRun", "file_cover_123")

	dry := decodeDocDryRun(t, DocCoverUpdate.DryRun(context.Background(), common.TestNewRuntimeContext(cmd, nil)))
	if len(dry.API) != 1 {
		t.Fatalf("expected 1 API call, got %d", len(dry.API))
	}
	if dry.API[0].Method != "PATCH" {
		t.Fatalf("method = %q, want PATCH", dry.API[0].Method)
	}
	if dry.API[0].URL != "/open-apis/docx/v1/documents/doxcnCoverDryRun" {
		t.Fatalf("url = %q", dry.API[0].URL)
	}
	cover := dryCoverBody(t, dry.API[0].Body)
	if cover["token"] != "file_cover_123" {
		t.Fatalf("token = %v", cover["token"])
	}
	if _, ok := cover["offset_ratio_x"]; ok {
		t.Fatalf("offset_ratio_x should be omitted when flag is not passed: %#v", cover)
	}
	if _, ok := cover["offset_ratio_y"]; ok {
		t.Fatalf("offset_ratio_y should be omitted when flag is not passed: %#v", cover)
	}
}

func TestDocCoverUpdateDryRunIncludesExplicitOffsets(t *testing.T) {
	cmd := newDocCoverUpdateTestCmd(t, "doxcnCoverDryRun", "file_cover_123")
	mustSetDocFlag(t, cmd, "offset-ratio-x", "0")
	mustSetDocFlag(t, cmd, "offset-ratio-y", "1")

	dry := decodeDocDryRun(t, DocCoverUpdate.DryRun(context.Background(), common.TestNewRuntimeContext(cmd, nil)))
	cover := dryCoverBody(t, dry.API[0].Body)
	if cover["offset_ratio_x"] != float64(0) {
		t.Fatalf("offset_ratio_x = %#v, want 0", cover["offset_ratio_x"])
	}
	if cover["offset_ratio_y"] != float64(1) {
		t.Fatalf("offset_ratio_y = %#v, want 1", cover["offset_ratio_y"])
	}
}

func TestDocCoverUpdateDryRunWikiAddsResolveStep(t *testing.T) {
	cmd := newDocCoverUpdateTestCmd(t, "https://example.larksuite.com/wiki/wikcnCover", "file_cover_123")

	dry := decodeDocDryRun(t, DocCoverUpdate.DryRun(context.Background(), common.TestNewRuntimeContext(cmd, nil)))
	if len(dry.API) != 2 {
		t.Fatalf("expected wiki resolve + patch, got %d calls", len(dry.API))
	}
	if dry.API[0].URL != "/open-apis/wiki/v2/spaces/get_node" {
		t.Fatalf("first URL = %q", dry.API[0].URL)
	}
	if dry.API[1].URL != "/open-apis/docx/v1/documents/resolved_docx_token" {
		t.Fatalf("second URL = %q", dry.API[1].URL)
	}
}

func TestDocCoverUpdateValidateRejectsBadOffsets(t *testing.T) {
	tests := []string{"NaN", "+Inf", "-0.01", "1.01", "abc", ""}
	for _, raw := range tests {
		t.Run(raw, func(t *testing.T) {
			cmd := newDocCoverUpdateTestCmd(t, "doxcnCoverDryRun", "file_cover_123")
			mustSetDocFlag(t, cmd, "offset-ratio-x", raw)
			err := DocCoverUpdate.Validate(context.Background(), common.TestNewRuntimeContext(cmd, nil))
			if err == nil {
				t.Fatal("expected validation error, got nil")
			}
			if !strings.Contains(err.Error(), "[0,1]") {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestDocCoverRejectsOldDocURL(t *testing.T) {
	f, _, _, _ := cmdutil.TestFactory(t, docsTestConfigWithAppID("docs-cover-old-doc-app"))
	err := mountAndRunDocs(t, DocCoverGet, []string{
		"+cover-get",
		"--doc", "https://example.larksuite.com/doc/oldDocToken",
		"--as", "bot",
	}, f, nil)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
	if !strings.Contains(err.Error(), "only support Docx") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDocCoverUpdateExecuteOmitsOffsetsWhenNotPassed(t *testing.T) {
	f, stdout, _, reg := cmdutil.TestFactory(t, docsTestConfigWithAppID("docs-cover-update-app"))
	updateStub := &httpmock.Stub{
		Method: "PATCH",
		URL:    "/open-apis/docx/v1/documents/doxcnCoverExec",
		Body: map[string]interface{}{
			"code": 0,
			"msg":  "ok",
			"data": map[string]interface{}{
				"document": map[string]interface{}{
					"cover": map[string]interface{}{"token": "file_cover_123"},
				},
			},
		},
	}
	reg.Register(updateStub)

	err := mountAndRunDocs(t, DocCoverUpdate, []string{
		"+cover-update",
		"--doc", "doxcnCoverExec",
		"--token", "file_cover_123",
		"--as", "bot",
	}, f, stdout)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cover := capturedCoverBody(t, updateStub)
	if _, ok := cover["offset_ratio_x"]; ok {
		t.Fatalf("offset_ratio_x should be omitted: %#v", cover)
	}
	if !strings.Contains(stdout.String(), "file_cover_123") {
		t.Fatalf("stdout missing cover token: %s", stdout.String())
	}
}

func TestDocCoverUpdateExecuteIncludesExplicitOffset(t *testing.T) {
	f, _, _, reg := cmdutil.TestFactory(t, docsTestConfigWithAppID("docs-cover-update-offset-app"))
	updateStub := &httpmock.Stub{
		Method: "PATCH",
		URL:    "/open-apis/docx/v1/documents/doxcnCoverExec",
		Body:   map[string]interface{}{"code": 0, "msg": "ok", "data": map[string]interface{}{}},
	}
	reg.Register(updateStub)

	err := mountAndRunDocs(t, DocCoverUpdate, []string{
		"+cover-update",
		"--doc", "doxcnCoverExec",
		"--token", "file_cover_123",
		"--offset-ratio-x", "0.2",
		"--as", "bot",
	}, f, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cover := capturedCoverBody(t, updateStub)
	if cover["offset_ratio_x"] != 0.2 {
		t.Fatalf("offset_ratio_x = %#v, want 0.2", cover["offset_ratio_x"])
	}
	if _, ok := cover["offset_ratio_y"]; ok {
		t.Fatalf("offset_ratio_y should be omitted: %#v", cover)
	}
}

func TestDocCoverGetExecuteOutputsCover(t *testing.T) {
	f, stdout, _, reg := cmdutil.TestFactory(t, docsTestConfigWithAppID("docs-cover-get-app"))
	reg.Register(&httpmock.Stub{
		Method: "GET",
		URL:    "/open-apis/docx/v1/documents/doxcnCoverGet",
		Body: map[string]interface{}{
			"code": 0,
			"msg":  "ok",
			"data": map[string]interface{}{
				"document": map[string]interface{}{
					"cover": map[string]interface{}{
						"token":          "file_cover_123",
						"offset_ratio_x": 0.2,
						"offset_ratio_y": 0.3,
					},
				},
			},
		},
	})

	err := mountAndRunDocs(t, DocCoverGet, []string{
		"+cover-get",
		"--doc", "doxcnCoverGet",
		"--as", "bot",
	}, f, stdout)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(stdout.String(), "file_cover_123") {
		t.Fatalf("stdout missing cover token: %s", stdout.String())
	}
}

func TestDocCoverDeleteDryRunBodyUsesNullCover(t *testing.T) {
	cmd := &cobra.Command{Use: "docs +cover-delete"}
	cmd.Flags().String("doc", "", "")
	mustSetDocFlag(t, cmd, "doc", "doxcnCoverDelete")

	dry := decodeDocDryRun(t, DocCoverDelete.DryRun(context.Background(), common.TestNewRuntimeContext(cmd, nil)))
	body := dry.API[0].Body
	updateCover, _ := body["update_cover"].(map[string]interface{})
	if updateCover == nil {
		t.Fatalf("missing update_cover: %#v", body)
	}
	if v, ok := updateCover["cover"]; !ok || v != nil {
		t.Fatalf("cover = %#v, want nil", updateCover["cover"])
	}
}

func newDocCoverUpdateTestCmd(t *testing.T, docID, token string) *cobra.Command {
	t.Helper()
	cmd := &cobra.Command{Use: "docs +cover-update"}
	cmd.Flags().String("doc", "", "")
	cmd.Flags().String("token", "", "")
	cmd.Flags().String("offset-ratio-x", "", "")
	cmd.Flags().String("offset-ratio-y", "", "")
	mustSetDocFlag(t, cmd, "doc", docID)
	mustSetDocFlag(t, cmd, "token", token)
	return cmd
}

func mustSetDocFlag(t *testing.T, cmd *cobra.Command, name, value string) {
	t.Helper()
	if err := cmd.Flags().Set(name, value); err != nil {
		t.Fatalf("set --%s: %v", name, err)
	}
}

func dryCoverBody(t *testing.T, body map[string]interface{}) map[string]interface{} {
	t.Helper()
	updateCover, _ := body["update_cover"].(map[string]interface{})
	if updateCover == nil {
		t.Fatalf("missing update_cover: %#v", body)
	}
	cover, _ := updateCover["cover"].(map[string]interface{})
	if cover == nil {
		t.Fatalf("missing cover: %#v", updateCover)
	}
	return cover
}

func capturedCoverBody(t *testing.T, stub *httpmock.Stub) map[string]interface{} {
	t.Helper()
	var raw map[string]interface{}
	if err := json.NewDecoder(bytes.NewReader(stub.CapturedBody)).Decode(&raw); err != nil {
		t.Fatalf("decode captured body: %v; body=%s", err, string(stub.CapturedBody))
	}
	updateCover, _ := raw["update_cover"].(map[string]interface{})
	if updateCover == nil {
		t.Fatalf("missing update_cover: %#v", raw)
	}
	cover, _ := updateCover["cover"].(map[string]interface{})
	if cover == nil {
		t.Fatalf("missing cover: %#v", updateCover)
	}
	return cover
}
