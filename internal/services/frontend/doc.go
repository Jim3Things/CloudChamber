
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
// frontend.go : all URIs
// This is the starting point for the http processing.  It initializes the http
// service attributes, including the initialization of the modules above.
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
// etag.go
// This module implements functions to convert between a revision number and a
// legal ETag string.
//
// httpErrors.go
// This module implements custom errors designed to be returned in an HTTP
// response, and validation functions used on incoming HTTP requests.
//
// session_manager.go
// Implements the handling and lifecycle management for logged in user sessions.
//
//
// Finally, there are several files that provide access to the underlying data
// stores.  Each is responsible for mediating access to a single table.  These
// are:
//
// DBInventory.go: access to the inventory defined state table.
// DBInventoryActual.go: access to the inventory current state table.
// DBUsers.go: access to the known users state table.
//
//
// REST Service Syntax
// ===================
//
// In the below definitions, the "{" and "}" characters surround an object
// name/key and are not included in the URI. For example, to read the record
// of the user with the name "Bob", the URI would be "/api/users/Bob".
//
// First, note that /api is a reserved subtree for all REST operations.  Any
// URI that is outside of that subtree is assumed to be a static file that
// needs to be served.
//
// Second, note that most calls do not require ETag validation.  Those that do
// are called out in their description.
//
//
// /api/logs
// ----------
//
// GET /api/logs?from={startingID}&for={limit}
// Returns log entries, starting with the startingID value, up to the limit
// number of entries.  If the last log entry is not yet at the startingID, the
// call will wait for new log entries to be created before returning.
//
// GET /api/logs/policy
// Returns the current log stream policy, including the earliest known log ID
// and the maximum number of log entries that may be returned in a single call.
//
// /api/ping
//
// GET /api/ping
// This is a simple operation that resets the inactivity timeout on a logged in
// session.
//
// /api/racks
// -----------
//
// GET /api/racks
// Returns a summary list of all racks. This includes, for each rack, its name,
// the URI to use to get detailed information, and the maximum capacity and
// population limits for the blades in that rack.  This is sufficient to block
// in a graphical display of all racks.
//
// GET /api/racks/{rackid}
// Returns the detail information about the rack identified by rackid.  This
// includes information about the blades and rack-level support components,
// such as the TOR and PDU.
//
// GET /api/racks/{rackid}/blades
// Returns the summary list of the blades in the rack identified by rackid.
// This includes the blade ID number and associated URI.
//
// GET /api/racks/{rackid}/blades/{bladeid}
// Returns the detail information for the blade numbed by bladeid in the rack
// identified by rackid.
//
//
// /api/stepper
// ------------
//
// GET /api/stepper
// Returns the current simulated time service status, including policy revision,
// policy mode and advance rate
//
// GET /api/stepper/now
// Returns the current simulated time
//
// GET /api/stepper/now?after={duetime}
// Waits until the simulated time is equal or greater than the duetime.  It then
// returns the current simulated time.
//
// PUT - /api/stepper?advance[={ticks}]
// Moves simulated time forward, so long as the mode is manual.  The number of
// ticks to advance can be specified, with a default value of 1.
//
// PUT - /api/stepper?mode=(manual|automatic[={rate])
// Sets the new mode to either manual advance, or automatic.  If automatic, a
// rate in terms of ticks per second can be provided, with 1 tick per second as
// the default.
//
// This operation is normally gated by an If-Match on a previously returned
// revision ETag.
//
// /api/users/
// -----------
//
// GET /api/users
// list all the user records.
//
// GET /api/users/{username}
// Returns a single user record for the user matching username or an error if
// the supplied username is not known.
//
// PUT /api/users/{username}
// Updates the record for the user matching username, or an error if the
// supplied document is invalid in some way, or if the supplied username is not
// known.
//
// POST /api/users/{username}
// Creates a record for the user matching username, or an error if the supplied
// document is invalid in some way. If the user record is successfully created,
// the response will be an HTTP 201 (Created) status.
//
// If the supplied username is already known, and the supplied document exactly
// matches the existing record, this will be interpreted as a duplicate request
// and will result in an HTTP 200 (OK) status code.
//
// If the username already exists and the supplied document differs from the
// existing user record in any significant way, the response will be an HTTP
// 409 (Conflict) status.
//
// This operation is normally gated by an If-Match on a previously returned
// revision ETag.
//
// DELETE /api/users/{username}
// Deletes the record for the user matching username, or an error is the
// supplied username is not known or is protected against deletion.
//
// PUT /api/users/{username}?op=[login|logout]
// Updates the user record the for user matching username according to the
// supplied operation, or returns and error if the supplied username is not
// known, or the operation code is invalid in some way.
//
// PUT /api/users/{username}?password
// Updates the password for the user with value specified in the request body.
//
// This operation is normally gated by an If-Match on a previously returned
// revision ETag.
//
// Outstanding or Future items
// ===========================
//
// All REST calls that return collections of items need to support filtering at
// the service, and paging (the ability to incrementally return the response).
//
// Operations that wait need to support cancellation.
//
// Ping, and the various waiting operations, need to be merged into a single
// change notification call with a timeout.
//
// Update operations that do not require ETag validation need to be reviewed as
// to whether or not they should.
//
// Some REST subtrees are not currently implemented:
// /api/faults
// -----------
// (This is to be filled in as the design progresses)
//
// /api/racks
// ----------
// This subtree needs to add support for retrieval of the different viewpoints
// on the inventory: its defined state, its target state, the last observed
// state in the controller, and the current simulated state.
//
// This subtree needs to add support for modifying the target and defined states
// of the inventory.
//
// /api/workloads
// --------------
// (This is still a to be done subtree, and the list below is therefore likely
// incomplete and subject to change.)
//
// Much like /api/racks, this subtree needs to add support for retrieval of the
// different viewpoints: its defined state, its target state, the last observed
// state in the controller, and the current simulated state.
//
// GET - /api/workloads
// list all the workload records.
//
// GET - /api/workloads/{workloadname}
// Returns a single workload record for the workload matching workloadname or
// an error if the supplied workloadname is not known.
//
// PUT - /api/workloads/{workloadname}
// Updates the record for the workload matching workloadname, or an error if the
// supplied document is invalid in some way, or if the supplied workloadname is
// not known.
//
// POST - /api/workloads/{workloadname}
// Creates a record for the workload matching workloadname, or an error if the
// supplied document is invalid in some way, or if the supplied workloadname is
// already known.
//
// DELETE - /api/workloads/{workloadname}
// Deletes the record for the workload matching workloadname, or an error is the
// supplied workloadname is not known.
//
package frontend
