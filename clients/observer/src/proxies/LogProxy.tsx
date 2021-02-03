// This module handles getting trace log data from the CloudChamber service.
// It is responsible for maintaining an ongoing listener, as well as cleaning
// up and translating the incoming Json stream into the internal format used
// throughout the rest of the UI.

import {getJson} from "./Session";

// +++ JSON data definitions

interface JsonPolicy {
    maxEntriesHeld: number
    firstId: number
}

interface JsonEntry {
    id: number
    entry: JsonLogEntry
}

interface JsonLogEvent {
    tick: number
    severity: string
    name: string
    text: string
    eventAction: string
    stackTrace: string
    spanId: string
    linkId: string
}

interface JsonLogEntry {
    name: string
    spanID: string
    parentID: string
    traceID: string
    status: string
    stackTrace: string
    event: JsonLogEvent[]
    infrastructure: boolean
    startingLink: string
    linkSpanID: string
    linkTraceID: string
    reason: string
}

interface JsonLogEntries {
    lastId: number
    missed: boolean
    entries: JsonEntry[]
}

// --- JSON data definitions

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
    event: LogEvent[]
    infrastructure: boolean
    startingLink: string
    linkSpanID: string
    linkTraceID: string
    reason: string
}

export interface LogEvent {
    tick: number
    severity: number
    name: string
    text: string
    eventAction: number
    stackTrace: string
    spanId: string
    linkId: string
}

export enum LogSeverity {
    Debug,
    Reason,
    Info,
    Warning,
    Error,
    Fatal
}

function LogSeverityToNumber(sev: string): number {
    switch (sev) {
        case "Debug":   return LogSeverity.Debug
        case "Reason":  return LogSeverity.Reason
        case "Info":    return LogSeverity.Info
        case "Warning": return LogSeverity.Warning
        case "Error":   return LogSeverity.Error
        default:        return LogSeverity.Info
    }
}

export enum LogEventType {
    Trace ,
    SpanStart,
    AddLink
}

function LogActionToNumber(action: string): number {
    switch (action) {
        case "SpanStart":   return LogEventType.SpanStart
        case "AddLink":     return LogEventType.AddLink
        case "Trace":       return LogEventType.Trace
        default:            return LogEventType.Trace
    }
}

// --- UI Internal data definitions

export interface LogArrivalHandler {
    (toHold: number, entries: LogEntries): any;
}

export class LogProxy {
    abortController : AbortController | undefined = undefined

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
        const request = new Request("/api/logs/policy", { method: "GET"})
        getJson<JsonPolicy>(request, this.getSignal())
            .then(policy => {
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
            const request = new Request("/api/logs?from=" + this.startId + "&for=100", { method: "GET"})
            getJson<JsonLogEntries>(request, this.getSignal())
                .then(jsonEntries => {
                    const entries = this.convertToInternal(jsonEntries)

                    this.startId = entries.lastId
                    this.notify(entries)
                    this.getLogs(lastEpoch)
                })
        }
    }

    convertToInternal(input: JsonLogEntries): LogEntries {
        let entries: LogEntries = {
            lastId: input.lastId,
            missed: input.missed,
            entries: new Array(input.entries.length)
        }

        for (let i = 0; i < input.entries.length; i++) {
            const jsonEntry = input.entries[i]

            const span: JsonLogEntry = {
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
                const jsonEvent = span.event[j]

                const inEvent = {
                    severity: "Debug",
                    text: "",
                    eventAction: "Trace",
                    spanId: nullSpanID,
                    ...jsonEvent
                }

                entry.event[j] = {
                    tick: inEvent.tick,
                    severity: LogSeverityToNumber(inEvent.severity),
                    name: inEvent.name,
                    text: inEvent.text,
                    eventAction: LogActionToNumber(inEvent.eventAction),
                    stackTrace: inEvent.stackTrace,
                    spanId: inEvent.spanId,
                    linkId: inEvent.linkId
                }
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
    getSignal() : AbortSignal | undefined {
        if (this.abortController === undefined) {
            return undefined
        }

        return this.abortController.signal
    }
}