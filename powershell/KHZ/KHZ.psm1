Set-StrictMode -Version 2.0

function Invoke-KhzJson {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)]
        [string[]] $Arguments
    )

    $text = & khz @Arguments
    $code = $LASTEXITCODE
    if ($code -ne 0) {
        throw "khz exited with code $code"
    }
    return ($text | Out-String | ConvertFrom-Json)
}

function Get-KhzStatus {
    [CmdletBinding()]
    param()
    Invoke-KhzJson -Arguments @('status', '--json')
}

function Invoke-KhzRun {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true, Position = 0)]
        [string] $Executable,
        [Parameter(Position = 1, ValueFromRemainingArguments = $true)]
        [string[]] $ArgumentList = @()
    )
    & khz 'run' '--' $Executable @ArgumentList
    $code = $LASTEXITCODE
    if ($code -ne 0) { throw "khz child command exited with code $code" }
}

function Invoke-KhzCheck {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)]
        [string] $Name,
        [Parameter(Mandatory = $true)]
        [string] $Executable,
        [string[]] $ArgumentList = @()
    )
    & khz 'check' '--name' $Name '--' $Executable @ArgumentList
    $code = $LASTEXITCODE
    if ($code -ne 0) { throw "khz check exited with code $code" }
}

function Get-KhzReceipt {
    [CmdletBinding()]
    param([Parameter(Mandatory = $true)][string] $Id)
    Invoke-KhzJson -Arguments @('receipt', 'show', $Id)
}

function Test-KhzReceipt {
    [CmdletBinding()]
    param([Parameter(Mandatory = $true)][string] $Id)
    $text = & khz 'receipt' 'verify' $Id '--json'
    $obj = $text | Out-String | ConvertFrom-Json
    if ($LASTEXITCODE -ne 0 -and $obj.valid) { throw 'khz returned inconsistent receipt verification state' }
    return $obj
}

function Test-KhzPolicy {
    [CmdletBinding()]
    param()
    Invoke-KhzJson -Arguments @('policy', 'check', '--json')
}

function Get-KhzDoctor {
    [CmdletBinding()]
    param()
    Invoke-KhzJson -Arguments @('doctor', '--json')
}

function Enable-KhzArabicAliases {
    [CmdletBinding()]
    param()

    $aliases = @{
        'حالة' = 'Get-KhzStatus'
        'شغل' = 'Invoke-KhzRun'
        'تحقق' = 'Invoke-KhzCheck'
        'ايصال' = 'Get-KhzReceipt'
    }

    foreach ($name in $aliases.Keys) {
        $existing = Get-Alias -Name $name -ErrorAction SilentlyContinue
        if ($null -ne $existing -and $existing.Definition -ne $aliases[$name]) {
            throw "Alias collision: $name already maps to $($existing.Definition)"
        }
    }

    foreach ($name in $aliases.Keys) {
        Set-Alias -Name $name -Value $aliases[$name] -Scope Global
    }
}

Export-ModuleMember -Function @(
    'Get-KhzStatus',
    'Invoke-KhzRun',
    'Invoke-KhzCheck',
    'Get-KhzReceipt',
    'Test-KhzReceipt',
    'Test-KhzPolicy',
    'Get-KhzDoctor',
    'Enable-KhzArabicAliases'
) -Alias @()
