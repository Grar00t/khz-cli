# KHZ

KHZ is not a terminal emulator.
KHZ is not a shell replacement.
KHZ is a structured decision and receipt layer for existing command-line workflows.

`khz` is a Go CLI. It runs real commands as argv (no `sh -c`, no `cmd /c`), records Git state from `git status --porcelain=v2 -z --branch`, and writes SHA-256-sealed JSON receipts. A human decision board prints **OK / WARN / FAIL / SKIP / INFO** as text. Color is never the only signal.

This repository is **khz-cli**. It is a new project. It does not replace or modify [Grar00t/khz](https://github.com/Grar00t/khz) (the Python GGUF runtime).

## Not this project

- not a PowerShell replacement
- not a Git replacement
- not a terminal emulator or shell language
- not a prompt theme
- not an AI wrapper
- not a web application

## Requirements

- Go 1.22 or newer to build
- `git` on PATH for Git status / check examples

## Install

```bash
git clone https://github.com/Grar00t/khz-cli.git
cd khz-cli
go build -o khz ./cmd/khz
```

## Commands

```text
khz version
khz doctor
khz status
khz status --json
khz run -- <command> [args...]
khz check --name <name> -- <command> [args...]
khz board
khz board --json
khz receipt list
khz receipt show <id>
khz receipt verify <id>
khz policy init
khz policy check
khz git status
khz git status --json
```

Global flags: `--json`, `--no-color`, `--color`, `--lang en|ar`, `-h`.

`NO_COLOR` and `TERM=dumb` disable ANSI. Non-TTY stdout is uncolored unless `--color` is set.

## Example

```bash
khz policy init
khz check --name unit -- git version
khz board
khz receipt list
khz receipt verify <id>
```

`khz check` treats child exit 0 as **OK** and any other exit as **FAIL**. KHZ itself then exits 0 or 1. The child's numeric code is stored on the receipt.

## Receipts

Written atomically under `.khz/receipts/<id>.json`. Hash is SHA-256 of canonical JSON with `receipt_hash` empty.

For arbitrary commands, `side_effect_observability` is always `"partial"`. A receipt does **not** prove: no filesystem writes, no network, no registry activity, no subprocess effects, or a sandbox.

See [docs/receipts.md](docs/receipts.md).

## Policy

`khz policy` is a small local JSON file. It is **not** GitHub branch protection.

See [docs/policy.md](docs/policy.md).

## PowerShell

Optional module: `powershell/KHZ`. It calls `khz ... --json` and `ConvertFrom-Json`. It does not use `Invoke-Expression`.

See [docs/powershell.md](docs/powershell.md).

## Tests

```bash
gofmt -l .
go test ./...
go vet ./...
go test -race ./...
```

## License

MIT. See [LICENSE](LICENSE).
