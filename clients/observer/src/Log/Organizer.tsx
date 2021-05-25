// Organizer holds the traces, keyed by span ID, and a list of known root
// spans, in reverse order (newest first).

import {Action, nullSpanID, Severity} from "../pkg/protos/log/entry"
import {LogEntry} from "../proxies/LogProxy"

export class Organizer {
    roots: string[]

    all: Map<string, LogEntry>
    links: Map<string, string>

    constructor(values: LogEntry[]) {
        this.roots = []
        this.all = new Map<string, LogEntry>()
        this.links = new Map<string, string>()

        for (const span of values) {
            const entry = span.entry
            this.all.set(entry.spanID, span)
        }

        this.all.forEach((v): void => {
            const entry = v.entry
            if (entry.parentID === nullSpanID) {
                if ((entry.linkSpanID !== nullSpanID) && this.all.has(entry.linkSpanID)) {
                    const key = this.formatLink(entry.linkSpanID, entry.linkTraceID, entry.startingLink)
                    this.links.set(key, entry.spanID)
                } else {
                    this.roots = [entry.spanID, ...this.roots]
                }
            }
        })

        for (const item of this.roots) {
            this.upliftSeverity(item)
        }
    }

    upliftSeverity(root: string): Severity {
        let entry = this.all.get(root)
        if (entry != null) {
            if (entry.entry.event.length === 0) {
                entry.maxSeverity = Severity.Info
            } else {
                let startingSev = Severity.Debug

                entry.entry.event.forEach(v => {
                    switch (v.eventAction) {
                        case Action.SpanStart: {
                            const res = this.upliftSeverity(v.spanId)
                            startingSev = Math.max(startingSev, res)
                            break
                        }

                        case Action.AddLink: {
                            const res = this.upliftSeverity(v.linkId)
                            startingSev = Math.max(startingSev, res)
                            break
                        }

                        default:
                            startingSev = Math.max(startingSev, v.severity)
                            break
                    }
                })
                entry.maxSeverity = startingSev
            }

            return entry.maxSeverity
        }

        return Severity.Debug
    }

    formatLink(spanID: string, traceID: string, linkID: string): string {
        return spanID + ":" + traceID + "@" + linkID
    }

    get(spanID: string): LogEntry | undefined {
        return this.all.get(spanID)
    }

    getViaLink(spanID: string, traceID: string, linkID: string): string | undefined {
        return this.links.get(this.formatLink(spanID, traceID, linkID))
    }

    isExpanded(key: string): boolean {
        const val = this.all.get(key)
        return val !== undefined && val.expanded
    }
}
