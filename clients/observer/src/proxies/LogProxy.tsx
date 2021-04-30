// This module handles getting trace log data from the CloudChamber service.
// It is responsible for maintaining an ongoing listener, as well as cleaning
// up and translating the incoming Json stream into the internal format used
// throughout the rest of the UI.

import {getJson} from "./Session"
import {GetAfterResponse, GetPolicyResponse} from "../pkg/protos/services/requests"

export interface LogArrivalHandler {
    (toHold: number, entries: GetAfterResponse): any;
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
        getJson<any>(request, this.getSignal())
            .then(jsonPolicy => {
                const policy = new GetPolicyResponse(jsonPolicy)
                this.startId = policy.firstId
                this.maxHeld = policy.maxEntriesHeld
                this.getLogs(this.epoch)
            })
            .catch(() => {
                // Retry on failure
                window.setTimeout(() => this.start(), 100)
            })
    }

    cancelUpdates() {
        this.epoch++
        this.issueAbort()
    }

    // Issue the time change notification
    notify(entries: GetAfterResponse) {
        if (this.onLogArrivalHandler) {
            this.onLogArrivalHandler(this.maxHeld, entries)
        }
    }

    getLogs(lastEpoch: number) {
        if (lastEpoch === this.epoch) {
            const request = new Request("/api/logs?from=" + this.startId + "&for=100", {method: "GET"})
            getJson<any>(request, this.getSignal())
                .then(jsonMsg => {
                    const entries = new GetAfterResponse(jsonMsg)

                    this.startId = entries.lastId
                    this.notify(entries)
                    this.getLogs(lastEpoch)
                })
                .catch(() => {
                    window.setTimeout(() => this.start(), 100)
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
