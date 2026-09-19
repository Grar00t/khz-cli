# Security Policy

## Supported versions

KHZ is pre-1.0. Security fixes target the latest `main` state until a stable release policy is published.

## Reporting

Open a GitHub security advisory for vulnerabilities when repository hosting supports it. Avoid publishing live credentials or exploit payloads containing secrets.

## Security boundary

KHZ records evidence around command execution. It is not a sandbox and does not prove absence of filesystem, network, registry, subprocess, or other side effects. Arbitrary-command receipts therefore use `side_effect_observability = "partial"`.

KHZ executes child programs directly from argv and does not use `sh -c`, `cmd /c`, or `Invoke-Expression` for ordinary command execution.
