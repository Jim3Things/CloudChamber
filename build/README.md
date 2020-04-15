This directory contains the scripts used to build Cloud Chamber, 
along with associated scripts used to validate the build, and clean it up.

The procedures are:
- buildall.cmd: builds the full Cloud Chamber and places the result in the
\deployments directory.
- cleanall.cmd: deletes the results from buildall.cmd, effectively clearing 
the \deployements directory.
- run_all_uts.cmd: runs all unit tests that have been determined to be stable.
At the limit, this should the same coverage as running 'go test .\...' from the
CloudChamber directory

Future work:
- We should use a better build/make system than a simple command procedure.
- It should work for both building and cleaning.
- We should have UTs use some form of tagging so that we can just use the
'go test .\...' with some filter from the CloudChamber directory