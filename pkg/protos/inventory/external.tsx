/* eslint-disable */
import { RackDetails, ZoneDetails } from "./common";
import { BladeCapacity } from "./capacity";
import { asNumber, asString } from "../utils"

export const protobufPackage = "inventory";

// export interface External {}

/** Power distribution unit.  Network accessible power controller */
export class External_Pdu {
  constructor(object: any) {}
}

/** Rack-level network switch. */
export class External_Tor {
  constructor(object: any) {}
}

export class External_Rack {
  details: RackDetails
  pdu: External_Pdu
  tor: External_Tor
  /**
   * specify the blades in the rack.  Each blade is defined by an integer index within that
   * rack, which is used here as the key.
   */
  blades: Map<number, BladeCapacity>

  constructor(object: any) {
    this.details = new RackDetails(object.details)
    this.pdu = new External_Pdu(object.pdu)
    this.tor = new External_Tor(object.tor)
    this.blades = new Map<number, BladeCapacity>()

    if (object.blades !== undefined && object.blades !== null) {
      Object.entries(object.blades).forEach(([key, value]) => {
        this.blades.set(Number(key), new BladeCapacity(value))
      });
    }
  }
}

/**
 * Finally, a zone is a collection of racks.  Each rack has a name, which is used as a key in
 * the map below.
 */
export class External_Zone {
  racks: Map<string, External_Rack>

  constructor(object: any) {
    this.racks = new Map<string, External_Rack>()

    if (object.racks !== undefined && object.racks !== null) {
      Object.entries(object.racks).forEach(([key, value]) => {
        this.racks.set(key, new External_Rack(value))
      });
    }
  }
}

/** Rack list entry item */
export class External_RackSummary {
  /** host relative URI that can be used to retrieve its details */
  uri: string;

  constructor(object: any) {
    this.uri = asString(object.uri)
  }
}

/** Summary of the full inventory */
export class External_ZoneSummary {
  name: string

  /** Summary information about all known racks */
  racks: Map<string, External_RackSummary>
  /** The largest number of blades held in any rack */
  maxBladeCount: number;
  /** The largest capacity values found in any blade */
  maxCapacity: BladeCapacity;

  details: ZoneDetails

  constructor(object: any) {
    this.name = asString(object.name)
    this.maxBladeCount = asNumber(object.maxBladeCount)
    this.maxCapacity = new BladeCapacity(object.maxCapacity)
    this.details = new ZoneDetails(object.details)
    this.racks = new Map<string, External_RackSummary>()

    if (object.racks !== undefined && object.racks !== null) {
      Object.entries(object.racks).forEach(([key, value]) => {
        this.racks.set(key, new External_RackSummary(value))
      });
    }
  }
}
