// This directory holds the inventory monitor microservice.  It tracks the health reports from the simulated
// inventory, updating the actual inventory status accordingly.

// Note that it also tracks non-reporting.  Sustained non-reporting by an element of the inventory is interpreted as
// a failure of that element.

package monitor
