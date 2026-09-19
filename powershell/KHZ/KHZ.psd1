@{
    RootModule        = 'KHZ.psm1'
    ModuleVersion     = '0.1.0'
    GUID              = '7eaf4ca8-5e8f-42d7-982e-46e29df853f2'
    Author            = 'KHZ contributors'
    Description       = 'PowerShell integration for the KHZ Human Decision CLI.'
    PowerShellVersion = '5.1'
    FunctionsToExport = @(
        'Get-KhzStatus',
        'Invoke-KhzRun',
        'Invoke-KhzCheck',
        'Get-KhzReceipt',
        'Test-KhzReceipt',
        'Test-KhzPolicy',
        'Get-KhzDoctor',
        'Enable-KhzArabicAliases'
    )
    CmdletsToExport   = @()
    VariablesToExport = @()
    AliasesToExport   = @()
}
