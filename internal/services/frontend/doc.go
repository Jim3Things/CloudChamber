
// Package frontend implements the primary service front end to receive the
// user HTTP requests, route the request to the appropriate libraries and/or
// micro-services and then format the response.
//
// File Layout
// ===========
//
// Each major object type is represented by a subtree root in the URL, and a
// specific source file associated with it.  These, and their associated
// URL subtrees, are:
//
// faults.go : /api/faults subtree
// Implements the service handler to retrieve information on faults, to inject
// faults into the simulated structure, to create and view devops actions.
//
// inventory.go : /api/racks subtree
// Implements the service handler to retrieve information about the simulated
// inventory.  This spans declared content, as well as target, observed, and
// current status.
//
// logs.go : /api/logs subtree
// Implements the service handler to retrieve log entries, or to wait for new
// ones.
//
// ping.go : /api/ping subtree
// Implements a simple service handler used to reset the inactivity timer for
// the logged in session.
//
// stepper.go : /api/stepper subtree
// Implements the service handler to provide the API to query and change the
// simulated time.
//
// users.go : /api/users subtree
// Implements the service handler to provide the API to manage user records, as
// well as logging into or out of a simulation.
//
// workloads.go : /api/workloads subtree
// Implements the service handler to provide the API to manage workload records.
//
// There are also a number of supporting files:
//
// frontend.go
// Base file containing the global definition for the package and the main() entry point
//
// Service Syntax
// ==============
//
// In the below definitions, the "{" and "}" characters surround an object name/key and are not included
// in the URI. For example, to read the record of the user with the name "Bob", the URI would be
//
// First, note that /api is a reserved subtree for all REST operations.  Any URI that is outside
// of that subtree is assumed to be a static file that needs to be served.
//
//
// /api/users/
// -----------
//
// GET /api/users
// list all the user records. At some point we may need to add a filter to restrict
// the set being returned.
//
// GET /api/users/{username}
// Returns a single user record for the user matching username or an error if the
// supplied username is not known.
//
// PUT /api/users/{username}
// Updates the record for the user matching username, or an error if the supplied
// document is invalid in some way, or if the supplied username is not known.
//
// POST - /api/users/{username}
// Creates a record for the user matching username, or an error if the supplied
// document is invalid in some way. If the user record is successfully created,
// the response will be an HTTP 201 (Created) status.
//
// If the supplied username is already known, and the supplied document exactly
// matches the existing record, this will be interpreted as a duplicate request
// and will result in an HTTP 200 (OK) status code. If the username already exists
// and the supplied document differs from the existing user record in any significant
// way, the response will be an HTTP 409 (Conflict) status.
//
// DELETE - /api/users/{username}
// Deletes the record for the user matching username, or an error is the supplied
// username is not known.
//
// PUT /api/users/{username}?op=[login|logout]
// Updates the user record the for user matching username according to the supplied
// operation, or returns and error if the supplied username is not known, or the
// operation code is invalid in some way.
//
// /api/workloads
// --------------
//
// GET - /api/workloads
// list all the workload records. At some point we may need to add a filter to restrict
// the set being returned.
//
// GET - /api/workloads/{workloadname}
// Returns a single workload record for the workload matching workloadname or an error if the
// supplied workloadname is not known.
//
// PUT - /api/workloads/{workloadname}
// Updates the record for the workload matching workloadname, or an error if the supplied
// document is invalid in some way, or if the supplied workloadname is not known.
//
// POST - /api/workloads/{workloadname}
// Creates a record for the workload matching workloadname, or an error if the supplied
// document is invalid in some way, or if the supplied workloadname is already known.
//
// DELETE - /api/workloads/{workloadname}
// Deletes the record for the workload matching workloadname, or an error is the supplied
// workloadname is not known.
//
// PUT /api/workloads/{workloadname}?op=[login|logout|enable|disable]
// Updates the workload record the for workload matching workloadname according to the supplied
// operation, or returns and error if the supplied workloadname is not known, or the
// operation code is invalid in some way.
//
// /api/racks/
// -----------
//
// GET - /api/racks  //Get the list of all known racks.
//
// GET - /api/racks/{rackid}  //Returns a single rack ID record.
//
// GET - /api/racks/{rack-id}/blades //Get list of known blades in a rack.
//
// GET - /api/racks/{rack-id}/TOR  //Get Top of Racks details .
//
// GET - /api/racks/{rack-id}/PDU //Gets Power distribution Unit details.
//
// GET - /api/racks/{rack-id}/blades/{blade-id} //Returns a record details of a single blade.
//
// GET - /api/racks/{rack-id}/TOR //To get details about a specific blade in a specific rack
//
// GET - /api/racks/rack-id/PDU //Gets Power distribution Unit details.
//
//
// /api/stepper
// ------------
//
// GET - /api/stepper
// Returns the current simulated time, mode and advance rate
//
// GET - /api/stepper/now
// Returns the current simulated time
//
// PUT - /api/stepper?advance
// Moves simulated time forward, so long as the mode is manual
// This relies on an ETag returned by either of the two preceding GET operations
//
// PUT - /api/stepper?mode={manual|automatic[=rate]}
// Sets the new mode and advance rate
// This relies on an ETag returned by the first GET operation above
//
// /api/faults
// -----------
//
// GET - /api/faults
//
// TODO
//
// /api/logs
//
// /api/faults
// /api/inventory
// will be a configuration file. you will be
// /point to cmd. Few of them per server.
// */
package frontend
