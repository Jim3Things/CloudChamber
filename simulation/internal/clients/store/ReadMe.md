The persistent store component used by the project is etcd which
provides a mean of storing key/value pairs along with a selection of
other features. For the purposes of this document, the important feature
is the naming scheme used by the project.

# Production Namespace

For production, all persisted keys have a standard prefix of
"/CloudChamber/V0.1". Under this prefix, names fall into two classes,

1.  names used for discovery

2.  names used for entities

The primary difference is the manner in which these keys are
“discovered”. To support “list” type operation, the store needs to
be able to provide a means to locate all the listable elements of the
type being located. The underlying mechanism is a prefix search where
all the elements in the list have a common prefix.

For example, there is a key/value pair store for each user with an
account. All the user keys are stored with single common prefix so a
single prefix search will locate all the user records in a single query
(size limits allowing).

## Names Used for Discovery

The most fundamental elements search for are users, racks and workloads.

| Element   | Prefix used for initial List() operation |
| --------- | ---------------------------------------- |
| users     | /CloudChamber/V0.1/users                 |
| racks     | /CloudChamber/V0.1/racks                 |
| workloads | /CloudChamber/V0.1/workloads             |

Note the use of the plural when searching for the set of elements.

## Discovering Users

User records are listable and so to list the user records using a
prefix, or a key to retrieve a single user record, use the following
table to construct the prefix/key accordingly

| Prefix/Key | Location (excluding prefix /CloudChamber/V0.1) | Returns                  |
| ---------- | ---------------------------------------------- | ------------------------ |
| prefix     | /users                                         | List of user definitions |
| key        | /users/{username}                              | Single user details      |

Since there are no listable elements within a single user account, all
the per-user data can be returned as part of the user listing.

## Discovering Racks

Where appropriate, the returned list of items can then be used to issue
a second search for the next listable elements. Consider “racks”. Once
we have a list of racks, each rack contains a “pdu”, a “tor”, and a
collection of blades. To reference a specific rack a key can be
constructed using prefix containing the singular of the element, along
with some additional information to identify what part of the rack is of
interest. For example, each rack contains a single “pdu”, a single “tor”
and a set of blades. Use the following table to construct an appropriate
prefix/key to reference the tor or pdu records.

However, since there are multiple blades expected, the blade records are
listable using a prefix or the details for a specific blade using a key,
both of which can constructed using the following table.

| Prefix/Key | Location (excluding prefix /CloudChamber/V0.1) | Returns                       |
| ---------- | ---------------------------------------------- | ----------------------------- |
| prefix     | /racks                                         | List of rack definitions      |
| key        | /racks/{rackid}                                | Single rack details           |
| Key        | /rack/{rackid}/tor                             | Details of tor                |
| Key        | /rack/{rackid}/pdu                             | Details of pdu                |
| prefix     | /rack/{rackid}/blades                          | List of instance descriptions |
| key        | /rack/{rackid}/blades/{bladeid}                | Single blade details          |

Fortunately, nothing within a blade is listable, so there all the
per-blade data can be returned as part of the blade listing.

## Discovering Workloads

Workloads are a little more complicated as there is an additional level
so more care must be taken when forming prefixes and keys.

| Prefix/Key | Location (excluding prefix /CloudChamber/V0.1)      | Returns                                                 |
| ---------- | --------------------------------------------------- | ------------------------------------------------------- |
| prefix     | /workloads                                          | List of workload named definitions                      |
| key        | /workloads/{workloadid}                             | Single workload definition                              |
| prefix     | /workload/{workloadid}/instances                    | List of instance descriptions                           |
| key        | /workload/{workloadid}/instance/{instanceid}/actual | Actual state for specific instance of specific workload |
| key        | /workload/{workloadid}/instance/{instanceid}/target | target state for specific instance of specific workload |

# Test Usage and Test Data

For test usage, as enabled via the UseTestNamespace boolean option in
the configuration file, the standard prefix is changed to
"/CloudChamber/V0.1/Test/Xxxx” where the value of “Xxxx” is determined
by the value for the “UseUniqueInstance” config boolean option. If not
set (default) then “Xxxx” is “Standard” to form a standard prefix of

> "/CloudChamber/V0.1/Test/Standard”

whereas if “UseUniqueInstance” is set to true, then a timestamp is used
(determined by the store initiailization time) to allow the results from
individual test passes to be retained from pass to pass.

These prefixes give a number of useful behaviours. By having a project
specific prefix we should be able to share any etcd instance, assuming
of course that no one else uses the same prefix. By having a
well-defined test prefix, it becomes easy to monitor, control and
prevent any interference between test data and production data.

Also, if the “UseTestNamespace” is set and the “PreCleanStore” is also
set (UseUniqueInstance must NOT be set), then the store will
automatically remove ALL keys and values with the prefix
"/CloudChamber/V0.1/Test/Standard” to allow each test pass to start with
a clean namespace.

Further, by using the prefix "/CloudChamber/V0.1/Test" it becomes very
easy to manually monitor, identify or purge purge the store any and all
test related records using the etcdctl utility without having to disturb
the production data at all. This can be achieved with

etcdctl watch --prefix /CloudChamber/V0.1/
