# KHZ — Human Decision CLI

KHZ is not a terminal emulator.
KHZ is not a shell replacement.
KHZ is a structured decision and receipt layer for existing command-line workflows.

KHZ converts command execution, Git machine state, checks, local policy gates, and verifiable local receipts into structured machine-readable evidence and concise human-readable decisions.

It does not add telemetry, accounts, a cloud backend, AI dependencies, a database, or a shell language.

## Commands

```text
khz version
khz doctor [--json]
khz status [--json]
khz run -- <command> [args...]
khz check --name <name> -- <command> [args...]
khz board [--json]
khz receipt list [--json]
khz receipt show <id>
khz receipt verify <id> [--json]
khz policy init
khz policy check [--json]
khz git status [--json]
```

Global human-output options:

```text
--lang en|ar
--no-color
```

Machine JSON field names remain English.

## Example workflow

The following commands produce results from the current machine; the README does not claim any particular result in advance.

```sh
khz policy init
khz git status
khz check --name tests -- go test ./...
khz board
khz receipt list
```

`khz check` maps child exit `0` to `OK` and nonzero exit to `FAIL`. `khz run` and `khz check` execute the supplied executable and argv directly; they do not concatenate a shell command string.

## Git adapter

KHZ calls:

```text
git status --porcelain=v2 -z --branch --untracked-files=all
```

and parses the machine format, including NUL-separated paths. Human-decorated `git status` output is not parsed.

## Decision board

States are explicit text: `OK`, `WARN`, `FAIL`, `SKIP`, `INFO`.

- any `FAIL` -> `STOP`
- otherwise any `WARN` or `SKIP` -> `REVIEW`
- otherwise -> `PROCEED`

Color is supplementary only and is disabled for `NO_COLOR`, `--no-color`, `TERM=dumb`, redirected output, or policy configuration.

## Receipts and evidence boundary

Receipts are versioned JSON under `.khz/receipts/` and carry a SHA-256 integrity hash. Arbitrary command receipts use:

```json
"side_effect_observability": "partial"
```

KHZ does **not** prove absence of filesystem writes, network activity, registry changes, subprocess effects, or complete sandboxing. Integrity is not semantic truth.

Argument redaction is best effort for common key names such as `password`, `passwd`, `token`, `api-key`, `apikey`, `secret`, and `authorization`. Arbitrary secrets can still appear if the user places them in forms KHZ does not recognize.

## Local state

```text
.khz/
  config.json
  policy.json
  receipts/
  checks/
```

`khz policy init` creates a `.khz/.gitignore` containing `receipts/` and refuses to overwrite an existing policy.

## Build and test

KHZ has no third-party Go module dependency.

```sh
gofmt -w .
go test ./...
go vet ./...
go test -race ./...
go build ./cmd/khz
```

GitHub Actions is configured for Ubuntu, Windows, and macOS and cross-compiles the core for:

- windows/amd64
- windows/arm64
- linux/amd64
- linux/arm64
- darwin/amd64
- darwin/arm64

A platform is considered verified only after its CI job actually runs successfully.

## PowerShell

`powershell/KHZ` contains an integration module for Windows PowerShell 5.1 and PowerShell 7+ semantics. It invokes the `khz` binary with argument arrays and contains no `Invoke-Expression`. Arabic aliases are opt-in through `Enable-KhzArabicAliases` and collision-checked.

See [docs/powershell.md](docs/powershell.md).

## Documentation

- [Architecture](docs/architecture.md)
- [Receipts](docs/receipts.md)
- [Policy](docs/policy.md)
- [PowerShell](docs/powershell.md)
- [Accessibility](docs/accessibility.md)
- [Security policy](SECURITY.md)
- [Contributing](CONTRIBUTING.md)

## License

MIT.
