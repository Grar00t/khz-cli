# Pester tests for the KHZ module. Run when pwsh + Pester are available.
# This environment may not have PowerShell; CI on windows-latest imports the module.

Describe 'KHZ module' {
    BeforeAll {
        $mod = Join-Path $PSScriptRoot 'KHZ.psd1'
        Import-Module $mod -Force
    }

    It 'exports the documented functions and no wildcards' {
        $cmds = (Get-Command -Module KHZ).Name | Sort-Object
        $cmds | Should -Contain 'Get-KhzStatus'
        $cmds | Should -Contain 'Invoke-KhzRun'
        $cmds | Should -Contain 'Invoke-KhzCheck'
        $cmds | Should -Contain 'Get-KhzReceipt'
        $cmds | Should -Contain 'Test-KhzReceipt'
        $cmds | Should -Contain 'Test-KhzPolicy'
        $cmds | Should -Contain 'Get-KhzDoctor'
        $cmds | Should -Contain 'Enable-KhzArabicAliases'
    }

    It 'module source does not contain Invoke-Expression' {
        $src = Get-Content -Raw (Join-Path $PSScriptRoot 'KHZ.psm1')
        $src | Should -Not -Match 'Invoke-Expression'
        $src | Should -Not -Match '\bIEX\b'
    }

    It 'Enable-KhzArabicAliases does not overwrite existing aliases' {
        if (Get-Alias -Name 'حالة' -ErrorAction SilentlyContinue) {
            Remove-Item -Path 'Alias:حالة' -Force -ErrorAction SilentlyContinue
        }
        Set-Alias -Name 'حالة' -Value 'Get-Date' -Scope Global
        $warn = $null
        Enable-KhzArabicAliases -WarningVariable warn -WarningAction SilentlyContinue
        (Get-Alias -Name 'حالة').Definition | Should -Be 'Get-Date'
        $warn | Should -Not -BeNullOrEmpty
        Remove-Item -Path 'Alias:حالة' -Force
    }
}
