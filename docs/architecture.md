# Architecture

KHZ is a small Go CLI. It does not replace the shell or Git.

```text
existing command / Git
        |
        v
 direct argv runner + Git porcelain-v2 adapter
        |
        +--> versioned local evidence (.khz)
        |
        +--> deterministic states: OK/WARN/FAIL/SKIP/INFO
        |
        v
 human board / machine JSON
```

The core uses the Go standard library. PowerShell is an integration surface that invokes the compiled `khz` binary with argument arrays.

The board derives `STOP` from any `FAIL`, `REVIEW` from `WARN` or `SKIP`, and `PROCEED` otherwise. It never converts incomplete evidence into a global green result.
