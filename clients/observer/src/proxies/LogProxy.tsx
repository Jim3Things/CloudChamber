// This module handles getting trace log data from the CloudChamber service.
// It is responsible for maintaining an ongoing listener, as well as cleaning
// up and translating the incoming Json stream into the internal format used
// throughout the rest of the UI.

import {getJson} from "./Session";
import {Entry, Event} from "../pkg/protos/log/entry";
import {GetAfterResponse, GetPolicyResponse} from "../pkg/protos/services/requests";

const nullTraceID: string = "00000000000000000000000000000000"
export const nullSpanID: string = "0000000000000000"
const missingSpanID: string = "Missing"

// +++ UI Internal data definitions
//
// This section contains the log entry data structures after translation and
// normalization

export interface LogEntries {
    lastId: number
    missed: boolean
    entries: LogEntry[]
}

export interface LogEntry {
    id: number

    name: string
    spanID: string
    parentID: string
    traceID: string
    status: string
    stackTrace: string
    event: Event[]
    infrastructure: boolean
    startingLink: string
    linkSpanID: string
    linkTraceID: string
    reason: string
}

// --- UI Internal data definitions

export interface LogArrivalHandler {
    (toHold: number, entries: LogEntries): any;
}

export class LogProxy {
    abortController: AbortController | undefined = undefined

    epoch: number = 0

    startId: number = -1

    maxHeld: number = 100

    onLogArrivalHandler?: LogArrivalHandler

    // Construct the proxy, with the notification handler, and kick off the
    // processing
    constructor(handler: LogArrivalHandler) {
        this.onLogArrivalHandler = handler
    }

    start() {
        const request = new Request("/api/logs/policy", {method: "GET"})
        getJson<GetPolicyResponse>(request, this.getSignal())
            .then(jsonPolicy => {
                const policy = GetPolicyResponse.fromJSON(jsonPolicy)
                this.startId = policy.firstId
                this.maxHeld = policy.maxEntriesHeld
                this.getLogs(this.epoch)
            })
            .catch(() => {
                // Retry on failure
                window.setTimeout(() => this.start(), 100);
            })
    }

    cancelUpdates() {
        this.epoch++
        this.issueAbort()
    }

    // Issue the time change notification
    notify(entries: LogEntries) {
        if (this.onLogArrivalHandler) {
            this.onLogArrivalHandler(this.maxHeld, entries);
        }
    }

    getLogs(lastEpoch: number) {
        if (lastEpoch === this.epoch) {
            const request = new Request("/api/logs?from=" + this.startId + "&for=100", {method: "GET"})
            getJson<GetAfterResponse>(request, this.getSignal())
                .then(jsonMsg => {
                    const jsonEntries = GetAfterResponse.fromJSON(jsonMsg)
                    const entries = this.convertToInternal(jsonEntries)

                    this.startId = entries.lastId
                    this.notify(entries)
                    this.getLogs(lastEpoch)
                })
        }
    }

    convertToInternal(input: GetAfterResponse): LogEntries {
        let entries: LogEntries = {
            lastId: input.lastId,
            missed: input.missed,
            entries: new Array(input.entries.length)
        }

        for (let i = 0; i < input.entries.length; i++) {
            const jsonEntry = input.entries[i]

            const span: Entry = {
                name: "",
                parentID: nullSpanID,
                spanID: missingSpanID,
                traceID: nullTraceID,
                status: "",
                stackTrace: "",
                event: [],
                infrastructure: false,
                startingLink: "",
                linkSpanID: nullSpanID,
                linkTraceID: nullTraceID,
                reason: "",
                ...jsonEntry.entry
            }

            let entry: LogEntry = {
                id: jsonEntry.id,
                name: span.name,
                spanID: span.spanID,
                parentID: span.parentID,
                traceID: span.traceID,
                status: span.status,
                stackTrace: span.stackTrace,
                event: new Array(span.event.length),
                infrastructure: span.infrastructure,
                startingLink: span.startingLink,
                linkSpanID: span.linkSpanID,
                linkTraceID: span.linkTraceID,
                reason: span.reason
            }

            for (let j = 0; j < span.event.length; j++) {
                entry.event[j] = span.event[j]
            }

            entries.entries[i] = entry
        }

        return entries
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
