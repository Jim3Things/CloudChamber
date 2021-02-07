This project produces the powershell cmdlets that act as client operations to a
CloudChamber cluster.

The directory structure is:

- .: contains the source files for the cmdlet logic.
- protos: contains the structure definitions that are passed as json-encoded text between the cmdlets
and the CloudChamber cluster.
- script_tests: contains the integration tests that exercise the cmdlet functions.

Note that protos initially contains .cs files that are hand written.  Over time
it will be the destination target for .pb.cs files that are generated from
common .proto definitions used by all parts of CloudChamber.

Also note that the script_tests scripts use pester, and are invoked by issuing
invoke-pester in that directory.
