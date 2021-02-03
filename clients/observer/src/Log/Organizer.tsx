// Organizer holds the traces, keyed by span ID, and a list of known root
// spans, in reverse order (newest first).  It also tracks which spans are
// currently expanded, and which are not
import {LogEntry, nullSpanID} from "../proxies/LogProxy";

export class Organizer {
    roots: string[]

    all: Map<string, LogEntry>
    links: Map<string, string>

    expanded: Map<string, boolean>

    constructor(values: LogEntry[], cur?: Organizer) {
        this.roots = []
        this.all = new Map<string, LogEntry>()
        this.links = new Map<string, string>()
        this.expanded = new Map<string, boolean>()

        if (cur !== undefined) {
            for (const span of values) {
                this.all.set(span.spanID, span)

                const v = cur.expanded.get(span.spanID)
                if (v !== undefined && v) {
                    this.expanded.set(span.spanID, true)
                }
            }

            this.all.forEach((v): void => {
                if (v.parentID === nullSpanID) {
                    if ((v.linkSpanID !== nullSpanID) && this.all.has(v.linkSpanID))
                    {
                        const key = this.formatLink(v.linkSpanID, v.linkTraceID, v.startingLink)
                        this.links.set(key, v.spanID)
                    } else {
                        this.roots = [v.spanID, ...this.roots]
                    }
                }
            })
        }
    }

    formatLink(spanID: string, traceID: string, linkID: string): string {
        return spanID + ":" + traceID + "@" + linkID
    }

    getRoots() : string[] {
        return this.roots
    }

    get(spanID: string): LogEntry | undefined {
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
