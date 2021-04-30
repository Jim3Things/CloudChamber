// This module contains the proxy handler for calling the REST Stepper service
// in the Cloud Chamber backend.

// Known stepper policies
import {getErrorDetails, getJson} from "./Session"
import {StepperPolicy} from "../pkg/protos/services/requests"
import {Duration} from "../pkg/protos/utils"

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
function modeToString(policy: StepperMode): string {
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
export function policyToMode(policy: StepperPolicy): StepperMode {
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
export function durationToRate(item: Duration | undefined): number {
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


export interface ErrorHandlerFunc {
    (text: string): any
}

// Utility class that provides a proxy to the Cloud Chamber Stepper REST service.
//
//
// TODO: Add proper tracking of the ETags to qualify the updates, and code to
//       resynchronize the ETags once they get out of sync.
//
export class StepperProxy {
    onErrorHandler?: ErrorHandlerFunc

    // Construct the proxy, with the notification handler, and kick off the
    // processing
    constructor(errorHandler: ErrorHandlerFunc) {
        this.onErrorHandler = errorHandler
    }


    // Manually move the time forward one tick
    advance() {
        getJson<any>(new Request("/api/stepper?advance", {method: "PUT"}))
            .then(() => {
            })
            .catch((msg: any) => this.sendError(msg))
    }

    // Set the simulated time mode and ticks-per-second rate
    setMode(mode: StepperMode, postfix: string): Promise<any> {
        const path = "/api/stepper?mode=" + modeToString(mode) + postfix
        const request = new Request(path, {method: "PUT"})
        request.headers.append("If-Match", "-1")

        return getJson<any>(request)
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

    changePolicy(policy: SetStepperPolicy, cur: TimeContext) {
        // Fake what will be the REST call to the Stepper service, including
        // a fake delay for the response.
        switch (policy) {
            case SetStepperPolicy.Pause:
                if (cur.mode !== StepperMode.Paused) {
                    this.setMode(StepperMode.Paused, "")
                        .then(() => {
                        })
                        .catch((msg: any) => this.sendError(msg))
                }
                break

            case SetStepperPolicy.Step:
                if (cur.mode !== StepperMode.Paused) {
                    this.setMode(StepperMode.Paused, "")
                        .then(() => {
                            this.advance()
                        })
                        .catch((msg: any) => this.sendError(msg))
                } else {
                    this.advance()
                }
                break

            case SetStepperPolicy.Run:
                if (cur.mode !== StepperMode.Running || cur.rate !== 1) {
                    this.setMode(StepperMode.Running, ":1")
                        .then(() => {
                        })
                        .catch((msg: any) => this.sendError(msg))
                }
                break

            case SetStepperPolicy.Faster:
                const rate = Math.min(cur.rate + 1, 5)
                this.setMode(StepperMode.Running, ":" + rate)
                    .then(() => {
                    })
                    .catch((msg: any) => this.sendError(msg))
                break
        }
    }
}
