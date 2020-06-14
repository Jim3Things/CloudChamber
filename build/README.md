This directory contains the scripts used to build Cloud Chamber, 
along with associated scripts used to validate the build, and clean it up.

# Build, Test, Clean
The procedures are:
- buildall.cmd: builds the full Cloud Chamber and places the result in the
\deployments directory.
- cleanall.cmd: deletes the results from buildall.cmd, effectively clearing 
the \deployements directory.
- test.cmd: runs unit tests.  Invoke with a test name, or with * to run all
known stable tests.  '-v' can be added as a trailing qualifier to get the
trace output when it runs.

Future work:
- We should use a better build/make system than a simple command procedure.
- It should work for both building and cleaning.

# Build environment setup

Additionally, the subtree 'dev_tools' contains support scripts and tools used to set up the
CloudChamber development environment.

- fetchall.cmd: initial retrieval of all required go packages
- updateall.cmd: update all required go packages