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
        $status.Inactivity | Should Be $span

        $now = (Get-Date).ToUniversalTime()
        $status.Started | Should BeLessThan $now
    }

    It "Gets the summary list of current active sessions" {
        $now = (Get-Date).ToUniversalTime()
        $list = Get-CcSessions -Session $sess
        $list.Count | Should BeGreaterThan 0

        $found = $false
        for ($i = 0; $i -lt $list.Count; $i++) {
            $entry = $list[$i]
            $item = Get-CcSession -Session $sess -Id $entry.Id
            $found = $found -or ($item.UserName -eq "admin")
            $item.Expires | Should BeGreaterThan $now
        }

        $found | Should Be $true
    }
}
