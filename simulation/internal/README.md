This subtree contains the internal logic that makes up CloudChamber.  There are three basic types of packages below this
subroot: microservices, client-side packages that ease connections to other services, and support logic to integrate
external packages into CloudChamber.

# Services

Most of the logic is embedded in microservices.  Each microservice is in a directory under 'services'.  All interactions
with any microservice by the rest of CloudChamber is through gRPC.  External interactions are through RESTful patterns on http.

Do be aware that most microservices use the store both as long term state, and as notification for actions.  

The microservices are:

- frontend: Implements the web service functions for CloudChamber.

- monitor: Implements the inventory monitor service.  This internal service acts as the proxy communication point to the
elements of the inventory.  It is responsible for sending commands to the inventory elements, and for checking on current
health.

- repair_manager (inventory): Implements the logic to transition an inventory item from some current active state to the
designated target state. 

- repair_manager (workloads): Implements the logic to transition a workload from some current active state to the designated
target state.

- stepper: Implements the simulated time handling withing CloudChamber.  The service provides a global clock and control over
its progression.


[TODO: Scheduler, Fault Injector, Inventory]

# Clients
The client packages are there to simplify access to another service.  They run in the caller's context, and use a protocol
to access to the remote service.  These are located under 'clients', and are:

- store: Provides access to the ETCD store used by CloudChamber for metadata state

- timestamp: Provides simplified access to the stepper service

# Support Libraries
Finally, the remainder are packages that integrate with some other external package, and do not use a remote protocol to
do so.  These are:

- config: Implements the global configuration support, using the viper package for the actual config file parsing.

- tracing: Customizes the tracing infrastructure for CloudChamber.  It uses the opentelemetry.io packages for the actual tracing support.