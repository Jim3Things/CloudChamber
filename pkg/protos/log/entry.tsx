import { asArray, asBool, asItem, asNumber, asString } from "../utils"

/* eslint-disable */
export const protobufPackage = "log";

/** Describe the type of impact that this event has on a module. */
export enum Impact {
  Invalid = 0,
  Read = 1,
  Create = 2,
  Modify = 3,
  Delete = 4,
  Execute = 5,
  UNRECOGNIZED = -1,
}

export function impactFromJSON(object: any): Impact {
  if (object === undefined || object === null) {
    return Impact.Invalid
  }

  switch (object) {
    case 0:
    case "Invalid":
      return Impact.Invalid;
    case 1:
    case "Read":
      return Impact.Read;
    case 2:
    case "Create":
      return Impact.Create;
    case 3:
    case "Modify":
      return Impact.Modify;
    case 4:
    case "Delete":
      return Impact.Delete;
    case 5:
    case "Execute":
      return Impact.Execute;
    case -1:
    case "UNRECOGNIZED":
    default:
      return Impact.UNRECOGNIZED;
  }
}

export function impactToJSON(object: Impact): string {
  switch (object) {
    case Impact.Invalid:
      return "Invalid";
    case Impact.Read:
      return "Read";
    case Impact.Create:
      return "Create";
    case Impact.Modify:
      return "Modify";
    case Impact.Delete:
      return "Delete";
    case Impact.Execute:
      return "Execute";
    default:
      return "UNKNOWN";
  }
}

/** Describe the actions to take when reading an event entry. */
export enum Action {
  /**
   * Trace - Trace is the most common type of event.  The contents are added to a serial
   * list in the span, and the formatters will display the entry's data as a
   * child trace event.
   */
  Trace = 0,
  /**
   * UpdateSpanName - UpdateSpanName and UpdateReason are directives to edit the containing span
   * information.  The first replaces the span's name field, and the second
   * replaces the span's reason text. This allows for better descriptions for a
   * span once the details have been better understood - e.g. 'logging in a user'
   * vs. 'logging in user "admin"'.
   */
  UpdateSpanName = 1,
  UpdateReason = 2,
  /**
   * SpanStart - SpanStart is used to place the child span in the correct spot in the
   * sequence of events in the containing span.  It identifies the child span's
   * ID.  Structured formatters will expand the child span at this point in the
   * sequence in order to keep a time order.  Note that parent/child span
   * relationships are strong - they can safely assume that both spans will
   * complete, they will execute in the same process, and that completion of the
   * overall trace ID is not complete until both are complete.
   */
  SpanStart = 3,
  /**
   * AddLink - AddLink is used to place a the request point that may result in a linked
   * span.  It has an associated ID that is assigned by the active span at the
   * point of the request, as the future linked span id cannot yet be known.
   * Note that linked spans have a much looser relationship than parent/child
   * spans.  The linked span may not be required to complete a logical trace
   * sequence.  It may not execute in the same process as the initiator.  It
   * may not even execute.  Consequently, structured formatters consider the
   * linked information as soft (optional) parent/child relationships.  If they
   * can put them into a logical execution tree, they do so.  If they cannot,
   * then they do not.
   */
  AddLink = 4,
  UNRECOGNIZED = -1,
}

export function actionFromJSON(object: any): Action {
  if (object === undefined || object === null) {
    return Action.Trace
  }

  switch (object) {
    case 0:
    case "Trace":
      return Action.Trace;
    case 1:
    case "UpdateSpanName":
      return Action.UpdateSpanName;
    case 2:
    case "UpdateReason":
      return Action.UpdateReason;
    case 3:
    case "SpanStart":
      return Action.SpanStart;
    case 4:
    case "AddLink":
      return Action.AddLink;
    case -1:
    case "UNRECOGNIZED":
    default:
      return Action.UNRECOGNIZED;
  }
}

export enum Severity {
  Debug = 0,
  /** Reason - This is the severity use to denote a purely explanatory entry */
  Reason = 1,
  Info = 2,
  Warning = 3,
  Error = 4,
  Fatal = 5,
  UNRECOGNIZED = -1,
}

export function severityFromJSON(object: any): Severity {
  if (object === undefined || object === null) {
    return Severity.Debug
  }

  switch (object) {
    case 0:
    case "Debug":
      return Severity.Debug;
    case 1:
    case "Reason":
      return Severity.Reason;
    case 2:
    case "Info":
      return Severity.Info;
    case 3:
    case "Warning":
      return Severity.Warning;
    case 4:
    case "Error":
      return Severity.Error;
    case 5:
    case "Fatal":
      return Severity.Fatal;
    case -1:
    case "UNRECOGNIZED":
    default:
      return Severity.UNRECOGNIZED;
  }
}

/** Describe an impacted module */
export interface Module {
  impact: Impact;
  name: string;
}

/** Define an individual trace event */
export interface Event {
  /** Simulated time when it was logged. */
  tick: number;
  /** Event severity */
  severity: Severity;
  /** Label to quickly mark the event */
  name: string;
  /** The event text itself. */
  text: string;
  /** Formatted caller's stack trace */
  stackTrace: string;
  /** The set of modules impacted, and the type of impact. */
  impacted: Module[];
  /** Action to take when this trace is encountered. */
  eventAction: Action;
  /** Child's span ID.  Ignored if the action is not SpanStart. */
  spanId: string;
  /** Outgoing link ID.  Ignored if the action is not AddLink. */
  linkId: string;
}

/** Describe a full correlated span, consisting of zero or more events. */
export interface Entry {
  /** Name of the span */
  name: string;
  /** The IDs for the span, and its parent */
  spanID: string;
  parentID: string;
  traceID: string;
  /** Final status of the span */
  status: string;
  /** Formatted stack trace */
  stackTrace: string;
  /** The set of events emitted by this span */
  event: Event[];
  /** True, if this span represents internal-only operations. */
  infrastructure: boolean;
  /**
   * Friendly string describing the purpose of the logic covered by this
   * entry.
   */
  reason: string;
  /**
   * The link tag associated with an AddLink event at the source span,
   * if present.
   */
  startingLink: string;
  /**
   * The link span ID and trace ID identify the active span at the point
   * where the request to start a new related span was made.
   */
  linkSpanID: string;
  linkTraceID: string;
}

const baseModule: object = { impact: 0, name: "" };

const nullTraceID: string = "00000000000000000000000000000000"
const missingSpanID: string = "Missing"

export const nullSpanID: string = "0000000000000000"

export const Module = {
  fromJSON(object: any): Module {
    return {
      impact: impactFromJSON(object.impact),
      name: asString(object.name),
    }
  },
};

export const Event = {
  fromJSON(object: any): Event {
    return {
      impacted: asArray<Module>(Module.fromJSON, object.impacted),
      tick: asNumber(object.tick),
      severity: severityFromJSON(object.severity),
      name: asString(object.name),
      text: asString(object.text),
      stackTrace: asString(object.stackTrace),
      eventAction: actionFromJSON(object.eventAction),
      spanId: asItem<string>(String, object.spanId, nullSpanID),
      linkId: asString(object.linkId),
    }
  },
};

export const Entry = {
  fromJSON(object: any): Entry {
    return {
      event: asArray<Event>(Event.fromJSON, object.event),
      name: asString(object.name),
      spanID: asItem<string>(String, object.spanID, missingSpanID),
      parentID: asItem<string>(String, object.parentID, nullSpanID),
      traceID: asItem<string>(String, object.traceID, nullTraceID),
      status: asString(object.status),
      stackTrace: asString(object.stackTrace),
      infrastructure: asBool(object.infrastructure),
      reason: asString(object.reason),
      startingLink: asString(object.startingLink),
      linkSpanID: asItem<string>(String, object.linkSpanID, nullSpanID),
      linkTraceID: asItem<string>(String, object.linkTraceID, nullTraceID),
    }
  },
};
