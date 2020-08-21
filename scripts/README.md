This directory contains the scripts used to deploy and start a
CloudChamber simulation. These are copied into the “deployments”
directory and form part of the deployment package.

# Deploy.cmd \[-Etcd | -NoEtcd\] \[-TargetDir \<path\>\]

Uses the results of the CloudChamber project builds and deploys
generated files to the specified directory. The build results are copied
from

  - %GoPath%\\src\\github.com\\Jim3Things\\CloudChamber\\deployments

  - %GoPath%\\src\\github.com\\Jim3Things\\cloud\_chamber\_react\_ts\\build

If specified, will also search for a copy of the etcd.exe and
etcdctl.exe binaries and add them to the deployment target. Searches for
the binaries at

  - %ETCDBINPATH%\\\<binary\>.exe

  - %GOPATH%\\bin\\\<binary\>.exe

  - %PATH%

# MonitorEtcd.cmd

Starts a Windows command window using etcdctl to watch for and changes
to keys/values under the standard CloudChamber namespace in the
currently running etcd instance.

# StartAll.cmd

Starts all the services and monitors for a complete CloudChamber
simulation. It uses the scripts

  - StartEtcd.cmd

  - MonitorEtcd.cmd

  - StartCloudChamber.cmd

# StartCloudChamber.cmd

Starts each of the CloudChamber simulation services in their own console
session using the binaries and configuration files from the deployment.
The services are

  - controllerd.exe

  - inventoryd.exe

  - sim\_supportd.exe

  - web\_server.exe

The services expect an etcd instance will already be running on the
port\[s\] identified in the configuration file.

# StartEtcd.cmd \[\<DataStorePath\>\]

Starts an etcd instance where the data store is located at the supplied
path. If the data store does not exist, the store is created.

If the parameter is not supplied, the store is located wither at the
location identified in the ETCDDATA environment variable, or if that is
not set, at %SystemDrive%\\etcd.

Available environment variables are

  - ETCDINSTANCE - name of the ETCD instance

  - ETCDNODEADDR - IP address of the ETCD instance

  - ETCDPORTCLNT - IP port to be used for communication with the client

  - ETCDDATA - directory where the ETCD data files are to be placed

Use the “–help” parameter to display the default values if the
environment variables are not defined.

# start\_cloud\_chamber.cmd \[DEPRACATED\]

Start the CloudChamber services. Superseded by StartCloudChamber.cmd
