# Contributing to KHZ

KHZ uses evidence-first engineering. Changes should be small enough to explain, test, and review.

## Development gate

Run before proposing a change:

```sh
gofmt -w .
go test ./...
go vet ./...
go test -race ./...
```

Add regression tests for defects. Do not weaken warning/error gates to make a change pass. Do not introduce shell-string execution, telemetry, network services, AI dependencies, or third-party dependencies without a concrete reviewed requirement.

## Commit scope

Keep unrelated refactors separate from behavior changes. Never commit generated binaries, `.khz/receipts/`, credentials, tokens, or machine-specific paths.
