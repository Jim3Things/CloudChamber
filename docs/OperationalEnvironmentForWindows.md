At present the only requirement for external (to Cloud Chamber)
components that need to be running is an etcd persistent store. This can
be placed wherever is convenient but typically the local node is used.

# Etcd Store

This document assumes the standard command script

  - github.com\\Jim3Things\\CloudChamber\\scripts\\startetcd.cmd

will be used to start etcd running locally using the default network
ports. If any other options are required, the start method and
configuration should be altered as required.

## Etcd Documentation

The primary etcd documentation can be found at <https://etcd.io/docs/>
with the most recent version (v3.4.0) being used.

## Installing Etcd

There are several methods which can be used to obtain the etcd binaries
and then install them. Any one of the methods below is sufficient, based
on personal preference

1.  Download a set of pre-built etcd binaries and install them in one of
    
    1.  a directory somewhere on the system %PATH%
    
    2.  or a directory pointed to by the %ETCDBINPATH% environment
        variable

2.  Build etcd from source by pulling the latest version from github.com
    into the local git repository which will install the generated
    binaries to %GOPATH%\\bin

## Using the Pre-Built Etcd Binaries

The link to the current release v3.4.9 can be found at
<https://github.com/etcd-io/etcd/releases/>

<https://github.com/etcd-io/etcd/releases/download/v3.4.9/etcd-v3.4.9-windows-amd64.zip>

which contains two binaries etcd.exe and etcdctl.exe. These binaries
should be placed in a convenient directory which should either be on the
system PATH or any other directory which is defined in the %ETCDBINPATH%
environment variable.

## Building from source

To build etcd and etcdctl from source, open a Windows Command line and
set it up to a local git repository and ensure the GO environment is
properly configured. Then

  - go get -v go.etcd.io/etcd

  - go get -v go.etcd.io/etcd/etcdctl

which should fetch the sources, build both etcd.exe and etcdctl.exe and
finally copy then to the %GOPATH%\\bin directory. They can be left in
place and run from the %GOPATH%\\bin.

## Running Etcd

There are several methods of running etcd but for the purposes of
CloudChamber, the assumption is that the

  - %GOPATH%\\src\\github.com\\Jim3Things\\CloudChamber\\Scripts\\startetcd.cmd

command script is being used. This script will attempt to locate a copy
of the etcd binary to use and will search the following locations in the
given order and will use the first copy it finds.

1.  In a directory pointer to with the ETCDBINPATH

2.  In the %GOPATH%\\bin directory

3.  In a directory included in the system %PATH% environment variable.

This will start an instance of etcd on the local machine using the
default network ports and place the data store in the etcd directory on
the SystemDrive (typically c:\\etcd). The instance will start in a new
window which will contain various log messages from the etcd instance as
it runs.

If this is not the first time etcd has been started, the new instance
will re-start using the same data store as the previous instance and all
the previously persisted data will be made available via the new
instance.

## Stopping Etcd

To stop etcd, just stop the etcd process. If etcd was started using the
startetcd.cmd script, a new window will have been opened with the output
of the etcd instance. Exiting from this window will stop the etcd
instance. Exiting from the instance will not lead to any lost data.

## Cleaning Out the Etcd Store

Just stopping an etcd instance will not lead to the data store
directories being cleaned up and removed. If there is a need to delete
all the instance data and control files, stop the instance and then just
delete the complete directory tree where the instance was storing its
data. If etcd was started using the startetcd.cmd script, this will
likely be %SystemDrive%\\etcd (typically c:\\etcd)

## Using the Etcdctl Utility

The following all assumes the etcdctl.exe utility was placed in the
%GOPATH%\\bin directory either as a result of building from source, or
by coping the pre-built binaries. If an alternative location was used,
the path should be modified to suit.

Although the running etcd instance will print various message to the
console window it opens, these messages are not particularly useful. To
query and control the etcd instance the etcdctl.exe utility is used.

If the etcd instance was started using the startetcd.cmd script, the
etcd instance will be running on the local machine using the default
ports. If some other method was used, the appropriate location will need
to be supplied to the etcdctl instance via the --endpoints (double ‘-‘
character’) option.

To verify etcdctl can see and communicate with the etcd instance, using
a Windows Command line type

  - %gopath%\\bin\\etcdctl member list

(assuming the etcdctl.exe binary was built from source) which should
reply with a message which includes the computers name and some URLs.

### Dumping values from the Etcd store

To dump a single value from the store use

  - %gopath%\\bin\\etcdctl get \<key\>

Or for a range of keys

  - %gopath%\\bin\\etcdctl get \<key-low\> \<key-high\>

For all keys with a known prefix

  - %gopath%\\bin\\etcdctl get --prefix \<prefix\>

For example

  - %gopath%\\bin\\etcdctl get --prefix /CloudChamber

### Deleting values from the Etcd Store

To delete a single value from the store use

  - %gopath%\\bin\\etcdctl del \<key\>

Or to delete a range of keys

  - %gopath%\\bin\\etcdctl del \<key-low\> \<key-high\>

Or all the keys with a known prefix

  - %gopath%\\bin\\etcdctl del --prefix \<prefix\>

This is a very dangerous command as it is very easy to accidentally
delete all the data in a store if the wrong prefix is specified.

For example, to delete all the temporary data written in the store test
pass, use

  - %gopath%\\bin\\etcdctl del --prefix /CloudChamber/v0.1/Test/

But if the “Test/” were left off, all the CloudChamber related data
would be removed. Be careful.

### Monitoring keys and values in the Etcd Store

It can sometime be very convenient to monitor what changes are being
made to an etcd store. This can be achieved with the etcdctl “watch”
command. For example, in a new Windows Command line

To watch a single key

  - %gopath%\\bin\\etcdctl watch \<key\>

Or to watch for changes to a range of keys

  - %gopath%\\bin\\etcdctl watch \<key-low\> \<key-high\>

Or to watch for changes to any key under a known prefix

  - %gopath%\\bin\\etcdctl watch --prefix \<prefix\>

For example, to watch for any activity to any Cloud Chamber related keys
use

  - %gopath%\\bin\\etcdctl watch --prefix /CloudChamber/v0.1/
