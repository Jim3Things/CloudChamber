// Organizer holds the traces, keyed by span ID, and a list of known root
// spans, in reverse order (newest first).  It also tracks which spans are
// currently expanded, and which are not
import {nullSpanID} from "../pkg/protos/log/entry"
import {GetAfterResponse_traceEntry} from "../pkg/protos/services/requests"

export class Organizer {
    roots: string[]

    all: Map<string, GetAfterResponse_traceEntry>
    links: Map<string, string>

    expanded: Map<string, boolean>

    constructor(values: GetAfterResponse_traceEntry[], cur?: Organizer) {
        this.roots = []
        this.all = new Map<string, GetAfterResponse_traceEntry>()
        this.links = new Map<string, string>()
        this.expanded = new Map<string, boolean>()

        for (const span of values) {
            const entry = span.entry
            this.all.set(entry.spanID, span)

            const v = cur?.expanded.get(entry.spanID)
            if (v !== undefined && v) {
                this.expanded.set(entry.spanID, true)
            }
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
        const val = this.expanded.get(key)
        if (val === undefined) {
            return false
        }

        return val
    }

    // switch the expanded/collapsed flag
    flip(key: string) {
        this.expanded.set(key, !this.isExpanded(key))
    }
}
