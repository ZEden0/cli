# Document Cover Shortcuts

Use these commands for Docx document cover metadata. They support a Docx document ID, a Docx URL, or a wiki URL that resolves to Docx. Old Doc URLs are rejected.

## Get

```bash
lark-cli docs +cover-get --doc "<docx-token-or-url>"
```

The command prints JSON containing `document_id` and `cover`. When a cover exists, `cover` includes fields such as `token`, `offset_ratio_x`, and `offset_ratio_y`.

## Update

Cover update accepts an already uploaded `file_token`:

```bash
lark-cli docs +cover-update \
  --doc "<docx-token-or-url>" \
  --token "<file_token>" \
  --offset-ratio-x 0.2 \
  --offset-ratio-y 0.3
```

`--offset-ratio-x` and `--offset-ratio-y` are optional finite numbers in `[0,1]`. When an offset flag is omitted, the CLI omits that field from `update_cover.cover` and lets OpenAPI apply its default cropping behavior.

## Prepare A Local Cover Image

Local files stay a two-step flow. First upload the image as a `docx_image` resource bound to the target Docx document, then pass the returned `file_token` to `+cover-update`.

```bash
lark-cli docs +media-upload \
  --file ./cover.png \
  --parent-type docx_image \
  --parent-node "<document_id>" \
  --doc-id "<document_id>"

lark-cli docs +cover-update \
  --doc "<document_id>" \
  --token "<file_token>"
```

Do not use `docs +media-insert` to prepare a cover token. That command uploads media for a body image or file block; its token is bound to the body block relation and can fail cover update with relation mismatch. Do not reuse IM upload tokens or ordinary Drive file tokens either.

## Delete

```bash
lark-cli docs +cover-delete --doc "<docx-token-or-url>"
```

Deleting or updating a cover is a document write operation. The selected user or bot identity must be able to edit the target Docx.
