# PowerShell Integration

The PowerShell module in `powershell/KHZ` is an integration adapter, not the KHZ execution core.

Exported functions:

- `Get-KhzStatus`
- `Invoke-KhzRun`
- `Invoke-KhzCheck`
- `Get-KhzReceipt`
- `Test-KhzReceipt`
- `Test-KhzPolicy`
- `Get-KhzDoctor`
- `Enable-KhzArabicAliases`

Structured functions call KHZ with `--json` and pipe the resulting JSON to `ConvertFrom-Json`. The module uses argument arrays and never uses `Invoke-Expression`.

Arabic aliases are opt-in for the current PowerShell session and refuse collisions.
