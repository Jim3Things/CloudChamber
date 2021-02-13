Describe "Operations against inventory" {
    BeforeAll {
        Import-Module ..\bin\Debug\netstandard2.0\CloudChamber.Cmdlets.dll
    }

    BeforeEach {
        $sess = Connect-CcAccount -ClusterUri http://localhost:8080 -Name admin -Password adminPassword
    }

    AfterEach {
        Disconnect-CcAccount -Session $sess
    }

    It "Gets the list of racks" {
        $list = Get-CcRacks -Session $sess

        $list.Racks.Count | Should Be 8

        # Find and verify the rack we'll use later
        $racks = $list.Racks
        $found = $false
        foreach ($key in $racks.Keys) {
            if ($key -ieq "rack1") {
                $found = $true
            }
        }
        $found | Should Be $true

        $list.MaxBladeCount | Should Be 8
        $capacity = $list.MaxCapacity
        $capacity.Cores | Should Be 32
        $capacity.MemoryInMb | Should Be 16834
        $capacity.DiskInGb | Should Be 240
        $capacity.NetworkBandwidthInMbps | Should Be 2048
        $capacity.Accelerators.Count | Should Be 0

        $rack = $list.Racks["rack1"]
        $rack.Uri | Should Be "/api/racks/rack1/"
    }

    It "Gets the details on one rack" {
        $rack = Get-CcRack -Session $sess -Name rack1

        $rack.Blades.Count | Should Be 8
        $blade = $rack.Blades[1]
        $blade.Cores | Should Be 16
        $blade.MemoryInMb | Should Be 16834
        $blade.DiskInGb | Should Be 240
        $blade.NetworkBandWidthInMbps | Should be 2048
        $blade.Arch | Should Be "X64"
    }

    It "Tries to get a missing rack" {
        { Get-CcRack -Session $sess -Name rackfoo } | `
            Should Throw "CloudChamber: rack ""rackfoo"" not found`n" `
                -ExceptionType [System.Net.Http.HttpRequestException]
    }

    It "Gets the details on one blade" {
        $blade = Get-CcBlade -Session $sess -Name rack1 -Id 1

        $blade.Cores | Should Be 16
        $blade.MemoryInMb | Should Be 16834
        $blade.DiskInGb | Should Be 240
        $blade.NetworkBandWidthInMbps | Should be 2048
        $blade.Arch | Should Be "X64"
    }

    It "Tries to get a missing blade" {
        { Get-CcBlade -Session $sess -Name rack1 -Id 10 } | `
            Should Throw "CloudChamber: blade 10 not found in rack ""rack1""`n" `
                -ExceptionType [System.Net.Http.HttpRequestException]

    }
}
