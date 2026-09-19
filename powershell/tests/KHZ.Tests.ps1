BeforeAll {
    $ModulePath = Join-Path $PSScriptRoot '..\KHZ\KHZ.psd1'
    Import-Module $ModulePath -Force
}

Describe 'KHZ PowerShell module' {
    It 'exports the explicit function list' {
        $expected = @(
            'Enable-KhzArabicAliases',
            'Get-KhzDoctor',
            'Get-KhzReceipt',
            'Get-KhzStatus',
            'Invoke-KhzCheck',
            'Invoke-KhzRun',
            'Test-KhzPolicy',
            'Test-KhzReceipt'
        )
        $actual = (Get-Command -Module KHZ -CommandType Function).Name | Sort-Object
        $actual | Should -Be $expected
    }

    It 'contains no Invoke-Expression' {
        $text = Get-Content (Join-Path $PSScriptRoot '..\KHZ\KHZ.psm1') -Raw
        $text | Should -Not -Match '\bInvoke-Expression\b'
    }

    It 'returns an object from doctor' {
        $result = Get-KhzDoctor
        $result.khz_version | Should -Not -BeNullOrEmpty
        $result.os | Should -Not -BeNullOrEmpty
    }


    It 'invokes the KHZ binary through argument arrays' {
        $text = Invoke-KhzRun -Executable $env:KHZ_BIN -ArgumentList @('version') | Out-String
        $text | Should -Match '^khz 0\.1\.0-dev'
    }

    It 'propagates child command failure' {
        { Invoke-KhzRun -Executable $env:KHZ_BIN -ArgumentList @('definitely-not-a-command') } | Should -Throw
    }

    It 'returns structured status output' {
        $result = Get-KhzStatus
        $result.khz_version | Should -Be '0.1.0-dev'
        $result.git | Should -Not -BeNull
    }

    It 'refuses alias collisions' {
        Set-Alias -Name 'حالة' -Value Get-Item -Scope Global
        try { { Enable-KhzArabicAliases } | Should -Throw }
        finally { Remove-Item Alias:\حالة -ErrorAction SilentlyContinue }
    }
}
