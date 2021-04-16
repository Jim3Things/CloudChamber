/* eslint-disable */
import { asArray, asBool, asNumber, asString, asItem, Duration, durationFromJson } from "../utils"

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

/** Define the set of known accelerators, such as GPUs or FPGAs. */
export interface Accelerator {
  v100: Accelerator_NVIDIAV100 | undefined;
}

export interface Accelerator_NVIDIAV100 {}

/** Defines the capacity dimensions and values for a blade */
export interface BladeCapacity {
  /** The number of cores on the blade. */
  cores: number;
  /** The amount of memory, in megabytes */
  memoryInMb: number;
  /**
   * The amount of local disk space, in gigabytes.  Note that this assumes either one disk,
   * or that the disks are mounted collectively as a single volume
   */
  diskInGb: number;
  /** The network bandwidth from the host in megabits per second */
  networkBandwidthInMbps: number;
  /** The processor architecture */
  arch: string;
  /** Supply the set of accelerators for this blade, including none. */
  accelerators: Accelerator[];
}

// export interface InstanceRequirements {
//   /** The number of (potentially fractional) cores used by the instance. */
//   cores: number;
//   /** The amount of memory, in megabytes */
//   memoryInMb: number;
//   /** The network bandwidth required in megabits per second */
//   networkBandwidthInMbps: number;
//   /** The processor architecture */
//   arch: string;
//   /** Supply the set of accelerators required by this instance, including none. */
//   accelerators: Accelerator[];
// }

export const Accelerator = {
  fromJSON(object: any): Accelerator {
    return {
      v100: asItem<Accelerator_NVIDIAV100|undefined>(Accelerator_NVIDIAV100.fromJSON, object.v100, undefined),
    }
  },
}

export const Accelerator_NVIDIAV100 = {
  fromJSON(_: any): Accelerator_NVIDIAV100 {
    return {  } as Accelerator_NVIDIAV100
  },
};

export const BladeCapacity = {
  fromJSON(object: any): BladeCapacity {
    return {
      cores: asNumber(object.cores),
      memoryInMb: asNumber(object.memoryInMb),
      diskInGb: asNumber(object.diskInGb),
      networkBandwidthInMbps: asNumber(object.networkBandwidthInMbps),
      arch: asString(object.arch),
      accelerators: asArray(Accelerator.fromJSON, object.accelerators),
    }
  }
}

// const baseInstanceRequirements: object = {
//   cores: 0,
//   memoryInMb: 0,
//   networkBandwidthInMbps: 0,
//   arch: "",
// };

// export const InstanceRequirements = {
//   fromJSON(object: any): InstanceRequirements {
//     const message = { ...baseInstanceRequirements } as InstanceRequirements;
//     message.accelerators = [];
//     if (object.cores !== undefined && object.cores !== null) {
//       message.cores = Number(object.cores);
//     } else {
//       message.cores = 0;
//     }
//     if (object.memoryInMb !== undefined && object.memoryInMb !== null) {
//       message.memoryInMb = Number(object.memoryInMb);
//     } else {
//       message.memoryInMb = 0;
//     }
//     if (
//       object.networkBandwidthInMbps !== undefined &&
//       object.networkBandwidthInMbps !== null
//     ) {
//       message.networkBandwidthInMbps = Number(object.networkBandwidthInMbps);
//     } else {
//       message.networkBandwidthInMbps = 0;
//     }
//     if (object.arch !== undefined && object.arch !== null) {
//       message.arch = String(object.arch);
//     } else {
//       message.arch = "";
//     }
//     if (object.accelerators !== undefined && object.accelerators !== null) {
//       for (const e of object.accelerators) {
//         message.accelerators.push(Accelerator.fromJSON(e));
//       }
//     }
//     return message;
//   },

//   toJSON(message: InstanceRequirements): unknown {
//     const obj: any = {};
//     message.cores !== undefined && (obj.cores = message.cores);
//     message.memoryInMb !== undefined && (obj.memoryInMb = message.memoryInMb);
//     message.networkBandwidthInMbps !== undefined &&
//       (obj.networkBandwidthInMbps = message.networkBandwidthInMbps);
//     message.arch !== undefined && (obj.arch = message.arch);
//     if (message.accelerators) {
//       obj.accelerators = message.accelerators.map((e) =>
//         e ? Accelerator.toJSON(e) : undefined
//       );
//     } else {
//       obj.accelerators = [];
//     }
//     return obj;
//   },
// };
