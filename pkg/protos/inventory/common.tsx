/* eslint-disable */
import { asBool, asString } from "../utils"

export const protobufPackage = "inventory";

/**
 * Defines the data structures used to describe the capacity of an inventory
 * item. Capacity is a multi-dimensional aspect of any inventory item, since
 * each item has multiple internal resources, any one or combination of which
 * can be exhausted.
 *
 * The multi-dimensionality is important for schedulers to understand, as the
 * exhaustion of one dimension can lead to a case where the other unused
 * capacity dimensions are stranded.  For instance, if all cores are consumed
 * then no free memory, disk, or other dimensions of a blade can be used.
 */

/**
 * Underlying condition of hardware items within the inventory. Allows control of the basic
 * state of the item. Can be applied to racks, blades, tors, pdus, cables (power and network).
 */
export enum Condition {
  not_in_service = 0,
  operational = 1,
  burn_in = 2,
  out_for_repair = 3,
  retiring = 4,
  retired = 5,
  UNRECOGNIZED = -1,
}

export function conditionFromJSON(object: any): Condition {
  if (object === null || object === undefined) {
    return Condition.UNRECOGNIZED
  }

  switch (object) {
    case 0:
    case "not_in_service":
      return Condition.not_in_service;
    case 1:
    case "operational":
      return Condition.operational;
    case 2:
    case "burn_in":
      return Condition.burn_in;
    case 3:
    case "out_for_repair":
      return Condition.out_for_repair;
    case 4:
    case "retiring":
      return Condition.retiring;
    case 5:
    case "retired":
      return Condition.retired;
    case -1:
    case "UNRECOGNIZED":
    default:
      return Condition.UNRECOGNIZED;
  }
}

/**
 * Underlying state of logical items within the inventory. Allows the basic state to be
 * described. Applies to zones and regions.
 */
export enum State {
  out_of_service = 0,
  in_service = 1,
  commissioning = 2,
  assumed_failed = 3,
  decommissioning = 4,
  decommissioned = 5,
  UNRECOGNIZED = -1,
}

export function stateFromJSON(object: any): State {
  if (object === null || object === undefined) {
    return State.UNRECOGNIZED
  }

  switch (object) {
    case 0:
    case "out_of_service":
      return State.out_of_service;
    case 1:
    case "in_service":
      return State.in_service;
    case 2:
    case "commissioning":
      return State.commissioning;
    case 3:
    case "assumed_failed":
      return State.assumed_failed;
    case 4:
    case "decommissioning":
      return State.decommissioning;
    case 5:
    case "decommissioned":
      return State.decommissioned;
    case -1:
    case "UNRECOGNIZED":
    default:
      return State.UNRECOGNIZED;
  }
}

// /**
//  * Describes potential targets for wiring connections between a Pdu or Tor port and a specific
//  * item of equipment.
//  */
// export interface Hardware {
//   /** The type or item or piece of equipment */
//   type: Hardware_HwType;
//   /**
//    * Defines an instance of the piece of equipment. For example there are likely to be multiple
//    * blades and the id is used to distinguish amongst them.
//    */
//   id: number;
//   /**
//    * If the item has multiple connectors, the port field can be used to indicate which connector
//    * is used for this port.
//    */
//   port: number;
// }

// /** Defines the type of hardware that can be wired up to a Pdu power port or a Tor network port. */
// export enum Hardware_HwType {
//   /** unknown - The type of hardware is not yet known */
//   unknown = 0,
//   /** pdu - This item is a PDU (Power Distribution Unit). */
//   pdu = 1,
//   /** tor - Equipment is a TOR (Top of Rack network switch) */
//   tor = 2,
//   /** blade - Equipment is a blade computer */
//   blade = 3,
//   UNRECOGNIZED = -1,
// }

// export function hardware_HwTypeFromJSON(object: any): Hardware_HwType {
//   switch (object) {
//     case 0:
//     case "unknown":
//       return Hardware_HwType.unknown;
//     case 1:
//     case "pdu":
//       return Hardware_HwType.pdu;
//     case 2:
//     case "tor":
//       return Hardware_HwType.tor;
//     case 3:
//     case "blade":
//       return Hardware_HwType.blade;
//     case -1:
//     case "UNRECOGNIZED":
//     default:
//       return Hardware_HwType.UNRECOGNIZED;
//   }
// }

// export function hardware_HwTypeToJSON(object: Hardware_HwType): string {
//   switch (object) {
//     case Hardware_HwType.unknown:
//       return "unknown";
//     case Hardware_HwType.pdu:
//       return "pdu";
//     case Hardware_HwType.tor:
//       return "tor";
//     case Hardware_HwType.blade:
//       return "blade";
//     default:
//       return "UNKNOWN";
//   }
// }

// export interface PowerPort {
//   /** Defines whether or not the port is actually connected to the associated item of equipment. */
//   wired: boolean;
//   /** Defines what the port is wired up to. */
//   item: Hardware | undefined;
// }

// export interface NetworkPort {
//   /** Defines whether or not the port is actually connected to the associated item of equipment. */
//   wired: boolean;
//   /** Defines what the port is wired up to. */
//   item: Hardware | undefined;
// }

// export interface BladeBootInfo {
//   source: BladeBootInfo_Method;
//   image: string;
//   version: string;
//   parameters: string;
// }

// export enum BladeBootInfo_Method {
//   local = 0,
//   network = 1,
//   UNRECOGNIZED = -1,
// }

// export function bladeBootInfo_MethodFromJSON(
//   object: any
// ): BladeBootInfo_Method {
//   switch (object) {
//     case 0:
//     case "local":
//       return BladeBootInfo_Method.local;
//     case 1:
//     case "network":
//       return BladeBootInfo_Method.network;
//     case -1:
//     case "UNRECOGNIZED":
//     default:
//       return BladeBootInfo_Method.UNRECOGNIZED;
//   }
// }

// export function bladeBootInfo_MethodToJSON(
//   object: BladeBootInfo_Method
// ): string {
//   switch (object) {
//     case BladeBootInfo_Method.local:
//       return "local";
//     case BladeBootInfo_Method.network:
//       return "network";
//     default:
//       return "UNKNOWN";
//   }
// }

// /** Power distribution unit.  Network accessible power controller */
// export interface PduDetails {
//   /**
//    * Whether or not the pdu is enabled. This is orthogonal to the condition of the
//    * pdu. To schedule resources within the pdu, the pdu must be both enabled and
//    * the condition must be operational.
//    */
//   enabled: boolean;
//   /**
//    * Defines the overall condition of the pdu. This is orthogonal to the enabling of
//    * the pdu. To schedule resources within the pdu, the pdu must be both enabled and
//    * the condition must be operational.
//    */
//   condition: Condition;
// }

// /** Rack-level network switch. */
// export interface TorDetails {
//   /**
//    * Whether or not the tor is enabled. This is orthogonal to the condition of the
//    * tor. To schedule resources within the tor, the tor must be both enabled and
//    * the condition must be operational.
//    */
//   enabled: boolean;
//   /**
//    * Defines the overall condition of the tor. This is orthogonal to the enabling of
//    * the tor. To schedule resources within the tor, the tor must be both enabled and
//    * the condition must be operational.
//    */
//   condition: Condition;
// }

// /** Rack-level blade computer */
// export interface BladeDetails {
//   /**
//    * Whether or not the blade is enabled. This is orthogonal to the condition of the
//    * blade. To schedule resources within the blade, the blade must be both enabled
//    * and the condition must be operational.
//    */
//   enabled: boolean;
//   /**
//    * Defines the overall condition of the blade. This is orthogonal to the enabling of
//    * the blade. To schedule resources within the blade, the blade must be both enabled
//    * and the condition must be operational.
//    */
//   condition: Condition;
// }

/**
 * This assumes a single overhead item per rack.  May want to allow multiple to handle
 * subdivisions for power or network, say.
 */
export class RackDetails {
  /**
   * Whether or not the rack as a whole is enabled. This is orthogonal to the condition
   * of the rack. To schedule resources within the rack, the rack must be both enabled
   * and the condition must be operational.
   */
  enabled: boolean;
  /**
   * Defines the overall condition of the rack. This is orthogonal to the enabling of
   * the rack. To schedule resources within the rack, the rack must be both enabled
   * and the condition must be operational.
   */
  condition: Condition;
  /**
   * Arbitrary string used to allow the physical location of the rack to be recorded in
   * a user defined format. Has no effect on the operation of the rack, for display
   * purposes only.
   */
  location: string;
  /**
   * Arbitrary string used to allow any operational notes for the blade to be recorded
   * in a user defined format. Has no effect on the operation of the rack, for display
   * purposes only.
   */
  notes: string;

  constructor(object: any) {
    this.enabled = asBool(object.enabled)
    this.condition = conditionFromJSON(object.condition)
    this.location = asString(object.location)
    this.notes = asString(object.notes)
  }
}

export class ZoneDetails {
  /**
   * Whether or not the zone as a whole is enabled. This is orthogonal to the condition
   * of the zone. To schedule resources within the zone, the zone must be both enabled
   * and the condition must be operational.
   */
  enabled: boolean;
  /**
   * Defines the overall condition of the zone. This is orthogonal to the enabling of
   * the zone. To schedule resources within the zone, the zone must be both enabled
   * and the condition must be operational.
   */
  state: State;
  /**
   * Arbitrary string used to allow the physical location of the zone to be recorded in
   * a user defined format. Has no effect on the operation of the zone, for display
   * purposes only.
   */
  location: string;
  /**
   * Arbitrary string used to allow any operational notes for the zone to be recorded
   * in a user defined format. Has no effect on the operation of the zone, for display
   * purposes only.
   */
  notes: string;

  constructor(object: any) {
    this.enabled = asBool(object.enabled)
    this.state = stateFromJSON(object.state)
    this.location = asString(object.location)
    this.notes = asString(object.notes)
  }
}

// export interface RegionDetails {
//   /**
//    * The name of a region.
//    *
//    * NOTE: Not sure we need an explicit name field since the name of the record is
//    * implicit in identifying the record.
//    */
//   name: string;
//   /**
//    * Defines the overall condition of the region. To schedule resources within the
//    * region, the region's condition must be operational.
//    */
//   state: State;
//   /**
//    * Arbitrary string used to allow the physical location of the zone to be recorded in
//    * a user defined format. Has no effect on the operation of the zone, for display
//    * purposes only.
//    */
//   location: string;
//   /**
//    * Arbitrary string used to allow any operational notes for the zone to be recorded
//    * in a user defined format. Has no effect on the operation of the zone, for display
//    * purposes only.
//    */
//   notes: string;
// }

// export interface RootDetails {
//   *
//    * The name of the root of the configuration / simulation.
//    *
//    * NOTE: Not sure we need an explicit name field since the name of the record is
//    * implicit in identifying the record.
   
//   name: string;
//   /**
//    * Arbitrary string used to allow any operational notes for the zone to be recorded
//    * in a user defined format. Has no effect on the operation of the zone, for display
//    * purposes only.
//    */
//   notes: string;
// }

// const baseHardware: object = { type: 0, id: 0, port: 0 };

// export const Hardware = {
//   fromJSON(object: any): Hardware {
//     const message = { ...baseHardware } as Hardware;
//     if (object.type !== undefined && object.type !== null) {
//       message.type = hardware_HwTypeFromJSON(object.type);
//     } else {
//       message.type = 0;
//     }
//     if (object.id !== undefined && object.id !== null) {
//       message.id = Number(object.id);
//     } else {
//       message.id = 0;
//     }
//     if (object.port !== undefined && object.port !== null) {
//       message.port = Number(object.port);
//     } else {
//       message.port = 0;
//     }
//     return message;
//   },

//   toJSON(message: Hardware): unknown {
//     const obj: any = {};
//     message.type !== undefined &&
//       (obj.type = hardware_HwTypeToJSON(message.type));
//     message.id !== undefined && (obj.id = message.id);
//     message.port !== undefined && (obj.port = message.port);
//     return obj;
//   },
// };

// const basePowerPort: object = { wired: false };

// export const PowerPort = {
//   fromJSON(object: any): PowerPort {
//     const message = { ...basePowerPort } as PowerPort;
//     if (object.wired !== undefined && object.wired !== null) {
//       message.wired = Boolean(object.wired);
//     } else {
//       message.wired = false;
//     }
//     if (object.item !== undefined && object.item !== null) {
//       message.item = Hardware.fromJSON(object.item);
//     } else {
//       message.item = undefined;
//     }
//     return message;
//   },

//   toJSON(message: PowerPort): unknown {
//     const obj: any = {};
//     message.wired !== undefined && (obj.wired = message.wired);
//     message.item !== undefined &&
//       (obj.item = message.item ? Hardware.toJSON(message.item) : undefined);
//     return obj;
//   },
// };

// const baseNetworkPort: object = { wired: false };

// export const NetworkPort = {
//   fromJSON(object: any): NetworkPort {
//     const message = { ...baseNetworkPort } as NetworkPort;
//     if (object.wired !== undefined && object.wired !== null) {
//       message.wired = Boolean(object.wired);
//     } else {
//       message.wired = false;
//     }
//     if (object.item !== undefined && object.item !== null) {
//       message.item = Hardware.fromJSON(object.item);
//     } else {
//       message.item = undefined;
//     }
//     return message;
//   },

//   toJSON(message: NetworkPort): unknown {
//     const obj: any = {};
//     message.wired !== undefined && (obj.wired = message.wired);
//     message.item !== undefined &&
//       (obj.item = message.item ? Hardware.toJSON(message.item) : undefined);
//     return obj;
//   },
// };

// const baseBladeBootInfo: object = {
//   source: 0,
//   image: "",
//   version: "",
//   parameters: "",
// };

// export const BladeBootInfo = {
//   fromJSON(object: any): BladeBootInfo {
//     const message = { ...baseBladeBootInfo } as BladeBootInfo;
//     if (object.source !== undefined && object.source !== null) {
//       message.source = bladeBootInfo_MethodFromJSON(object.source);
//     } else {
//       message.source = 0;
//     }
//     if (object.image !== undefined && object.image !== null) {
//       message.image = String(object.image);
//     } else {
//       message.image = "";
//     }
//     if (object.version !== undefined && object.version !== null) {
//       message.version = String(object.version);
//     } else {
//       message.version = "";
//     }
//     if (object.parameters !== undefined && object.parameters !== null) {
//       message.parameters = String(object.parameters);
//     } else {
//       message.parameters = "";
//     }
//     return message;
//   },

//   toJSON(message: BladeBootInfo): unknown {
//     const obj: any = {};
//     message.source !== undefined &&
//       (obj.source = bladeBootInfo_MethodToJSON(message.source));
//     message.image !== undefined && (obj.image = message.image);
//     message.version !== undefined && (obj.version = message.version);
//     message.parameters !== undefined && (obj.parameters = message.parameters);
//     return obj;
//   },
// };

// const basePduDetails: object = { enabled: false, condition: 0 };

// export const PduDetails = {
//   fromJSON(object: any): PduDetails {
//     const message = { ...basePduDetails } as PduDetails;
//     if (object.enabled !== undefined && object.enabled !== null) {
//       message.enabled = Boolean(object.enabled);
//     } else {
//       message.enabled = false;
//     }
//     if (object.condition !== undefined && object.condition !== null) {
//       message.condition = conditionFromJSON(object.condition);
//     } else {
//       message.condition = 0;
//     }
//     return message;
//   },

//   toJSON(message: PduDetails): unknown {
//     const obj: any = {};
//     message.enabled !== undefined && (obj.enabled = message.enabled);
//     message.condition !== undefined &&
//       (obj.condition = conditionToJSON(message.condition));
//     return obj;
//   },
// };

// const baseTorDetails: object = { enabled: false, condition: 0 };

// export const TorDetails = {
//   fromJSON(object: any): TorDetails {
//     const message = { ...baseTorDetails } as TorDetails;
//     if (object.enabled !== undefined && object.enabled !== null) {
//       message.enabled = Boolean(object.enabled);
//     } else {
//       message.enabled = false;
//     }
//     if (object.condition !== undefined && object.condition !== null) {
//       message.condition = conditionFromJSON(object.condition);
//     } else {
//       message.condition = 0;
//     }
//     return message;
//   },

//   toJSON(message: TorDetails): unknown {
//     const obj: any = {};
//     message.enabled !== undefined && (obj.enabled = message.enabled);
//     message.condition !== undefined &&
//       (obj.condition = conditionToJSON(message.condition));
//     return obj;
//   },
// };

// const baseBladeDetails: object = { enabled: false, condition: 0 };

// export const BladeDetails = {
//   fromJSON(object: any): BladeDetails {
//     const message = { ...baseBladeDetails } as BladeDetails;
//     if (object.enabled !== undefined && object.enabled !== null) {
//       message.enabled = Boolean(object.enabled);
//     } else {
//       message.enabled = false;
//     }
//     if (object.condition !== undefined && object.condition !== null) {
//       message.condition = conditionFromJSON(object.condition);
//     } else {
//       message.condition = 0;
//     }
//     return message;
//   },

//   toJSON(message: BladeDetails): unknown {
//     const obj: any = {};
//     message.enabled !== undefined && (obj.enabled = message.enabled);
//     message.condition !== undefined &&
//       (obj.condition = conditionToJSON(message.condition));
//     return obj;
//   },
// };

// const baseRegionDetails: object = {
//   name: "",
//   state: 0,
//   location: "",
//   notes: "",
// };

// export const RegionDetails = {
//   fromJSON(object: any): RegionDetails {
//     const message = { ...baseRegionDetails } as RegionDetails;
//     if (object.name !== undefined && object.name !== null) {
//       message.name = String(object.name);
//     } else {
//       message.name = "";
//     }
//     if (object.state !== undefined && object.state !== null) {
//       message.state = stateFromJSON(object.state);
//     } else {
//       message.state = 0;
//     }
//     if (object.location !== undefined && object.location !== null) {
//       message.location = String(object.location);
//     } else {
//       message.location = "";
//     }
//     if (object.notes !== undefined && object.notes !== null) {
//       message.notes = String(object.notes);
//     } else {
//       message.notes = "";
//     }
//     return message;
//   },

//   toJSON(message: RegionDetails): unknown {
//     const obj: any = {};
//     message.name !== undefined && (obj.name = message.name);
//     message.state !== undefined && (obj.state = stateToJSON(message.state));
//     message.location !== undefined && (obj.location = message.location);
//     message.notes !== undefined && (obj.notes = message.notes);
//     return obj;
//   },
// };

// const baseRootDetails: object = { name: "", notes: "" };

// export const RootDetails = {
//   fromJSON(object: any): RootDetails {
//     const message = { ...baseRootDetails } as RootDetails;
//     if (object.name !== undefined && object.name !== null) {
//       message.name = String(object.name);
//     } else {
//       message.name = "";
//     }
//     if (object.notes !== undefined && object.notes !== null) {
//       message.notes = String(object.notes);
//     } else {
//       message.notes = "";
//     }
//     return message;
//   },

//   toJSON(message: RootDetails): unknown {
//     const obj: any = {};
//     message.name !== undefined && (obj.name = message.name);
//     message.notes !== undefined && (obj.notes = message.notes);
//     return obj;
//   },
// };