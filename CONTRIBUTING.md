# Contributing

KHZ is a small Go CLI. Prefer the standard library. Do not add React, databases, servers, telemetry, or AI clients.

## Build

```bash
go build -o khz ./cmd/khz
```

## Verify

```bash
gofmt -l .
go test ./...
go vet ./...
go test -race ./...
```

## Rules

- Child processes: `os/exec` with argv. No `sh -c` / `cmd /c` unless a test documents why a platform has no other binary.
- Git: parse `git status --porcelain=v2 -z --branch` only.
- Untrusted text (branch names, filenames, check names) must pass `term.Sanitize` before UI output.
- Receipts: atomic write, SHA-256 seal, `side_effect_observability=partial` for arbitrary commands.
- Do not claim KHZ policy equals GitHub protection.
- JSON field names stay English. `--lang ar` is human-visible only.
- No secrets, no generated binaries, no machine-specific paths in commits.

## Pull requests

Keep diffs small. Add a regression test for the proven defect. Do not hide FAIL under a global PASS.
