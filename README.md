# KHZ — Human Decision CLI

KHZ is not a terminal emulator.
KHZ is not a shell replacement.
KHZ is a structured decision and receipt layer for existing command-line workflows.

KHZ records command execution, Git state, checks, local policy gates, and verifiable receipts so humans and automation can reason from explicit evidence rather than terminal scrollback.

## Status

Early development. The current baseline intentionally implements only `khz version` and help while the durable repository is established. Phase 1 adds the documented decision, receipt, Git, policy, and PowerShell integration commands with tests before they are claimed as supported.

## Build

```sh
go build ./cmd/khz
go test ./...
go vet ./...
```

## Principles

- evidence before decision;
- deterministic exit codes;
- direct argv execution, not shell-string concatenation;
- no telemetry;
- no AI dependency;
- integrity is not semantic truth;
- receipts describe only what KHZ can actually observe.

## License

MIT.
