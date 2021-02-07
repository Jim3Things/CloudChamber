Describe "Operations against users" {
    BeforeAll {
        Import-Module ..\bin\Debug\netstandard2.0\CloudChamber.Cmdlets.dll
    }

    BeforeEach {
        $sess = Connect-CcAccount -ClusterUri http://localhost:8080 -Name admin -Password adminPassword
    }

    AfterEach {
        Disconnect-CcAccount -Session $sess
    }

    It "Gets the list of known users" {
        $list = Get-CcUsers -session $sess

        $list.Count | Should BeGreaterThan 0

        # Find and verify the admin account
        $found = $false
        for ($i = 0; $i -lt $list.Count; $i++) {
            if ($list[$i].Name -ieq "admin") {
                $list[$i].Uri | Should Be "/api/users/admin"
                $list[$i].Protected | Should Be $true
                $found = $true
            }
        }

        $found | Should Be $true
    }

    It "Gets the details for the admin account" {
        $account = Get-CcUser -session $sess -Name "admin"
        $account.Name | Should Be "admin"
        $account.Enabled | Should Be $true
        $account.ManageAccounts | Should Be $true
    }

    It "Creates and deletes a test account" {
        $user = New-CcUser -Session $sess -Name "cliTest" -Enabled -Password "fooBar"
        $user.Name | Should Be "cliTest"
        $user.Enabled | Should Be $true
        $user.ManageAccounts | Should Be $false

        $msg = Remove-CcUser -Session $sess -Name "cliTest"
        $msg | Should Be "User cliTest deleted."
    }

    It "Tries to create an already created account" {
        try {
            New-CcUser -Session $sess -Name "admin" -Enabled -Password "bogus"
            $false | Should Be $true
        } catch [System.Net.Http.HttpRequestException] {
            $_.Exception.Message | Should Be "CloudChamber: user ""admin"" already exists`n"
        }
    }

    It "Tries to get details on a non-existent user" {
        try {
            Get-CcUser -Session $sess -Name "bogusUser"
            $false | Should Be $true
        } catch [System.Net.Http.HttpRequestException] {
            $_.Exception.Message | Should Be "CloudChamber: user ""bogususer"" not found`n"
        }
    }

    It "Tries to delete a non-existent user" {
        try {
            Remove-CcUser -Session $sess -Name "bogusUser"
            $false | Should Be $true
        } catch [System.Net.Http.HttpRequestException] {
            $_.Exception.Message | Should Be "CloudChamber: user ""bogususer"" not found`n"
        }
    }
}
