Describe "Status checks against the simulation" {
    BeforeAll {
        Import-Module ..\bin\Debug\netstandard2.0\CloudChamber.Cmdlets.dll
    }

    BeforeEach {
        $sess = Connect-CcAccount -ClusterUri http://localhost:8080 -Name admin -Password adminPassword
    }

    AfterEach {
        Disconnect-CcAccount -Session $sess
    }

    It "Gets the simulation status" {
        $status = Get-CcCluster -Session $sess
        $span = New-TimeSpan -Hours 1
        $status.InactivityTimeout | Should Be $span

        $now = (Get-Date).ToUniversalTime()
        $status.FrontEndStartedAt | Should BeLessThan $now
    }

    It "Gets the summary list of current active sessions" {
        $list = Get-CcSessions -Session $sess
        $list.Count | Should BeGreaterThan 0

        $found = $false
        for ($i = 0; $i -lt $list.Count; $i++) {
            $entry = $list[$i]
            $item = Get-CcSession -Session $sess -Id $entry.Id
            $found = $found -or ($item.Name -eq "admin")
        }

        $found | Should Be $true
    }
}
