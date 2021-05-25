// This module handles getting trace log data from the CloudChamber service.
// It is responsible for maintaining an ongoing listener, as well as cleaning
// up and translating the incoming Json stream into the internal format used
// throughout the rest of the UI.

import {getJson} from "./Session"
import {GetAfterResponse, GetAfterResponse_traceEntry, GetPolicyResponse} from "../pkg/protos/services/requests"
import {Severity} from "../pkg/protos/log/entry"

export class LogEntry extends GetAfterResponse_traceEntry {
    expanded: boolean
    maxSeverity: Severity

    constructor(source: GetAfterResponse_traceEntry) {
        super({})

        this.id = source.id
        this.entry = source.entry
        this.expanded = false
        this.maxSeverity = Severity.Debug
    }
}

export interface LogArrivalHandler {
    (toHold: number, entries: LogEntry[]): any;
}

export class LogProxy {
    abortController: AbortController | undefined = undefined

    epoch: number = 0

    startId: number = -1

    maxHeld: number = 100

    start(handler: LogArrivalHandler) {
        const request = new Request("/api/logs/policy", {method: "GET"})
        getJson<any>(request, this.getSignal())
            .then(jsonPolicy => {
                const policy = new GetPolicyResponse(jsonPolicy)
                this.startId = policy.firstId
                this.maxHeld = policy.maxEntriesHeld
                this.getLogs(handler, this.epoch)
            })
            .catch(() => {
                // Retry on failure
                window.setTimeout(() => this.start(handler), 100)
            })
    }

    cancelUpdates() {
        this.epoch++
        this.issueAbort()
    }

    getLogs(handler: LogArrivalHandler, lastEpoch: number) {
        if (lastEpoch === this.epoch) {
            const request = new Request("/api/logs?from=" + this.startId + "&for=100", {method: "GET"})
            getJson<any>(request, this.getSignal())
                .then(jsonMsg => {
                    const entries = new GetAfterResponse(jsonMsg)

                    this.startId = entries.lastId
                    handler(this.maxHeld, entries.entries.map((v) => new LogEntry(v)))
                    this.getLogs(handler, lastEpoch)
                })
                .catch(() => {
                    if (!Boolean(this.getSignal()?.aborted))
                    {
                        window.setTimeout(() => this.start(handler), 100)
                    }
                })
        }
    }

    // Issue the abort for any outstanding operation, assuming that aborts are
    // enabled (which they should be)
    issueAbort() {
        if (this.abortController !== undefined) {
            this.abortController.abort()
        }
    }

    // Get the listener to use to sign up for notification of an abort demand.
    getSignal(): AbortSignal | undefined {
        if (this.abortController === undefined) {
            return undefined
        }

        return this.abortController.signal
    }
}
