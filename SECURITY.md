# Security

## What KHZ records

Receipts store cwd, redacted argv, exit code, Git snapshots, and timestamps. They do **not** capture child stdout, the process environment, or a complete side-effect trace.

`side_effect_observability` is `"partial"` for `khz run` and `khz check`. Integrity of the JSON file (SHA-256) is not semantic truth about the command.

## Redaction

Arguments whose flags match `password`, `passwd`, `token`, `api-key`, `apikey`, `secret`, or `authorization` (case-insensitive) have their values replaced with `***` in receipts.

This is best-effort. KHZ cannot guarantee that arbitrary user-provided secrets never appear in argv positions it does not recognize, in filenames, or in Git paths.

KHZ never dumps the complete process environment.

## UI injection

Branch names, filenames, and check names are sanitized before board/status rendering so they cannot inject ANSI/control sequences into KHZ output.

## Policy

Local `.khz/policy.json` is not GitHub server-side branch protection and does not enforce repository permissions.

## Reporting

Open a GitHub issue on [Grar00t/khz-cli](https://github.com/Grar00t/khz-cli) for vulnerabilities in this repository. Do not file them against the unrelated [Grar00t/khz](https://github.com/Grar00t/khz) Python runtime.
