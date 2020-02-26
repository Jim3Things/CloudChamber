// This directory holds the primary frontend for the CloudChamber web service. It calls
// out to, or invokes the component libraries and microservices as nessesary that provide
// the actual implementation of the CloudChamber project.
//
//
//
// File Layout
// ===========
//
// frontend.go
// Base file comtaining the global defitnion for the package and the main() entry point
//
// files.go
// Implements the service handler to serve files back to user that comprise the primary
// UI to the client browser
//
// users.go
// Implements the service handler to provide the API to manage user records. Generally
// invoked by the UI files and scripts executing on the client browser.
//
// workloads.go
// Implements the service handler to provide the API to manage workload records. Generally
// invoked by the UI files and scripts executing on the client browser.
//
//
//
// Service Syntax
// ==============
//
// In the below definitions, the "{" and "}" characters surround an object name/key and are not included
// in the URI. For example, to read the record of the user with the name "Bob", the URI would be
//
// GET - /api/users/Bob
//
//
// The syntax of the commands to the web servers are
//
// /
// returns the root pages for the service which provide the primary interaction with
// the user. Any subsequent pages rereuied will also be served under this root path.
//
// GET - /api/users
// list all the user records. At some point we may need to add a filter to restrict
// the set being returned.
//
// GET - /api/users/{username}
// Returns a single user record for the user matching username or an error if the
// supplied username is not known.
//
// PUT - /api/users/{username}
// Updates the record for the user matching username, or an error if the supplied
// document is invalid in some way, or if the supplied username is not known.
//
// POST - /api/users/{username}
// Creates a record for the user matching username, or an error if the supplied
// document is invalid in some way, or if the supplied username is already known.
//
// DELETE - /api/users/{username}
// Deletes the record for the user matching username, or an error is the supplied
// username is not known.
//
// PUT /api/users/{username}?op=[login|logout|enable|disable]
// Updates the user record the for user matching username according to the supplied
// operation, or returns and error if the supplied username is not known, or the
// operation code is invalid in some way.
//
// /api/workloads
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
//
//
// TODO
// /api/logs
// /api/stepper
// /api/injector
// /api/inventory

package main
