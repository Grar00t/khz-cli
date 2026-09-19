@{
    RootModule        = 'KHZ.psm1'
    ModuleVersion     = '0.1.0'
    GUID              = 'a7c3e91f-4b2d-4e8a-9c11-6d5f2b8e0a41'
    Author            = 'SULIMAN ALSHAMMARI'
    CompanyName       = 'KHZ'
    Copyright         = '(c) 2026 SULIMAN ALSHAMMARI. MIT License.'
    Description       = 'PowerShell integration for the KHZ decision CLI. Not a shell replacement.'
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
    AliasesToExport   = @()
    CmdletsToExport   = @()
    VariablesToExport = @()
    PrivateData       = @{
        PSData = @{
            Tags         = @('khz', 'cli', 'receipts', 'git')
            LicenseUri   = 'https://github.com/Grar00t/khz-cli/blob/main/LICENSE'
            ProjectUri   = 'https://github.com/Grar00t/khz-cli'
            ReleaseNotes = 'Phase 1 integration module.'
        }
    }
}
