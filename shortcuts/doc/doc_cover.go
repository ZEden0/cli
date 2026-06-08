// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package doc

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/larksuite/cli/internal/output"
	"github.com/larksuite/cli/shortcuts/common"
)

const coverRelationHint = "Cover images must use a file_token uploaded for the target Docx with `docs +media-upload --parent-type docx_image --parent-node <document_id> --doc-id <document_id>`. Do not reuse tokens from docs +media-insert, IM uploads, or ordinary Drive files."

var DocCoverGet = common.Shortcut{
	Service:     "docs",
	Command:     "+cover-get",
	Description: "Get a Docx document cover",
	Risk:        "read",
	Scopes:      []string{"docx:document:readonly"},
	AuthTypes:   []string{"user", "bot"},
	Tips: []string{
		coverRelationHint,
	},
	Flags: []common.Flag{
		{Name: "doc", Desc: "Docx document ID, Docx URL, or wiki URL resolving to Docx", Required: true},
	},
	Validate: validateDocCoverDoc,
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		return dryRunDocCover(ctx, runtime, "get")
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		documentID, err := resolveDocxDocumentIDForCover(runtime, runtime.Str("doc"))
		if err != nil {
			return err
		}
		apiPath := docCoverDocumentAPIPath(documentID)
		data, err := doDocAPI(runtime, "GET", apiPath, nil)
		if err != nil {
			return docCoverRelationError(err)
		}
		runtime.Out(map[string]interface{}{
			"document_id": documentID,
			"cover":       common.GetMap(data, "document")["cover"],
		}, nil)
		return nil
	},
}

var DocCoverUpdate = common.Shortcut{
	Service:     "docs",
	Command:     "+cover-update",
	Description: "Update a Docx document cover with an uploaded docx_image token",
	Risk:        "write",
	Scopes:      []string{"docx:document:write_only", "docx:document:readonly"},
	AuthTypes:   []string{"user", "bot"},
	Tips: []string{
		coverRelationHint,
	},
	Flags: []common.Flag{
		{Name: "doc", Desc: "Docx document ID, Docx URL, or wiki URL resolving to Docx", Required: true},
		{Name: "token", Desc: "file_token from docs +media-upload with parent-type docx_image", Required: true},
		{Name: "offset-ratio-x", Desc: "optional cover crop horizontal ratio in [0,1]"},
		{Name: "offset-ratio-y", Desc: "optional cover crop vertical ratio in [0,1]"},
	},
	Validate: validateDocCoverUpdate,
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		return dryRunDocCover(ctx, runtime, "update")
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		documentID, err := resolveDocxDocumentIDForCover(runtime, runtime.Str("doc"))
		if err != nil {
			return err
		}
		body, err := buildDocCoverUpdateBody(runtime)
		if err != nil {
			return err
		}
		apiPath := docCoverDocumentAPIPath(documentID)
		data, err := doDocAPI(runtime, "PATCH", apiPath, body)
		if err != nil {
			return docCoverRelationError(err)
		}
		runtime.Out(map[string]interface{}{
			"document_id": documentID,
			"cover":       common.GetMap(data, "document")["cover"],
		}, nil)
		return nil
	},
}

var DocCoverDelete = common.Shortcut{
	Service:     "docs",
	Command:     "+cover-delete",
	Description: "Delete a Docx document cover",
	Risk:        "write",
	Scopes:      []string{"docx:document:write_only", "docx:document:readonly"},
	AuthTypes:   []string{"user", "bot"},
	Tips: []string{
		"Deleting a cover is a document write operation; the selected user or bot must be able to edit the target Docx.",
	},
	Flags: []common.Flag{
		{Name: "doc", Desc: "Docx document ID, Docx URL, or wiki URL resolving to Docx", Required: true},
	},
	Validate: validateDocCoverDoc,
	DryRun: func(ctx context.Context, runtime *common.RuntimeContext) *common.DryRunAPI {
		return dryRunDocCover(ctx, runtime, "delete")
	},
	Execute: func(ctx context.Context, runtime *common.RuntimeContext) error {
		documentID, err := resolveDocxDocumentIDForCover(runtime, runtime.Str("doc"))
		if err != nil {
			return err
		}
		body := buildDocCoverDeleteBody()
		apiPath := docCoverDocumentAPIPath(documentID)
		data, err := doDocAPI(runtime, "PATCH", apiPath, body)
		if err != nil {
			return docCoverRelationError(err)
		}
		out := map[string]interface{}{
			"document_id": documentID,
			"deleted":     true,
		}
		if doc := common.GetMap(data, "document"); len(doc) > 0 {
			out["document"] = doc
		}
		runtime.Out(out, nil)
		return nil
	},
}

func validateDocCoverDoc(_ context.Context, runtime *common.RuntimeContext) error {
	ref, err := parseDocumentRef(runtime.Str("doc"))
	if err != nil {
		return common.FlagErrorf("invalid --doc: %v", err)
	}
	if ref.Kind == "doc" {
		return common.FlagErrorf("docs cover commands only support Docx documents; use a docx token/URL or a wiki URL that resolves to docx")
	}
	return nil
}

func validateDocCoverUpdate(ctx context.Context, runtime *common.RuntimeContext) error {
	if err := validateDocCoverDoc(ctx, runtime); err != nil {
		return err
	}
	if strings.TrimSpace(runtime.Str("token")) == "" {
		return common.FlagErrorf("--token is required")
	}
	if _, err := buildDocCoverUpdateBody(runtime); err != nil {
		return err
	}
	return nil
}

func dryRunDocCover(_ context.Context, runtime *common.RuntimeContext, op string) *common.DryRunAPI {
	docRef, err := parseDocumentRef(runtime.Str("doc"))
	if err != nil {
		return common.NewDryRunAPI().Set("error", err.Error())
	}
	if docRef.Kind == "doc" {
		return common.NewDryRunAPI().Set("error", "docs cover commands only support Docx documents; use a docx token/URL or a wiki URL that resolves to docx")
	}
	documentID := docRef.Token
	d := common.NewDryRunAPI()
	if docRef.Kind == "wiki" {
		documentID = "resolved_docx_token"
		d.GET("/open-apis/wiki/v2/spaces/get_node").
			Desc("[1] Resolve wiki node to docx document").
			Params(map[string]interface{}{"token": docRef.Token})
	}
	apiPath := docCoverDocumentAPIPath(":document_id")
	switch op {
	case "get":
		d.GET(apiPath).Desc(docCoverStepDesc(docRef.Kind, "Get document cover"))
	case "update":
		body, err := buildDocCoverUpdateBody(runtime)
		if err != nil {
			return d.Set("error", err.Error())
		}
		d.PATCH(apiPath).
			Desc(docCoverStepDesc(docRef.Kind, "Update document cover")).
			Body(body)
	case "delete":
		d.PATCH(apiPath).
			Desc(docCoverStepDesc(docRef.Kind, "Delete document cover")).
			Body(buildDocCoverDeleteBody())
	}
	return d.Set("document_id", documentID)
}

func docCoverStepDesc(kind, desc string) string {
	if kind == "wiki" {
		return "[2] " + desc
	}
	return desc
}

func resolveDocxDocumentIDForCover(runtime *common.RuntimeContext, input string) (string, error) {
	docRef, err := parseDocumentRef(input)
	if err != nil {
		return "", err
	}
	if docRef.Kind == "doc" {
		return "", output.ErrValidation("docs cover commands only support Docx documents; use a docx token/URL or a wiki URL that resolves to docx")
	}
	return resolveDocxDocumentID(runtime, input)
}

func docCoverDocumentAPIPath(documentID string) string {
	return fmt.Sprintf("/open-apis/docx/v1/documents/%s", documentID)
}

func buildDocCoverUpdateBody(runtime *common.RuntimeContext) (map[string]interface{}, error) {
	cover := map[string]interface{}{
		"token": strings.TrimSpace(runtime.Str("token")),
	}
	if runtime.Changed("offset-ratio-x") {
		v, err := parseDocCoverOffset(runtime.Str("offset-ratio-x"), "--offset-ratio-x")
		if err != nil {
			return nil, err
		}
		cover["offset_ratio_x"] = v
	}
	if runtime.Changed("offset-ratio-y") {
		v, err := parseDocCoverOffset(runtime.Str("offset-ratio-y"), "--offset-ratio-y")
		if err != nil {
			return nil, err
		}
		cover["offset_ratio_y"] = v
	}
	return map[string]interface{}{
		"update_cover": map[string]interface{}{
			"cover": cover,
		},
	}, nil
}

func buildDocCoverDeleteBody() map[string]interface{} {
	return map[string]interface{}{
		"update_cover": map[string]interface{}{
			"cover": nil,
		},
	}
}

func parseDocCoverOffset(raw, flag string) (float64, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return 0, output.ErrValidation("%s must be a finite number in [0,1]", flag)
	}
	v, err := strconv.ParseFloat(trimmed, 64)
	if err != nil || math.IsNaN(v) || math.IsInf(v, 0) || v < 0 || v > 1 {
		return 0, output.ErrValidation("%s must be a finite number in [0,1]", flag)
	}
	return v, nil
}

func docCoverRelationError(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	if strings.Contains(strings.ToLower(msg), "relation") {
		return output.ErrWithHint(
			output.ExitAPI,
			"relation_mismatch",
			msg,
			coverRelationHint,
		)
	}
	return err
}
