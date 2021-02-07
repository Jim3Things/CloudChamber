// This module contains the proxy handler for calling the REST Stepper service
// in the Cloud Chamber backend.

// Known stepper policies
import {failIfError, getJson} from "./Session";

export enum StepperPolicy {
    Pause,
    Step,
    Run,
    Faster
}

// +++ Stepper mode handling

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
function stringToMode(jsonMode: string) : StepperMode {
    switch (jsonMode.toLowerCase()) {
        case "manual":
            return StepperMode.Paused

        case "measured":
            return StepperMode.Running

        default:
            return StepperMode.Paused
    }
}
// --- Stepper mode handling

// Hold the current simulated time context
export interface Timestamp {
    // The current simulated time's advance mode
    mode: StepperMode;

    // The ticks per second that simulate time advances, if in automatic mode
    rate: number;

    // The current simulated time tick
    now: number;
}

// Define the top level slice of the simulated time status from the server that
// we need here.
interface JsonStatus {
    policy: string
    measuredDelay: string
}

// +++ Timestamp handling

// Define the structure of the timestamp value from the server
interface JsonTimestamp {
    ticks?: string
}

// Get the simulated time tick value from the JSON timestamp
function tsToNumber(item: JsonTimestamp) : number {
    return item.ticks !== undefined ? +item.ticks : 0
}

// --- Timestamp handling

// +++ duration.Duration handling

// Define the structure parsed out from duration string
interface JsonDuration {
    seconds: number
    nanos: number
}

// Get the nanosecond component from the duration string
function parseNano(val: string) : number {
    let nanoIndex = val.indexOf("n")
    if (nanoIndex > -1) {
        return +val.substr(0, nanoIndex - 1)
    }

    return 0
}

// Convert the duration string into a duration structure
function parseDuration(duration: string) : JsonDuration {
    let val : JsonDuration = {seconds: 0, nanos: 0}

    let indexS = duration.indexOf("s")
    if (indexS > -1) {
        const segment1 = duration.substr(0, indexS - 1)
        val.seconds = +segment1

        val.nanos = parseNano(duration.substr(indexS + 1))
    } else {
        val.nanos = parseNano(duration)
    }

    return val
}

// Convert the duration structure value into a ticks-per-second rate
function durationToRate(item: JsonDuration) : number {
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
    (cur: Timestamp): any;
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
    cur: Timestamp = {
        mode: StepperMode.Paused,
        rate: 0,
        now: 0
    }

    abortController : AbortController | undefined = undefined

    epoch: number = 0

    onChangeHandler?: ChangeHandlerFunc;

    // Construct the proxy, with the notification handler, and kick off the
    // processing
    constructor(handler: ChangeHandlerFunc) {
        this.onChangeHandler = handler
    }

    // Get the initial status, load it as context, and then start the
    // background processing task to keep everything up to date.
    getStatus() {
        getJson<any>(new Request("/api/stepper", {method: "GET"}))
            .then((value: any) => {
                const status: JsonStatus = {...value}
                this.cur.mode = stringToMode(status.policy)
                this.cur.rate = durationToRate(parseDuration(status.measuredDelay))

                this.cur.now = tsToNumber({...value["now"]})

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
        getJson<JsonTimestamp>(new Request("/api/stepper?advance", {method: "PUT"}))
            .then((value) => {
                this.cur.rate = 0
                this.cur.mode = StepperMode.Paused

                this.cur.now = tsToNumber(value)
                this.notify()
            })
    }

    // Set the simulated time mode and ticks-per-second rate
    setMode(mode: StepperMode, postfix: string): Promise<any> {
        const path= "/api/stepper?mode=" + modeToString(mode) + postfix
        const request = new Request(path, {method: "PUT"})
        request.headers.append("If-Match", "-1")

        return fetch(request)
            .then((resp) => {
                failIfError(request, resp)

                this.cur.mode = mode
                return resp.blob()
            })
    }

    // Issue the time change notification
    notify() {
        if (this.onChangeHandler) {
            this.onChangeHandler(this.cur);
        }
    }

    // Notify the Stepper of a policy event.  Note that repeated calls are
    // passed to the Stepper, which allows for single stepping and for
    // increased automatic execution rates.

    changePolicy(policy: StepperPolicy) {
        // Fake what will be the REST call to the Stepper service, including
        // a fake delay for the response.
        switch (policy) {
            case StepperPolicy.Pause:
                this.setMode(StepperMode.Paused, "")
                    .then(() => {
                        this.cur.rate = 0

                        // Ensure the status bar mode gets updated
                        this.notify()
                    })
                break;

            case StepperPolicy.Step:
                this.setMode(StepperMode.Paused, "")
                    .then(() => this.advance())
                break;

            case StepperPolicy.Run:
                this.setMode(StepperMode.Running, ":1")
                    .then(() => {
                        this.cur.mode = StepperMode.Running;
                        this.cur.rate = 1;

                        // Ensure the status bar mode gets updated
                        this.notify()
                    })
                break;

            case StepperPolicy.Faster:
                const rate = this.cur.rate = Math.min(this.cur.rate + 1, 5)
                this.setMode(StepperMode.Running, ":" + rate)
                    .then(() => {
                        this.cur.mode = StepperMode.Running;
                        this.cur.rate = rate;

                        // Ensure the status bar mode gets updated
                        this.notify()
                    })
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
            getJson<JsonTimestamp>(request, this.getSignal())
                .then((value)=> {
                    this.cur.now = tsToNumber(value)
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
}
