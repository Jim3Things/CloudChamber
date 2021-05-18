// Organizer holds the traces, keyed by span ID, and a list of known root
// spans, in reverse order (newest first).

import {nullSpanID} from "../pkg/protos/log/entry"
import {GetAfterResponse_traceEntry} from "../pkg/protos/services/requests"
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
    }

    formatLink(spanID: string, traceID: string, linkID: string): string {
        return spanID + ":" + traceID + "@" + linkID
    }

    get(spanID: string): GetAfterResponse_traceEntry | undefined {
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
