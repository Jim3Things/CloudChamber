import {durationToRate, policyToMode, StepperMode, TimeContext} from "./StepperProxy";
import {getJson} from "./Session";
import {WatchResponse} from "../pkg/protos/services/requests";

export interface ChangeHandlerFunc {
    (cur: TimeContext): any
}

// WatchProxy runs a background async task that is always listening for change
// notifications from the CloudChamber service.  As those notifications arrive
// this proxy determines what changed, if anything, and posts the changes to
// the functions passed when the class is instantiated.
//
// Currently the notifications are limited to expiry, in order to force a
// refresh of the session inactivity time, and simulated time status changes.
//
// TODO: When REST calls fail we should update some UI to indicate that the
//       communications to the server are not working.  Need a notification
//       channel for that.
//
// TODO: The StatusResponse to internal TimeContext conversion is located in
//       the StepperProxy at the moment.  Re-evaluate as the watch function
//       matures.
//
export class WatchProxy {
    // Last seen time context
    cur: TimeContext = {
        mode: StepperMode.Paused,
        rate: 0,
        now: 0
    }

    // Last seen simulated time service settings epoch
    timeEpoch: number = 0

    abortController : AbortController | undefined = undefined

    // Background task generation (epoch), used to ensure any stale async
    // operations terminate.
    epoch: number = 0

    onChangeHandler?: ChangeHandlerFunc;

    constructor(handler: ChangeHandlerFunc) {
        this.onChangeHandler = handler
    }

    // Start the background watcher task, after ensuring that no existing
    // task will survive.
    start() {
        this.epoch++
        this.watch(this.epoch, this.cur.now, this.timeEpoch)
    }

    // Cancel the background task, lazily.
    cancel() {
        this.epoch++
        this.issueAbort()
    }

    // Issue the time change notification
    notify() {
        if (this.onChangeHandler) {
            this.onChangeHandler(this.cur);
        }
    }

    // watch is the background async thread that keeps a watch outstanding.
    watch(lastEpoch: number, tickParam: number, epochParam: number) {
        var tick = tickParam
        var epoch = epochParam

        if (lastEpoch === this.epoch) {
            const route = "/api/watch?tick=" + tick + "&epoch=" + epoch

            getJson<any>(new Request(route, {method: "GET"}), this.getSignal())
                .then((value: any) => {
                    const response = WatchResponse.fromJSON(value)
                    const sr = response.statusResponse

                    if (sr !== undefined) {
                        // Something about the simulated time changed, process the update.
                        tick = sr.now
                        epoch = sr.epoch

                        this.cur.mode = policyToMode(sr.policy)
                        this.cur.rate = durationToRate(sr.measuredDelay)
                        this.cur.now = tick
                        this.timeEpoch = epoch
                        this.notify()
                    }

                    this.watch(lastEpoch, tick, epoch)
                })
                .catch(() => {
                    // Retry on failure
                    window.setTimeout(() => this.watch(lastEpoch, tick, epoch), 500)
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
