// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package docs

import (
	"context"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	clie2e "github.com/larksuite/cli/tests/cli_e2e"
	"github.com/larksuite/cli/tests/cli_e2e/drive"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

// TestDocs_CoverWorkflow proves the upload, update, get, and delete cover path.
func TestDocs_CoverWorkflow(t *testing.T) {
	parentT := t
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	t.Cleanup(cancel)

	suffix := clie2e.GenerateSuffix()
	folderName := "lark-cli-e2e-cover-folder-" + suffix
	docTitle := "lark-cli-e2e-cover-" + suffix
	const defaultAs = "bot"

	folderToken := drive.CreateDriveFolder(t, parentT, ctx, folderName, defaultAs, "")
	docToken := createDocWithRetry(t, parentT, ctx, folderToken, docTitle, "# Cover\n\nCover workflow fixture.", defaultAs)

	imagePath := filepath.Join(t.TempDir(), "cover.png")
	writeDocCoverFixturePNG(t, imagePath)

	var coverToken string
	t.Run("upload cover image as bot", func(t *testing.T) {
		result, err := clie2e.RunCmd(ctx, clie2e.Request{
			Args: []string{
				"docs", "+media-upload",
				"--file", imagePath,
				"--parent-type", "docx_image",
				"--parent-node", docToken,
				"--doc-id", docToken,
			},
			DefaultAs: defaultAs,
		})
		require.NoError(t, err)
		result.AssertExitCode(t, 0)
		result.AssertStdoutStatus(t, true)
		coverToken = gjson.Get(result.Stdout, "data.file_token").String()
		require.NotEmpty(t, coverToken, "stdout:\n%s", result.Stdout)
	})

	t.Run("update cover as bot", func(t *testing.T) {
		require.NotEmpty(t, coverToken, "cover token should be uploaded before update")

		result, err := clie2e.RunCmd(ctx, clie2e.Request{
			Args: []string{
				"docs", "+cover-update",
				"--doc", docToken,
				"--token", coverToken,
				"--offset-ratio-x", "0.2",
				"--offset-ratio-y", "0.3",
			},
			DefaultAs: defaultAs,
		})
		require.NoError(t, err)
		result.AssertExitCode(t, 0)
		result.AssertStdoutStatus(t, true)
	})

	t.Run("get cover as bot", func(t *testing.T) {
		result, err := clie2e.RunCmd(ctx, clie2e.Request{
			Args: []string{
				"docs", "+cover-get",
				"--doc", docToken,
			},
			DefaultAs: defaultAs,
		})
		require.NoError(t, err)
		result.AssertExitCode(t, 0)
		result.AssertStdoutStatus(t, true)
		assert.Equal(t, coverToken, gjson.Get(result.Stdout, "data.cover.token").String(), "stdout:\n%s", result.Stdout)
	})

	t.Run("delete cover as bot", func(t *testing.T) {
		result, err := clie2e.RunCmd(ctx, clie2e.Request{
			Args: []string{
				"docs", "+cover-delete",
				"--doc", docToken,
			},
			DefaultAs: defaultAs,
		})
		require.NoError(t, err)
		result.AssertExitCode(t, 0)
		result.AssertStdoutStatus(t, true)
		assert.Equal(t, docToken, gjson.Get(result.Stdout, "data.document_id").String(), "stdout:\n%s", result.Stdout)
	})

	t.Run("verify cover removed as bot", func(t *testing.T) {
		result, err := clie2e.RunCmd(ctx, clie2e.Request{
			Args: []string{
				"docs", "+cover-get",
				"--doc", docToken,
			},
			DefaultAs: defaultAs,
		})
		require.NoError(t, err)
		result.AssertExitCode(t, 0)
		result.AssertStdoutStatus(t, true)
		assert.False(t, gjson.Get(result.Stdout, "data.cover.token").Exists(), "stdout:\n%s", result.Stdout)
	})
}

func writeDocCoverFixturePNG(t *testing.T, path string) {
	t.Helper()

	img := image.NewRGBA(image.Rect(0, 0, 1, 1))
	img.Set(0, 0, color.RGBA{R: 0x2f, G: 0x80, B: 0xed, A: 0xff})

	f, err := os.Create(path)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, f.Close())
	}()

	require.NoError(t, png.Encode(f, img))
}
