// This modules contains the proxy handler for calling the REST inventory management
// service in the Cloud Chamber backend

import {failIfError} from "./Session";

// Define the inventory schema as supplied by the REST service

export interface JsonRackSummary {
    uri: string
}

export interface JsonBladeCapacity {
    cores: number,
    memoryInMb: number,
    diskInGb: number,
    networkBandwidthInMbps: number
}

export interface JsonZoneSummary {
    racks: any,
    maxBladeCount: number,
    maxCapacity: JsonBladeCapacity
}

export interface JsonTor {

}

export interface  JsonPDU {

}

export interface JsonRack {
    tor: JsonTor,
    pdu: JsonPDU,
    blades: any
}

// Denote the running states of a workload instance
export enum InstanceState {
    escrow,         // Space allocated, but instance not yet running
    running,        // Instance is running
    faulted         // Instance is present, but failed
}

// Denote the state of the physical component
export enum PhysicalState {
    off,            // Powered off
    healthy,        // Powered on, working
    faulted         // Not working
}

// Describe a workload instance
export interface InstanceDetails {
    usage: number       // Amount of blade capacity used
    state: InstanceState    // Running state of the instance
}

// Describe a blade
export interface BladeDetails {
    capacity: JsonBladeCapacity // total capacity present in the blade
    state: PhysicalState        // The physical blade's health state
    usage: InstanceDetails[]    // Details on the workload instances present
}

// Describe a TOR switch
export interface TorDetails {
    state: PhysicalState        // Health state of the TOR
    linkTo: boolean[]           // Connections to the blade
    // TODO: How to represent SDN connections to workload instances?
}

// Describe a power distribution controller unit
export interface PduDetails {
    state: PhysicalState        // Health state of the PDU
    powerTo: boolean[]          // Power switch to each blade
}

// Describe a rack
export interface RackDetails {
    uri: string,                // Address to get rack details
    detailsLoaded: boolean,     // True if the rack details have been loaded
    tor: TorDetails             // The Tor
    pdu: PduDetails             // .. the pdu
    blades: Map<number, BladeDetails> // .. and the blades
}

// Describe a cluster
export interface ClusterDetails {
    name: string                // Descriptive name for the cluster
    maxBladeCount: number,
    maxCapacity: JsonBladeCapacity
    racks: Map<string, RackDetails>   // .. and the racks that make it up
}

export class InventoryProxy {
    // Get the top level description of the target cluster
    public getCluster(): Promise<ClusterDetails> {
        const path = "/api/racks"
        const request = new Request(path, { method: "GET" })

        return fetch(request)
            .then((resp: Response) => {
                failIfError(request, resp)

                return resp.json() as Promise<JsonZoneSummary>
            })
            .then((zone: JsonZoneSummary) => {
                let data : ClusterDetails = {
                    name: "My Test Cluster",        // Temporary name
                    maxBladeCount: zone.maxBladeCount,
                    maxCapacity: zone.maxCapacity,
                    racks: new Map<string, RackDetails>()
                }

                // Getting the rack summary information is a bit harder.  The
                // Json has it in a form that turns the racks collection into
                // a typescript object with fields that are named based on the
                // keys in the map.
                //
                // So we use reflection to get each entry in the map and then
                // use the '...xxx' notation to move the value into a properly
                // typed entry, and then put that into racks Map (along with
                // some temporary state)
                for (const name of Object.getOwnPropertyNames(zone.racks)) {
                    const obj = zone.racks[name]
                    const rack : JsonRackSummary = {...obj}
                    data.racks.set(name, {
                        blades: new Map<number, BladeDetails>(),
                        pdu: {
                            state: PhysicalState.healthy,
                            powerTo: [],
                        },
                        tor: {
                            state: PhysicalState.healthy,
                            linkTo: []
                        },
                        detailsLoaded: false,
                        uri: rack.uri
                    })
                }

                return data
            })
    }

    // Build up some fake usage, ensuring that it will fit...
    private static fakeUsage(avail: number): InstanceDetails[] {
        if (avail >= 8) {
            return [
                { usage: 2, state: InstanceState.running },
                { usage: 1, state: InstanceState.escrow },
                { usage: 3, state: InstanceState.running },
                { usage: 2, state: InstanceState.faulted}
            ]
        }

        if (avail >= 4) {
            return [
                { usage: 1, state: InstanceState.running },
                { usage: 1, state: InstanceState.escrow },
                { usage: 1, state: InstanceState.running },
                { usage: 1, state: InstanceState.faulted}
            ]
        }

        return [
            { usage: 1, state: InstanceState.running },
        ]
    }

    // Get the detail information for a rack.
    public getRackDetails(rack: RackDetails): Promise<RackDetails> {
        const request = new Request(rack.uri, { method: "GET" })

        return fetch(request)
            .then((resp: Response) => {
                failIfError(request, resp)

                return resp.json() as Promise<JsonRack>
            })
            .then((value: JsonRack) => {
                // Processing here is similar to the processing of the
                // Rack summary data above.
                let newRack: RackDetails = {...rack, detailsLoaded: true }
                for (const name of Object.getOwnPropertyNames(value.blades)) {
                    const blade: JsonBladeCapacity = {...value.blades[name]}
                    newRack.blades.set(+name, {
                        capacity: { ...blade },
                        state: PhysicalState.healthy,
                        usage: InventoryProxy.fakeUsage(blade.cores)
                    })
                }
                return newRack
            })
    }
}