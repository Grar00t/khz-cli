# PowerShell module

Path: `powershell/KHZ/`.

```powershell
Import-Module ./powershell/KHZ/KHZ.psd1
Get-KhzDoctor
Get-KhzStatus
Invoke-KhzRun -- git version
Invoke-KhzCheck -Name unit -- git version
Get-KhzReceipt
Test-KhzReceipt -Id <id>
Test-KhzPolicy
```

Each function invokes the `khz` executable with an argument array and `--json`, then `ConvertFrom-Json`. There is no `Invoke-Expression`.

The module does not modify `$PROFILE`.

## Arabic aliases (opt-in)

```powershell
Enable-KhzArabicAliases
```

Suggested aliases: `حالة`, `شغل`, `تحقق`, `ايصال`. Existing aliases are reported and **not** overwritten.

## Tests

Pester file: `powershell/KHZ/KHZ.Tests.ps1`. If PowerShell or Pester is not installed, those tests are `NOT_EXECUTED`.
