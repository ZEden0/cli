# Docs CLI E2E Coverage

## Metrics
- Denominator: 11 leaf commands
- Covered: 6
- Coverage: 54.5%

## Summary
- TestDocs_CreateAndFetchWorkflow: proves `docs +create` and `docs +fetch`; key `t.Run(...)` proof points are `create as bot` and `fetch as bot`.
- TestDocs_CreateAndFetchWorkflowAsUser: proves the same shortcut pair with UAT injection via `create as user` and `fetch as user`; creates its own Drive folder fixture first, then reads back the created doc by token.
- TestDocs_UpdateWorkflow: proves `docs +update` via `update-title-and-content as bot`, then re-fetches the same doc in `verify as bot` to assert persisted title/content changes.
- TestDocs_CoverWorkflow: proves `docs +cover-update`, `docs +cover-get`, and `docs +cover-delete`; it first uploads a PNG with `docs +media-upload --parent-type docx_image --parent-node <document_id> --doc-id <document_id>` so the cover token is bound to the target Docx relation.
- Setup note: docs workflows create a Drive folder through `drive files create_folder` in `helpers_test.go`; that helper is external to the docs domain and is not counted here.
- Blocked area: remaining media, search, and whiteboard shortcuts still need deterministic fixtures and rollback assertions.

## Command Table

| Status | Cmd | Type | Testcase | Key parameter shapes | Notes / uncovered reason |
| --- | --- | --- | --- | --- | --- |
| ✓ | docs +create | shortcut | docs/helpers_test.go::createDocWithRetry; docs_create_fetch_test.go::TestDocs_CreateAndFetchWorkflowAsUser/create as user | `--folder-token`; `--title`; `--markdown` | helper asserts returned doc id |
| ✓ | docs +fetch | shortcut | docs_create_fetch_test.go::TestDocs_CreateAndFetchWorkflow/fetch as bot; docs_update_test.go::TestDocs_UpdateWorkflow/verify as bot; docs_create_fetch_test.go::TestDocs_CreateAndFetchWorkflowAsUser/fetch as user | `--doc <docToken>` | |
| ✓ | docs +cover-delete | shortcut | docs_cover_test.go::TestDocs_CoverWorkflow/delete cover as bot | `--doc <docToken>` | verifies standard success envelope and document id |
| ✓ | docs +cover-get | shortcut | docs_cover_test.go::TestDocs_CoverWorkflow/get cover as bot; docs_cover_test.go::TestDocs_CoverWorkflow/verify cover removed as bot | `--doc <docToken>` | asserts updated cover token, then absence after delete |
| ✓ | docs +cover-update | shortcut | docs_cover_test.go::TestDocs_CoverWorkflow/update cover as bot | `--doc`; `--token`; `--offset-ratio-x`; `--offset-ratio-y` | uses relation-bound token from `docs +media-upload --parent-type docx_image --parent-node <document_id> --doc-id <document_id>` |
| ✕ | docs +media-download | shortcut |  | none | no media fixture workflow yet |
| ✕ | docs +media-insert | shortcut |  | none | requires deterministic upload fixture and rollback assertions |
| ✕ | docs +media-preview | shortcut |  | none | requires deterministic media fixture |
| ✕ | docs +search | shortcut |  | none | search results are ambient and not yet stabilized for E2E |
| ✓ | docs +update | shortcut | docs_update_test.go::TestDocs_UpdateWorkflow/update-title-and-content as bot | `--doc`; `--mode overwrite`; `--markdown`; `--new-title` | |
| ✕ | docs +whiteboard-update | shortcut |  | none | requires whiteboard fixture and DSL-specific assertions |
