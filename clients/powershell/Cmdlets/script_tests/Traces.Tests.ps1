Describe "Trace File Tests" {
    BeforeAll {
        Import-Module ..\bin\Debug\netstandard2.0\CloudChamber.Cmdlets.dll
    }

    BeforeEach {
        $sess = Connect-CcAccount -ClusterUri http://localhost:8080 -Name admin -Password adminPassword
    }

    AfterEach {
        Disconnect-CcAccount -Session $sess
    }

    It "Gets the current trace policy" {
        $policy = Get-CcTracePolicy -Session $sess
        $policy.MaxEntriesHeld | Should Be 1000
        $policy.FirstId | Should BeGreaterThan -2
    }

    It "Gets a set of traces from the start" {
        $traces = Get-CcTraces -Session $sess
        $traces.LastId | Should BeGreaterThan 0
        $traces.Entries.Count | Should Be 100
    }

    It "Gets a set of traces from a given point" {
        $traces = Get-CcTraces -Session $sess -From 10
        $traces.LastId | Should BeGreaterThan 0
        $traces.Entries.Count | Should Be 100
    }

    It "Gets a set of traces with a given length" {
        $traces = Get-CcTraces -Session $sess -From 10 -For 10
        $traces.LastId | Should BeGreaterThan 0
        $traces.Entries.Count | Should Be 10
    }
}
