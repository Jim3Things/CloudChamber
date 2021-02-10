This subtree contains the main.go files for the different executables that
make up Cloud Chamber.  The core logic for each on is located in the correct
internal subdirectories.

The backend executables are:

- controllerd: the simulated controller service.
- inventoryd: the simulated inventory.
- sim_supportd: the simulation support services (time and fault injection) 
- web_server: the web service that acts as the front end to the Cloud Chamber system.

The executables interact via grpc.  The web_server also provides an http endpoint.

All of these executables implement the following command qualifiers:
- -config=\<configuration file path\>:  This provides a path to the configuration
file.  It defaults to '.'.  The configuration file name is assumed to be 
'cloudchamber.yaml'.

- -showConfig=\<bool\>: If true, this causes the executable to echo the effective
configuration settings to stdout, and then exit.  The output is in a form that
can be directly used as a configuration file.

Note that configuration files define all the services of Cloud Chamber in one
place.  This means that one common configuration file can be used, and it
automatically provides the endpoint information for the set of services to each
service.