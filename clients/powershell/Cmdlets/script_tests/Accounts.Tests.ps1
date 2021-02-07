Describe "Log in and log out as admin" {
    BeforeAll {
        Import-Module ..\bin\Debug\netstandard2.0\CloudChamber.Cmdlets.dll
    }

    It "Successfully logs in and out on an active cluster" {
        $sess = Connect-CcAccount -ClusterUri http://localhost:8080 -Name admin -Password adminPassword

        Disconnect-CcAccount -Session $sess
    }

    It "Fails to log in when no cluster" {
        try {
            $sess = Connect-CcAccount -ClusterUri http://localhost:8088 -Name admin -Password adminPassword
            $false | Should Be $true
        } catch [Exception]{
            $error
        }
    }
}
