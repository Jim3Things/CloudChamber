// This directory holds the primary frontend for the CloudChamber web service. It calls
// out to, or invokes the component libraries and microservices as nessesary that provide
// the actual implementation of the CloudChamber project.

/*
Package frontend implements the primary service front end to receive the
user HTTP requests, route the request to the appropriate libraries
and/or micro-services and then format the response.

File Layout
===========

frontend.go
Base file containing the global defitnion for the package and the main() entry point

files.go
Implements the service handler to serve files back to user that comprise the primary
UI to the client browser

users.go
Implements the service handler to provide the API to manage user records. Generally
invoked by the UI files and scripts executing on the client browser.

workloads.go
Implements the service handler to provide the API to manage workload records. Generally
invoked by the UI files and scripts executing on the client browser.



Service Syntax
==============

In the below definitions, the "{" and "}" characters surround an object name/key and are not included
in the URI. For example, to read the record of the user with the name "Bob", the URI would be

GET - /api/users/Bob


The syntax of the commands to the web servers are

/
returns the root pages for the service which provide the primary interaction with
the user. Any subsequent pages rereuied will also be served under this root path.

GET - /api/users
list all the user records. At some point we may need to add a filter to restrict
the set being returned.

GET - /api/users/{username}
Returns a single user record for the user matching username or an error if the
supplied username is not known.

PUT - /api/users/{username}
Updates the record for the user matching username, or an error if the supplied
document is invalid in some way, or if the supplied username is not known.

POST - /api/users/{username}
Creates a record for the user matching username, or an error if the supplied
document is invalid in some way. If the user record is successfully created,
the reponse will be an HTTP 201 (Created) status.

If the supplied username is already known, and the supplied document excatly
matches the existing record, this will be interpreted as a duplicate request
and will result in an HTTP 200 (OK) status code. If the username already exists
and the supplied document differs from the existing user record in any significant
way, the response will be an HTTP 409 (Conflict) status.

DELETE - /api/users/{username}
Deletes the record for the user matching username, or an error is the supplied
username is not known.

PUT /api/users/{username}?op=[login|logout|enable|disable]
Updates the user record the for user matching username according to the supplied
operation, or returns and error if the supplied username is not known, or the
operation code is invalid in some way.

/api/workloads

GET - /api/workloads
list all the workload records. At some point we may need to add a filter to restrict
the set being returned.

GET - /api/workloads/{workloadname}
Returns a single workload record for the workload matching workloadname or an error if the
supplied workloadname is not known.

PUT - /api/workloads/{workloadname}
Updates the record for the workload matching workloadname, or an error if the supplied
document is invalid in some way, or if the supplied workloadname is not known.

POST - /api/workloads/{workloadname}
Creates a record for the workload matching workloadname, or an error if the supplied
document is invalid in some way, or if the supplied workloadname is already known.

DELETE - /api/workloads/{workloadname}
Deletes the record for the workload matching workloadname, or an error is the supplied
workloadname is not known.

PUT /api/workloads/{workloadname}?op=[login|logout|enable|disable]
Updates the workload record the for workload matching workloadname according to the supplied
operation, or returns and error if the supplied workloadname is not known, or the
operation code is invalid in some way.

GET - /api/racks
Get the list of all known racks.

GET - /api/racks/{rackid}/
Returns a single rack ID record.

GET - /api/racks/{rack-id}/blades

Get list of known blades in a rack.

GET - /api/racks/{rack-id}/TOR
Get Top of Racks details .

GET - /api/racks/{rack-id}/PDU
Gets Power distribution Unit details.

GET - /api/racks/{rack-id}/blades/{blade-id}

Returns a record details of a single blade.

GET - /api/racks/{rack-id}/TOR
To get details about a specific blade in a specific rack

GET - /api/racks/rack-id/PDU
Gets Power distribution Unit details.



TODO

/api/logs

/api/stepper
/api/injector
/api/inventory
will be a configuration file. you will be
/point to cmd. Few of them per server.
*/
package frontend
