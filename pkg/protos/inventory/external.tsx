/* eslint-disable */
import { 
  BladeBootInfo, BladeDetails, 
  BladeState_SM, CableState_SM, PduState_SM, TorState_SM,
  NetworkPort, PduDetails, 
  PowerPort, RackDetails, TorDetails, ZoneDetails,
  bladeState_SMFromJSON, cableState_SMFromJSON, pduState_SMFromJSON, torState_SMFromJSON
} from "./common";

import { BladeCapacity } from "./capacity";
import { asBool, asMap, asNumber, asString } from "../utils"

export const protobufPackage = "inventory";

// export interface External {}

export class External_ObservedCable {
  /** The simulated time when the observation was made */
  at: number;
  /** The state the PDU was in at that time. */
  smState: CableState_SM;
  /** The simulated time when it entered this state. */
  enteredAt: number;

  constructor(object: any) {
    if (object === null || object === undefined) {
      this.at = 0
      this.enteredAt = 0
      this.smState = CableState_SM.UNRECOGNIZED
      return

    }

    this.at = asNumber(object.at)
    this.smState = cableState_SMFromJSON(object.smState)
    this.enteredAt = asNumber(object.enteredAt)
  }
}

export class External_Pdu_ObservedState {
  /** The simulated time when the observation was made */
  at: number;
  /** The state the PDU was in at that time. */
  smState: PduState_SM;
  /** The simulated time when it entered this state. */
  enteredAt: number;

  constructor(object: any) {
    if (object === null || object === undefined) {
      this.at = 0
      this.enteredAt = 0
      this.smState = PduState_SM.UNRECOGNIZED
      return
    }

    this.at = asNumber(object.at)
    this.smState = pduState_SMFromJSON(object.smState)
    this.enteredAt = asNumber(object.enteredAt)
  }
}

export class External_Pdu_Port {
  port: PowerPort;
  observed: External_ObservedCable;

  constructor(object: any) {
    this.port = new PowerPort(object.port)
    this.observed = new External_ObservedCable(object.observed)
  }
}

/** Power distribution unit.  Network accessible power controller */
export class External_Pdu {
  details: PduDetails;
  /**
   * Defines a power "socket" which is used to provide power to a blade. There is
   * a 1 to 1 mapping of a power port to a blade within a single rack and it is an
   * error if there fewer power ports than blades.
   */
  ports: Map<number, External_Pdu_Port>

  observed: External_Pdu_ObservedState

  constructor(object: any) {
    if (object === null || object === undefined) {
      this.details = new PduDetails(undefined)
      this.ports = new Map<number, External_Pdu_Port>()
      this.observed = new External_Pdu_ObservedState(null)
      return
    }

    this.details = new PduDetails(object.details)
    this.ports = asMap(object.ports, (k, v) => [Number(k), new External_Pdu_Port(v)])
    this.observed = new External_Pdu_ObservedState(object.observed)
  }
}

export class External_Tor_ObservedState {
  /** The simulated time when the observation was made */
  at: number;
  /** The state the TOR was in at that time. */
  smState: TorState_SM;
  /** The simulated time when it entered this state. */
  enteredAt: number;

  constructor(object: any) {
    if (object === null || object === undefined) {
      this.at = 0
      this.enteredAt = 0
      this.smState = TorState_SM.UNRECOGNIZED
      return
    }
    
    this.at = asNumber(object.at)
    this.smState = torState_SMFromJSON(object.smState)
    this.enteredAt = asNumber(object.enteredAt)
  }
}

export class External_Tor_Port {
  port: NetworkPort;
  observed: External_ObservedCable;

  constructor(object: any) {
    this.port = new NetworkPort(object.port)
    this.observed = new External_ObservedCable(object.observed)
  }
}

/** Rack-level network switch. */
export class External_Tor {
  details: TorDetails
  /**
   * Defines a network "port" which is used to provide a network connection to a
   * blade. There is a 1 to 1 mapping of a network port to a blade within a single
   * rack and it is an error if there fewer network ports than blades.
   */
  ports: Map<number, External_Tor_Port>

  observed: External_Tor_ObservedState

  constructor(object: any) {
    if (object === null || object === undefined) {
      this.details = new TorDetails(undefined)
      this.ports = new Map<number, External_Tor_Port>()
      this.observed = new External_Tor_ObservedState(null)
      return
    }

    this.details = new TorDetails(object.details)
    this.ports = asMap(object.ports, (k, v) => [Number(k), new External_Tor_Port(v)])
    this.observed = new External_Tor_ObservedState(object.observed)
  }
}

/** Individual blade within the rack */
export class External_Blade {
  details: BladeDetails
  capacity: BladeCapacity
  /**
   * Defines whether or not the blade automatically begins a boot sequence when power is
   * applied to the blade.
   */
  bootOnPowerOn: boolean
  /** Describes the default boot mechanism */
  bootInfo: BladeBootInfo

  observed: External_Blade_ObservedState

  constructor(object: any) {
    this.details = new BladeDetails(object.details)
    this.capacity = new BladeCapacity(object.capacity)
    this.bootOnPowerOn = asBool(object.bootOnPowerOn)
    this.bootInfo = new BladeBootInfo(object.bootInfo)
    this.observed = new External_Blade_ObservedState(object.observed)
  }
}

/** Observed, actual, and target data follows on from here... */
export class External_Blade_ObservedState {
  /** The simulated time when the observation was made */
  at: number
  /** The state the blade was in at that time. */
  smState: BladeState_SM
  /** The simulated time when it entered this state. */
  enteredAt: number

  constructor(object: any) {
    this.at = asNumber(object.at)
    this.smState = bladeState_SMFromJSON(object.smState)
    this.enteredAt = asNumber(object.enteredAt)
  }
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

  pdus: Map<number, External_Pdu>
  tors: Map<number, External_Tor>
  fullBlades: Map<number, External_Blade>

  constructor(object: any) {
    this.details = new RackDetails(object.details)

    this.pdu = new External_Pdu(object.pdu)
    this.tor = new External_Tor(object.tor)

    this.blades = asMap(object.blades, (k, v) => [Number(k), new BladeCapacity(v)])
    this.pdus = asMap(object.pdus, (k, v) => [Number(k), new External_Pdu(v)])
    this.tors = asMap(object.tors, (k, v) => [Number(k), new External_Tor(v)])
    this.fullBlades = asMap(object.fullBlades, (k, v) => [Number(k), new External_Blade(v)])
  }
}

/**
 * Finally, a zone is a collection of racks.  Each rack has a name, which is used as a key in
 * the map below.
 */
export class External_Zone {
  racks: Map<string, External_Rack>

  constructor(object: any) {
    this.racks = asMap(object.racks, (k, v) => [asString(k), new External_Rack(v)])
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
  /** Summary information about all known racks */
  racks: Map<string, External_RackSummary>
  /** The largest number of blades held in any rack */
  maxBladeCount: number;
  /** The largest capacity values found in any blade */
  maxCapacity: BladeCapacity;
  /** Overview detail */
  details: ZoneDetails
  /** Friendly name for the zone */
  name: string;
  /** The largest number of TORs held in any rack. */
  maxTorCount: number;
  /** The largest number of PDUs held in any rack. */
  maxPduCount: number;
  /** The largest number of connectors in either a TOR or a PDU. */
  maxConnectors: number;


  constructor(object: any) {
    this.name = asString(object.name)
    this.maxBladeCount = asNumber(object.maxBladeCount)
    this.maxCapacity = new BladeCapacity(object.maxCapacity)
    this.details = new ZoneDetails(object.details)
    this.racks = asMap(object.racks, (k, v) => [asString(k), new External_RackSummary(v)])
    this.maxTorCount = asNumber(object.maxTorCount)
    this.maxPduCount = asNumber(object.maxPduCount)
    this.maxConnectors = asNumber(object.maxConnectors)
  }
}
