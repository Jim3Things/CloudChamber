/* eslint-disable */
import { asArray, asBool, asNumber, asItem, Duration, durationFromJson } from "../utils"

import { Timestamp } from "../common/Timestamp";
import { Entry } from "../log/entry";

export const protobufPackage = "services";

/** Define the various simulated time stepping policies */
export enum StepperPolicy {
  /** Invalid - Default value, indicates an uninitialized stepper */
  Invalid = 0,
  /**
   * NoWait - Policy that immediately moves the simulated time forward with any
   * wait operation.  Useful for shortening test runs.
   */
  NoWait = 1,
  /**
   * Measured - Policy that magnifies time, but still proceeds forward automatically.
   * This option requires a delay per tick to determine how fast time runs
   */
  Measured = 2,
  /**
   * Manual - Policy that requires manual stepping of time.  Simulated time only
   * moves forward as a result of an externally generated step command.
   */
  Manual = 3,
  UNRECOGNIZED = -1,
}

export function stepperPolicyFromJSON(object: any): StepperPolicy {
  if (object === undefined || object === null) {
    return StepperPolicy.Invalid
  }

  switch (object) {
    case 0:
    case "Invalid":
      return StepperPolicy.Invalid;
    case 1:
    case "NoWait":
      return StepperPolicy.NoWait;
    case 2:
    case "Measured":
      return StepperPolicy.Measured;
    case 3:
    case "Manual":
      return StepperPolicy.Manual;
    case -1:
    case "UNRECOGNIZED":
    default:
      return StepperPolicy.UNRECOGNIZED;
  }
}

export function stepperPolicyToJSON(object: StepperPolicy): string {
  switch (object) {
    case StepperPolicy.Invalid:
      return "Invalid";
    case StepperPolicy.NoWait:
      return "NoWait";
    case StepperPolicy.Measured:
      return "Measured";
    case StepperPolicy.Manual:
      return "Manual";
    default:
      return "UNKNOWN";
  }
}

// /** Define the request associated with a forcible reset */
// export interface ResetRequest {}

// /** Define the parameters to a stepper policy request parameters */
// export interface PolicyRequest {
//   /** Required policy (cannot be Invalid) */
//   policy: StepperPolicy;
//   /** Number of seconds between ticks.  Only valid for the "Measured" policy. */
//   measuredDelay: Duration | undefined;
//   /**
//    * If non-negative, require that the current policy revision number match
//    * Negative values do not require a match, and force an unconditional
//    * application of the new policy
//    */
//   matchEpoch: number;
// }

// /** Define the request associated with a Step operation */
// export interface StepRequest {}

// /** Define the request associated with a current time query */
// export interface NowRequest {}

// /** Define the parameters when requesting a delay */
// export interface DelayRequest {
//   /** The minimum simulated time before the delay is completed. */
//   atLeast: Timestamp | undefined;
//   /**
//    * An additional maximum number of ticks that can be added to the delay.
//    * A random number of ticks [0-jitter) are added to the delay time.  This
//    * is to simulate the random delays seen in, e.g., disk I/O or network
//    * communications.
//    */
//   jitter: number;
// }

// /** Define the request associated with a status request */
// export interface GetStatusRequest {}

// /**
//  * Internally used message to cause a simulate time advance based on timer
//  * expiry
//  */
// export interface AutoStepRequest {
//   /**
//    * The epoch number associated with the repeating timer call.  Ignore
//    * this message if this value does not match the last timer's epoch.
//    */
//   epoch: number;
// }

/** Define the current status response message */
export interface StatusResponse {
  /** Current stepper policy */
  policy: StepperPolicy;
  /** Current measured delay - should be zero if the policy is not "Measured" */
  measuredDelay: Duration;
  /** Current simulated time */
  now: number;
  /** Number of active waiters (number of outstanding delay calls) */
  waiterCount: number;
  /** Current policy version number */
  epoch: number;
}

// /**
//  * StepperState contains the state machine internal state necessary to restore
//  * the current simulated time on restart.
//  * NB: This is currently mostly latent - only the state machine state values are
//  * provided, in order to support the changes in the common state machine
//  * internals.
//  */
// export interface StepperState {
//   smState: StepperState_State;
// }

// export enum StepperState_State {
//   /** invalid - This is the state when no legal policy is in force. */
//   invalid = 0,
//   /** awaiting_start - This is the state prior to initialization. */
//   awaiting_start = 1,
//   /**
//    * no_wait - This state manages the policy where the simulated time is either
//    * manually stepped forward, or, if a Delay operation is called, it jumps
//    * forward to immediately complete any waiter.
//    */
//   no_wait = 2,
//   /**
//    * manual - This is the state where simulated time only moves forward due to
//    * specific Step operations.
//    */
//   manual = 3,
//   /**
//    * measured - This is the state where simulated time moves forward by one tick per
//    * the designated real time interval (e.g. 1 tick / second).
//    */
//   measured = 4,
//   /** faulted - An internal fault has occurred.  This is a terminal state. */
//   faulted = 5,
//   UNRECOGNIZED = -1,
// }

// export function stepperState_StateFromJSON(object: any): StepperState_State {
//   switch (object) {
//     case 0:
//     case "invalid":
//       return StepperState_State.invalid;
//     case 1:
//     case "awaiting_start":
//       return StepperState_State.awaiting_start;
//     case 2:
//     case "no_wait":
//       return StepperState_State.no_wait;
//     case 3:
//     case "manual":
//       return StepperState_State.manual;
//     case 4:
//     case "measured":
//       return StepperState_State.measured;
//     case 5:
//     case "faulted":
//       return StepperState_State.faulted;
//     case -1:
//     case "UNRECOGNIZED":
//     default:
//       return StepperState_State.UNRECOGNIZED;
//   }
// }

// export function stepperState_StateToJSON(object: StepperState_State): string {
//   switch (object) {
//     case StepperState_State.invalid:
//       return "invalid";
//     case StepperState_State.awaiting_start:
//       return "awaiting_start";
//     case StepperState_State.no_wait:
//       return "no_wait";
//     case StepperState_State.manual:
//       return "manual";
//     case StepperState_State.measured:
//       return "measured";
//     case StepperState_State.faulted:
//       return "faulted";
//     default:
//       return "UNKNOWN";
//   }
// }

// /** Specify the trace entry to append to the trace sink's store */
// export interface AppendRequest {
//   /** The trace entry */
//   entry: Entry | undefined;
// }

// /** Specify what trace entries to receive */
// export interface GetAfterRequest {
//   /**
//    * The last id that has been previously seen.  Note that '-1' will cause the
//    * earliest traces to be returned
//    */
//   id: number;
//   //* The maximum number of entries to return in in the reply. 
//   maxEntries: number;
//   /**
//    * True, if the call should wait for new trace entries if there are none
//    * that are later than the specified id when the call arrives.  If false,
//    * the call returns immediately with no entries.
//    */
//   wait: boolean;
// }

/** Return the traces for the request */
export interface GetAfterResponse {
  /**
   * The highest trace id returned.  Use this as the id on the next GetAfter
   * call in order to start returning the traces that immediately follow.
   */
  lastId: number;
  /**
   * True, if some entries were skipped - probably due to removal at the
   * trace sink in order to stay within the retention limit
   */
  missed: boolean;
  /** Set of trace entries we're returning */
  entries: GetAfterResponse_traceEntry[];
}

export interface GetAfterResponse_traceEntry {
  /** Sequential id for this trace */
  id: number;
  /** Contents of the trace entry */
  entry: Entry;
}

/** (Empty) payload to request the current policy */
// export interface GetPolicyRequest {}

/** Return the active policies for the trace sink service */
export interface GetPolicyResponse {
  /** The limit on the number of entries held in the trace sink */
  maxEntriesHeld: number;
  /** Earliest trace entry currently held. */
  firstId: number;
}

export interface WatchResponse {
  expired: boolean | undefined;
  statusResponse: StatusResponse | undefined;
}

// const baseResetRequest: object = {};

// export const ResetRequest = {
//   fromJSON(_: any): ResetRequest {
//     const message = { ...baseResetRequest } as ResetRequest;
//     return message;
//   },

//   toJSON(_: ResetRequest): unknown {
//     const obj: any = {};
//     return obj;
//   },
// };

// const basePolicyRequest: object = { policy: 0, matchEpoch: 0 };

// export const PolicyRequest = {
//   fromJSON(object: any): PolicyRequest {
//     const message = { ...basePolicyRequest } as PolicyRequest;
//     if (object.policy !== undefined && object.policy !== null) {
//       message.policy = stepperPolicyFromJSON(object.policy);
//     } else {
//       message.policy = 0;
//     }
//     if (object.measuredDelay !== undefined && object.measuredDelay !== null) {
//       message.measuredDelay = Duration.fromJSON(object.measuredDelay);
//     } else {
//       message.measuredDelay = undefined;
//     }
//     if (object.matchEpoch !== undefined && object.matchEpoch !== null) {
//       message.matchEpoch = Number(object.matchEpoch);
//     } else {
//       message.matchEpoch = 0;
//     }
//     return message;
//   },

//   toJSON(message: PolicyRequest): unknown {
//     const obj: any = {};
//     message.policy !== undefined &&
//       (obj.policy = stepperPolicyToJSON(message.policy));
//     message.measuredDelay !== undefined &&
//       (obj.measuredDelay = message.measuredDelay
//         ? Duration.toJSON(message.measuredDelay)
//         : undefined);
//     message.matchEpoch !== undefined && (obj.matchEpoch = message.matchEpoch);
//     return obj;
//   },
// };

// const baseStepRequest: object = {};

// export const StepRequest = {
//   fromJSON(_: any): StepRequest {
//     const message = { ...baseStepRequest } as StepRequest;
//     return message;
//   },

//   toJSON(_: StepRequest): unknown {
//     const obj: any = {};
//     return obj;
//   },
// };

// const baseNowRequest: object = {};

// export const NowRequest = {
//   fromJSON(_: any): NowRequest {
//     const message = { ...baseNowRequest } as NowRequest;
//     return message;
//   },

//   toJSON(_: NowRequest): unknown {
//     const obj: any = {};
//     return obj;
//   },
// };

// const baseDelayRequest: object = { jitter: 0 };

// export const DelayRequest = {
//   fromJSON(object: any): DelayRequest {
//     const message = { ...baseDelayRequest } as DelayRequest;
//     if (object.atLeast !== undefined && object.atLeast !== null) {
//       message.atLeast = Timestamp.fromJSON(object.atLeast);
//     } else {
//       message.atLeast = undefined;
//     }
//     if (object.jitter !== undefined && object.jitter !== null) {
//       message.jitter = Number(object.jitter);
//     } else {
//       message.jitter = 0;
//     }
//     return message;
//   },

//   toJSON(message: DelayRequest): unknown {
//     const obj: any = {};
//     message.atLeast !== undefined &&
//       (obj.atLeast = message.atLeast
//         ? Timestamp.toJSON(message.atLeast)
//         : undefined);
//     message.jitter !== undefined && (obj.jitter = message.jitter);
//     return obj;
//   },
// };

// const baseGetStatusRequest: object = {};

// export const GetStatusRequest = {
//   fromJSON(_: any): GetStatusRequest {
//     const message = { ...baseGetStatusRequest } as GetStatusRequest;
//     return message;
//   },

//   toJSON(_: GetStatusRequest): unknown {
//     const obj: any = {};
//     return obj;
//   },
// };

// const baseAutoStepRequest: object = { epoch: 0 };

// export const AutoStepRequest = {
//   fromJSON(object: any): AutoStepRequest {
//     const message = { ...baseAutoStepRequest } as AutoStepRequest;
//     if (object.epoch !== undefined && object.epoch !== null) {
//       message.epoch = Number(object.epoch);
//     } else {
//       message.epoch = 0;
//     }
//     return message;
//   },

//   toJSON(message: AutoStepRequest): unknown {
//     const obj: any = {};
//     message.epoch !== undefined && (obj.epoch = message.epoch);
//     return obj;
//   },
// };

export const StatusResponse = {
  fromJSON(object: any): StatusResponse {
    return {
      policy: stepperPolicyFromJSON(object.policy),
      measuredDelay: durationFromJson(object.measuredDelay),
      now: asNumber(object.now),
      waiterCount: asNumber(object.waiterCount),
      epoch: asNumber(object.epoch),
    }
  },
};

// const baseStepperState: object = { smState: 0 };

// export const StepperState = {
//   fromJSON(object: any): StepperState {
//     const message = { ...baseStepperState } as StepperState;
//     if (object.smState !== undefined && object.smState !== null) {
//       message.smState = stepperState_StateFromJSON(object.smState);
//     } else {
//       message.smState = 0;
//     }
//     return message;
//   },

//   toJSON(message: StepperState): unknown {
//     const obj: any = {};
//     message.smState !== undefined &&
//       (obj.smState = stepperState_StateToJSON(message.smState));
//     return obj;
//   },
// };

// const baseAppendRequest: object = {};

// export const AppendRequest = {
//   fromJSON(object: any): AppendRequest {
//     const message = { ...baseAppendRequest } as AppendRequest;
//     if (object.entry !== undefined && object.entry !== null) {
//       message.entry = Entry.fromJSON(object.entry);
//     } else {
//       message.entry = undefined;
//     }
//     return message;
//   },

//   toJSON(message: AppendRequest): unknown {
//     const obj: any = {};
//     message.entry !== undefined &&
//       (obj.entry = message.entry ? Entry.toJSON(message.entry) : undefined);
//     return obj;
//   },
// };

// const baseGetAfterRequest: object = { id: 0, maxEntries: 0, wait: false };

// export const GetAfterRequest = {
//   fromJSON(object: any): GetAfterRequest {
//     const message = { ...baseGetAfterRequest } as GetAfterRequest;
//     if (object.id !== undefined && object.id !== null) {
//       message.id = Number(object.id);
//     } else {
//       message.id = 0;
//     }
//     if (object.maxEntries !== undefined && object.maxEntries !== null) {
//       message.maxEntries = Number(object.maxEntries);
//     } else {
//       message.maxEntries = 0;
//     }
//     if (object.wait !== undefined && object.wait !== null) {
//       message.wait = Boolean(object.wait);
//     } else {
//       message.wait = false;
//     }
//     return message;
//   },

//   toJSON(message: GetAfterRequest): unknown {
//     const obj: any = {};
//     message.id !== undefined && (obj.id = message.id);
//     message.maxEntries !== undefined && (obj.maxEntries = message.maxEntries);
//     message.wait !== undefined && (obj.wait = message.wait);
//     return obj;
//   },
// };

export const GetAfterResponse = {
  fromJSON(object: any): GetAfterResponse {
    return {
      entries: asArray<GetAfterResponse_traceEntry>(GetAfterResponse_traceEntry.fromJSON, object.entries),
      lastId: asNumber(object.lastId),
      missed: asBool(object.missed),
    }
  },
};

export const GetAfterResponse_traceEntry = {
  fromJSON(object: any): GetAfterResponse_traceEntry {
    return {
      id: asNumber(object.id),
      entry: asItem<Entry>(Entry.fromJSON, object.entry, Entry.fromJSON({})),
    }
  },
};

// const baseGetPolicyRequest: object = {};

// export const GetPolicyRequest = {
//   fromJSON(_: any): GetPolicyRequest {
//     const message = { ...baseGetPolicyRequest } as GetPolicyRequest;
//     return message;
//   },

//   toJSON(_: GetPolicyRequest): unknown {
//     const obj: any = {};
//     return obj;
//   },
// };

export const GetPolicyResponse = {
  fromJSON(object: any): GetPolicyResponse {
    return {
      maxEntriesHeld: asNumber(object.maxEntriesHeld),
      firstId: asNumber(object.firstId),
    }
  },
};


// const basePingResponse: object = {};

// export const PingResponse = {
//   fromJSON(object: any): PingResponse {
//     const message = { ...basePingResponse } as PingResponse;
//     if (object.expired !== undefined && object.expired !== null) {
//       message.expired = Boolean(object.expired);
//     } else {
//       message.expired = undefined;
//     }
//     if (object.statusResponse !== undefined && object.statusResponse !== null) {
//       message.statusResponse = StatusResponse.fromJSON(object.statusResponse);
//     } else {
//       message.statusResponse = undefined;
//     }
//     return message;
//   },

//   toJSON(message: PingResponse): unknown {
//     const obj: any = {};
//     message.expired !== undefined && (obj.expired = message.expired);
//     message.statusResponse !== undefined &&
//       (obj.statusResponse = message.statusResponse
//         ? StatusResponse.toJSON(message.statusResponse)
//         : undefined);
//     return obj;
//   },
// };

export const WatchResponse = {
  fromJSON(object: any): WatchResponse {
    return {
      expired: asItem<boolean | undefined>(Boolean, object.expired, undefined),
      statusResponse: asItem<StatusResponse | undefined>(StatusResponse.fromJSON, object.statusResponse, undefined),
    }
  }
}