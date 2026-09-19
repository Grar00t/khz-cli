Set-StrictMode -Version Latest

function Get-KhzExecutable {
    $cmd = Get-Command -Name khz -ErrorAction SilentlyContinue
    if (-not $cmd) {
        throw 'khz executable not found on PATH'
    }
    return $cmd.Source
}

function Invoke-KhzNative {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)]
        [string[]]$KhzArgs,
        [switch]$AsObject
    )
    $exe = Get-KhzExecutable
    $output = & $exe @KhzArgs
    $code = $LASTEXITCODE
    if ($AsObject) {
        $text = if ($null -eq $output) { '' } elseif ($output -is [array]) { $output -join "`n" } else { [string]$output }
        if ([string]::IsNullOrWhiteSpace($text)) {
            $obj = [pscustomobject]@{ exit_code = $code }
        }
        else {
            $obj = $text | ConvertFrom-Json
        }
        if ($code -ne 0) {
            $err = New-Object System.Management.Automation.ErrorRecord (
                (New-Object System.Exception "khz exited $code"),
                'KhzExit',
                [System.Management.Automation.ErrorCategory]::NotSpecified,
                $obj
            )
            $PSCmdlet.WriteError($err)
        }
        return $obj
    }
    if ($code -ne 0) {
        $PSCmdlet.WriteError((New-Object System.Management.Automation.ErrorRecord (
                    (New-Object System.Exception "khz exited $code"),
                    'KhzExit',
                    [System.Management.Automation.ErrorCategory]::NotSpecified,
                    $output
                )))
    }
    return $output
}

function Get-KhzStatus {
    [CmdletBinding()]
    param()
    Invoke-KhzNative -KhzArgs @('status', '--json') -AsObject
}

function Invoke-KhzRun {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true, ValueFromRemainingArguments = $true)]
        [string[]]$Command
    )
    $args = @('run', '--json', '--') + $Command
    Invoke-KhzNative -KhzArgs $args -AsObject
}

function Invoke-KhzCheck {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)]
        [string]$Name,
        [Parameter(Mandatory = $true, ValueFromRemainingArguments = $true)]
        [string[]]$Command
    )
    $args = @('check', '--name', $Name, '--json', '--') + $Command
    Invoke-KhzNative -KhzArgs $args -AsObject
}

function Get-KhzReceipt {
    [CmdletBinding(DefaultParameterSetName = 'List')]
    param(
        [Parameter(ParameterSetName = 'Show', Mandatory = $true, Position = 0)]
        [string]$Id
    )
    if ($PSCmdlet.ParameterSetName -eq 'Show') {
        Invoke-KhzNative -KhzArgs @('receipt', 'show', $Id)
        return
    }
    Invoke-KhzNative -KhzArgs @('receipt', 'list', '--json') -AsObject
}

function Test-KhzReceipt {
    [CmdletBinding()]
    param(
        [Parameter(Mandatory = $true)]
        [string]$Id
    )
    Invoke-KhzNative -KhzArgs @('receipt', 'verify', $Id, '--json') -AsObject
}

function Test-KhzPolicy {
    [CmdletBinding()]
    param()
    Invoke-KhzNative -KhzArgs @('policy', 'check', '--json') -AsObject
}

function Get-KhzDoctor {
    [CmdletBinding()]
    param()
    Invoke-KhzNative -KhzArgs @('doctor', '--json') -AsObject
}

function Enable-KhzArabicAliases {
    [CmdletBinding()]
    param()
    $map = [ordered]@{
        'حالة' = 'Get-KhzStatus'
        'شغل'  = 'Invoke-KhzRun'
        'تحقق' = 'Invoke-KhzCheck'
        'ايصال' = 'Get-KhzReceipt'
    }
    foreach ($alias in $map.Keys) {
        $existing = Get-Alias -Name $alias -ErrorAction SilentlyContinue
        if ($existing) {
            Write-Warning "alias '$alias' already exists -> $($existing.Definition); not overwritten"
            continue
        }
        Set-Alias -Name $alias -Value $map[$alias] -Scope Global
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
)
