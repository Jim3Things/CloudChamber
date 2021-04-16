// This module contains the proxy handler for calling the REST Stepper service
// in the Cloud Chamber backend.

// Known stepper policies
import {getErrorDetails, getJson} from "./Session";
import {StatusResponse, StepperPolicy} from "../pkg/protos/services/requests";
import {Duration} from "../pkg/protos/utils";

// +++ Stepper mode handling

//
export enum SetStepperPolicy {
    Pause,
    Step,
    Run,
    Faster,
}

// Known Stepper operating modes
export enum StepperMode {
    Paused,
    Running
}

// Convert the stepper mode into a string that the REST interface would
// recognize
function modeToString(policy: StepperMode) : string {
    switch (policy) {
        case StepperMode.Paused:
            return "manual"

        case StepperMode.Running:
            return "automatic"

        default:
            return "manual"
    }
}

// Convert the REST string name into a stepper mode value
function policyToMode(policy: StepperPolicy) : StepperMode {
    switch (policy) {
        case StepperPolicy.Invalid:
        case StepperPolicy.Manual:
            return StepperMode.Paused

        case StepperPolicy.Measured:
        case StepperPolicy.NoWait:
            return StepperMode.Running
    }

    return StepperMode.Paused
}
// --- Stepper mode handling

// Hold the current simulated time context
export interface TimeContext {
    // The current simulated time's advance mode
    mode: StepperMode;

    // The ticks per second that simulate time advances, if in automatic mode
    rate: number;

    // The current simulated time tick
    now: number;
}


// +++ duration.Duration handling


// Convert the duration structure value into a ticks-per-second rate
function durationToRate(item: Duration | undefined) : number {
    if (item === undefined || item === null) {
        return 1
    }

    const seconds = +item.seconds
    const nanoseconds = +item.nanos

    const tps = (seconds !== 0 ? 1 / seconds : 0) +
        (nanoseconds !== 0 ? 1_000_000_000 / nanoseconds : 0)

    return Math.round(tps)
}

// --- duration.Duration handling

// +++ REST handling support functions

// Signature for the notification handler that wants to receive stepper time
// change events
export interface ChangeHandlerFunc {
    (cur: TimeContext): any
}

export interface ErrorHandlerFunc {
    (text: string): any
}

// Utility class that provides a proxy to the Cloud Chamber Stepper REST service.
//
// TODO: When REST calls fail we should update some UI to indicate that the
//       communications to the server are not working.  Need a notification
//       channel for that.
//
// TODO: Add proper tracking of the ETags to qualify the updates, and code to
//       resynchronize the ETags once they get out of sync.
//
export class StepperProxy {
    cur: TimeContext = {
        mode: StepperMode.Paused,
        rate: 0,
        now: 0
    }

    abortController : AbortController | undefined = undefined

    epoch: number = 0

    onChangeHandler?: ChangeHandlerFunc;
    onErrorHandler?: ErrorHandlerFunc

    // Construct the proxy, with the notification handler, and kick off the
    // processing
    constructor(handler: ChangeHandlerFunc, errorHandler: ErrorHandlerFunc) {
        this.onChangeHandler = handler
        this.onErrorHandler = errorHandler
    }

    // Get the initial status, load it as context, and then start the
    // background processing task to keep everything up to date.
    getStatus() {
        getJson<any>(new Request("/api/stepper", {method: "GET"}))
            .then((value: any) => {
                this.updateStatus(value)

                // Update the UI at the start, so we can minimize the delay in
                // getting the initial screen
                this.notify()

                // ... and then start the periodic updates
                this.abortController = new AbortController()
                this.epoch++
                this.updateNow(this.epoch)
            })
            .catch(() => {
                // Retry on failure
                window.setTimeout(() => this.getStatus(), 100);
            })
    }

    cancelUpdates() {
        this.epoch++
        this.issueAbort()
    }

    // Manually move the time forward one tick
    advance() {
        getJson<any>(new Request("/api/stepper?advance", {method: "PUT"}))
            .then((item) => {
                this.updateStatus(item)

                this.notify()
            })
            .catch((msg: any) => this.sendError(msg))
    }

    // Set the simulated time mode and ticks-per-second rate
    setMode(mode: StepperMode, postfix: string): Promise<StatusResponse> {
        const path= "/api/stepper?mode=" + modeToString(mode) + postfix
        const request = new Request(path, {method: "PUT"})
        request.headers.append("If-Match", "-1")

        return getJson<any>(request)
            .then((item) => {
                return this.updateStatus(item)
            })
    }

    // Issue the time change notification
    notify() {
        if (this.onChangeHandler) {
            this.onChangeHandler(this.cur);
        }
    }

    sendError(msg: any) {
        getErrorDetails(msg, (details: string) => {
            if (this.onErrorHandler) {
                this.onErrorHandler(details)
            }
        })
    }

    // Notify the Stepper of a policy event.  Note that repeated calls are
    // passed to the Stepper, which allows for single stepping and for
    // increased automatic execution rates.

    changePolicy(policy: SetStepperPolicy) {
        // Fake what will be the REST call to the Stepper service, including
        // a fake delay for the response.
        switch (policy) {
            case SetStepperPolicy.Pause:
                if (this.cur.mode !== StepperMode.Paused) {
                    this.setMode(StepperMode.Paused, "")
                        .then(() => {
                            // Ensure the status bar mode gets updated
                            this.notify()
                        })
                        .catch((msg: any) => this.sendError(msg))
                }
                break;

            case SetStepperPolicy.Step:
                if (this.cur.mode !== StepperMode.Paused) {
                    this.setMode(StepperMode.Paused, "")
                        .then(() => {
                            this.advance()
                        })
                        .catch((msg: any) => this.sendError(msg))
                } else {
                    this.advance()
                }
                break;

            case SetStepperPolicy.Run:
                if (this.cur.mode !== StepperMode.Running || this.cur.rate !== 1) {
                    this.setMode(StepperMode.Running, ":1")
                        .then(() => {
                            // Ensure the status bar mode gets updated
                            this.notify()
                        })
                        .catch((msg: any) => this.sendError(msg))
                }
                break;

            case SetStepperPolicy.Faster:
                const rate = Math.min(this.cur.rate + 1, 5)
                this.setMode(StepperMode.Running, ":" + rate)
                    .then(() => {
                        // Ensure the status bar mode gets updated
                        this.notify()
                    })
                    .catch((msg: any) => this.sendError(msg))
                break;
        }
    }

    // Background update task - this waits for the 'next' tick, issues an
    // update to the UI, and re-arms itself. This is how the UI is largely
    // kept in sync with the server
    updateNow(lastEpoch: number) {
        const after = this.cur.now
        const request = new Request("/api/stepper/now?after=" + after, {method: "GET"})

        if (lastEpoch === this.epoch) {
            getJson<any>(request, this.getSignal())
                .then((item) => {
                    this.updateStatus(item)
                    this.notify()
                    this.updateNow(lastEpoch)
                })
                .catch(() => {
                    if (lastEpoch === this.epoch) {
                        // Retry on failure
                        window.setTimeout(() => this.updateNow(lastEpoch), 500)
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
    getSignal() : AbortSignal | undefined {
        if (this.abortController === undefined) {
            return undefined
        }

        return this.abortController.signal
    }

    updateStatus(value: any): StatusResponse {
        console.log(value)
        const status = StatusResponse.fromJSON(value)
        console.log(status)
        this.cur.mode = policyToMode(status.policy)
        this.cur.rate = durationToRate(status.measuredDelay)
        this.cur.now = status.now

        return status
    }
}
