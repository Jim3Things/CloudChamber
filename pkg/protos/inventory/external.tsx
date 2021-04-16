/* eslint-disable */
import { RackDetails, ZoneDetails } from "./common";
import { BladeCapacity } from "./capacity";
import { asArray, asBool, asNumber, asItem, asString, Duration, durationFromJson } from "../utils"

export const protobufPackage = "inventory";

// export interface External {}

/** Power distribution unit.  Network accessible power controller */
export interface External_Pdu {}

/** Rack-level network switch. */
export interface External_Tor {}

export interface External_Rack {
  details: RackDetails
  pdu: External_Pdu
  tor: External_Tor
  /**
   * specify the blades in the rack.  Each blade is defined by an integer index within that
   * rack, which is used here as the key.
   */
  blades: { [key: number]: BladeCapacity }
}

/**
 * Finally, a zone is a collection of racks.  Each rack has a name, which is used as a key in
 * the map below.
 */
export interface External_Zone {
  racks: { [key: string]: External_Rack };
}

/** Rack list entry item */
export interface External_RackSummary {
  /** host relative URI that can be used to retrieve its details */
  uri: string;
}

/** Summary of the full inventory */
export interface External_ZoneSummary {
  name: string

  /** Summary information about all known racks */
  racks: { [key: string]: External_RackSummary };
  /** The largest number of blades held in any rack */
  maxBladeCount: number;
  /** The largest capacity values found in any blade */
  maxCapacity: BladeCapacity;

  details: ZoneDetails
}

export const External_Pdu = {
  fromJSON(_: any): External_Pdu {
    return { } as External_Pdu
  },
};

export const External_Tor = {
  fromJSON(_: any): External_Tor {
    return {  } as External_Tor
  },
}

export const External_Rack = {
  fromJSON(object: any): External_Rack {
    const rack: External_Rack = {
      details: RackDetails.fromJSON(object.details),
      pdu: External_Pdu.fromJSON(object.pdu),
      tor: External_Tor.fromJSON(object.tor),
      blades: {}
    }

    if (object.blades !== undefined && object.blades !== null) {
      Object.entries(object.blades).forEach(([key, value]) => {
        rack.blades[Number(key)] = BladeCapacity.fromJSON(value);
      });
    }

    return rack
  },
};

export const External_Zone = {
  fromJSON(object: any): External_Zone {
    const zone: External_Zone = {
      racks: {}
    }

    if (object.racks !== undefined && object.racks !== null) {
      Object.entries(object.racks).forEach(([key, value]) => {
        zone.racks[key] = External_Rack.fromJSON(value);
      });
    }

    return zone
  },
}

export const External_RackSummary = {
  fromJSON(object: any): External_RackSummary {
    return {
      uri: asString(object.uri),
    }
  }
}

export const External_ZoneSummary = {
  fromJSON(object: any): External_ZoneSummary {
    const summary: External_ZoneSummary = {
      name: asString(object.name),
      maxBladeCount: asNumber(object.maxBladeCount),
      maxCapacity: BladeCapacity.fromJSON(object.maxCapacity),
      details: ZoneDetails.fromJSON(object.details),
      racks: {}
    }

    if (object.racks !== undefined && object.racks !== null) {
      Object.entries(object.racks).forEach(([key, value]) => {
        summary.racks[key] = External_RackSummary.fromJSON(value);
      });
    }

    return summary
  }
}
