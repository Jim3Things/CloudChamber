Describe "Operations against simulated time" {
    BeforeAll {
        Import-Module ..\bin\Debug\netstandard2.0\CloudChamber.Cmdlets.dll
    }

    BeforeEach {
        $sess = Connect-CcAccount -ClusterUri http://localhost:8080 -Name admin -Password adminPassword
    }

    AfterEach {
        try {
            Remove-CcUser -Session $sess -Name "cliTest"
        } catch {}

        try {
            Disconnect-CcAccount -Session $sess
        } catch {}
    }

    It "Gets the current time" {
        $tick = Get-CcTime -Session $sess
        $tick | Should BeGreaterThan -1
    }

    It "Single Steps the time manually" {
        $tag = Suspend-CcTime -Session $sess -Force
        $numTag = [long]$tag.trim('\"')
        $numTag | Should BeGreaterThan -1

        $tick = Get-CcTime -Session $sess
        $tick | Should BeGreaterThan -1

        $newTime = Step-CcTime -Session $sess
        $newTime - $tick | Should Be 1

        $tick = Get-CcTime -Session $sess
        $tick | Should Be $newTime
    }

    It "Fails to step the time manually" {
        $user = New-CcUser -Session $sess -Name "cliTest" -Enabled -Password "fooBar"

        $sess = Disconnect-CcAccount -Session $sess
        $sess = Connect-CcAccount -ClusterUri http://localhost:8080 -Name cliTest -Password fooBar

        $tick = Get-CcTime -Session $sess
        $tick | Should BeGreaterThan -1

        { $tag = Suspend-CcTime -Session $sess -Force } | `
            Should Throw "CloudChamber: permission denied`n" `
                -ExceptionType [System.Net.Http.HttpRequestException]
    }

    It "Double Steps the time manaully" {
        $tag = Suspend-CcTime -Session $sess -Force
        $numTag = [long]$tag.trim('\"')
        $numTag | Should BeGreaterThan -1

        $tick = Get-CcTime -Session $sess
        $tick | Should BeGreaterThan -1

        $newTime = Step-CcTime -Session $sess -Ticks 2
        $newTime - $tick | Should Be 2
    }

    It "Sets automatic time advancement and waits for a tick" {
        $policy = Get-CcTimePolicy -Session $sess
        $tag = $policy.ETag

        $tag = Suspend-CcTime -Session $sess -Revision $tag

        $policy = Get-CcTimePolicy -Session $sess
        $policy.Policy | Should Be Manual

        $ts = New-TimeSpan
        $policy.Delay | Should Be $ts

        $tag = $policy.ETag

        $tick = Get-CcTime -Session $sess
        $waitTick = $tick + 1

        $tag = Resume-CcTime -Session $sess -Rate 1 -Revision $tag

        $policy = Get-CcTimePolicy -Session $sess

        $ts = New-TimeSpan -Seconds 1
        $policy.Delay | Should Be $ts

        $policy.Policy | Should Be Measured
        $tag = $policy.ETag

        $res = Wait-CcTime -Session $sess -Until $waitTick

        $tag = Suspend-CcTime -Session $sess -Revision $tag

        $res.Now | Should BeGreaterThan $tick
    }
}

