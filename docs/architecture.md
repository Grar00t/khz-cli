# Architecture

KHZ is one Go binary (`cmd/khz`) plus an optional PowerShell module that shells out to that binary.

```text
argv  →  internal/cli  →  execx | gitx | policy | receipt
                              ↓
                         .khz/receipts/*.json
                         .khz/checks/*.json
                              ↓
                           board
```

## Packages

| package | role |
|---|---|
| `internal/cli` | flags, commands, exit codes |
| `internal/gitx` | `git status --porcelain=v2 -z --branch` |
| `internal/execx` | argv child process, no shell |
| `internal/receipt` | schema, SHA-256 seal, verify |
| `internal/policy` | local JSON policy v1 |
| `internal/board` | decision rows; FAIL never hidden |
| `internal/term` | color policy + control-sequence sanitize |
| `internal/redact` | best-effort argv redaction |
| `internal/atomicfile` | temp + fsync + rename |
| `internal/state` | `.khz/` layout |
| `internal/checkstore` | last named checks |
| `internal/i18n` | human strings only |

## Platforms

The binary is intended to build for:

```text
windows/amd64 windows/arm64
linux/amd64 linux/arm64
darwin/amd64 darwin/arm64
```

A platform is verified only when CI or a local run on that OS actually executed tests. See `.github/workflows/ci.yml`.

## Observation model

KHZ records what it is written to record. It does not wrap the child in a sandbox, eBPF, or a filesystem watcher. See [receipts.md](receipts.md).
