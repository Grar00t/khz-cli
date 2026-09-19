# Receipts

Schema version: `1`.

File: `.khz/receipts/<receipt_id>.json` (atomic write).

## Fields

| field | meaning |
|---|---|
| `schema_version` | `1` |
| `receipt_id` | `khz_<UTC>_<randhex>` |
| `khz_version` | CLI version |
| `started_at` / `finished_at` | RFC3339Nano UTC |
| `duration_ms` | wall time |
| `cwd` | working directory |
| `operation` | `run` or `check` |
| `command` | resolved executable |
| `args_redacted` | argv after best-effort redaction |
| `exit_code` | child exit |
| `status` | `OK` or `FAIL` |
| `git_before` / `git_after` | porcelain v2 snapshots |
| `checks` | named check records |
| `policy_events` | local policy notes |
| `artifacts` | reserved; empty in Phase 1 |
| `side_effect_observability` | always `"partial"` for arbitrary commands |
| `receipt_hash` | SHA-256 hex |

## Hash

Canonical JSON is `encoding/json` of the struct with `receipt_hash` set to `""`, HTML escaping disabled, no extra newline. `receipt_hash` is `hex(sha256(canonical_bytes))`.

`khz receipt verify`:

- **OK** — schema 1, well-formed, hash matches
- **FAIL** — malformed JSON, missing required fields, unsupported `schema_version`, or hash mismatch (tamper)

## What a receipt does not prove

A valid hash means the JSON bytes were not modified after KHZ sealed them. It does not mean:

- the child wrote no files
- the child made no network calls
- no further processes were started
- the machine was sandboxed
