// This module contains the proxy handler for calling the REST Stepper service
// in the Cloud Chamber backend.

// Utility class that provides a proxy to the Cloud Chamber Stepper REST service.
// TODO: This does not current issue a REST call, the effect is faked to allow
//       further progress on the UI.
//
// TODO: The effect of automatic or spontaneous time advance in the Stepper REST
//       service is not yet modeled here.  It will show up via calls to the
//       onChangeHandler callback function.
import {ChangeHandlerFunc, StepperMode, StepperPolicy, Timestamp} from "./StepperProxy";

export class MockStepperProxy {
    cur: Timestamp = {
        mode: StepperMode.Paused,
        rate: 0,
        now: 0
    }

    onChangeHandler?: ChangeHandlerFunc;

    constructor(handler: ChangeHandlerFunc) {
        this.onChangeHandler = handler;
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

    changePolicy = (policy: StepperPolicy) => {
        // Fake what will be the REST call to the Stepper service, including
        // a fake delay for the response.
        switch (policy) {
            case StepperPolicy.Pause:
                this.cur.mode = StepperMode.Paused;
                this.cur.rate = 0;
                break;

            case StepperPolicy.Step:
                this.cur.mode = StepperMode.Paused;
                this.cur.rate = 0;
                this.cur.now += 1;
                break;

            case StepperPolicy.Run:
                this.cur.mode = StepperMode.Running;
                this.cur.rate = 1;
                this.cur.now += 1;
                break;

            case StepperPolicy.Faster:
                this.cur.mode = StepperMode.Running;
                this.cur.rate = Math.min(this.cur.rate  + 1, 5);
                this.cur.now += 1;
                break;
        }

        // TODO: This fakes the response from the Stepper.
        setTimeout(() => this.notify(), 100);
    }
}
