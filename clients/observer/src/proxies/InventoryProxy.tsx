// This modules contains the proxy handler for calling the REST inventory management
// service in the Cloud Chamber backend

import {getJson} from "./Session"
import {BladeCapacity} from "../pkg/protos/inventory/capacity"
import {
    External_Blade,
    External_Pdu,
    External_Rack,
    External_Tor,
    External_ZoneSummary
} from "../pkg/protos/inventory/external"


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
export interface BladeDescription {
    blade: External_Blade
    usage: InstanceDetails[]    // Details on the workload instances present
}

export interface PduDescription {
    pdu: External_Pdu
}

export interface TorDescription {
    tor: External_Tor
}

// Describe a rack
export interface RackDetails {
    uri: string,                // Address to get rack details
    detailsLoaded: boolean,     // True if the rack details have been loaded
    tors: Map<number, TorDescription>       // The Tors
    pdus: Map<number, PduDescription>       // .. the pdus
    blades: Map<number, BladeDescription>   // .. and the blades
}

// Describe a cluster
export interface ClusterDetails {
    name: string                // Descriptive name for the cluster
    location: string
    maxBladeCount: number
    maxTorCount: number
    maxPduCount: number
    maxConnectors: number
    maxCapacity: BladeCapacity
    racks: Map<string, RackDetails>   // .. and the racks that make it up
}

// Build up some fake usage, ensuring that it will fit...
function fakeUsage(avail: number): InstanceDetails[] {
    if (avail >= 8) {
        return [
            {usage: 2, state: InstanceState.running},
            {usage: 1, state: InstanceState.escrow},
            {usage: 3, state: InstanceState.running},
            {usage: 2, state: InstanceState.faulted}
        ]
    }

    if (avail >= 4) {
        return [
            {usage: 1, state: InstanceState.running},
            {usage: 1, state: InstanceState.escrow},
            {usage: 1, state: InstanceState.running},
            {usage: 1, state: InstanceState.faulted}
        ]
    }

    return [
        {usage: 1, state: InstanceState.running},
    ]
}

// Get the top level description of the target cluster
export function getCluster(): Promise<ClusterDetails> {
    const path = "/api/racks"
    const request = new Request(path, {method: "GET"})

    return getJson<any>(request)
        .then((item: any) => {
            const zone = new External_ZoneSummary(item)
            let data: ClusterDetails = {
                name: zone.name,
                location: zone.details.location,
                maxBladeCount: zone.maxBladeCount,
                maxTorCount: zone.maxTorCount,
                maxPduCount: zone.maxPduCount,
                maxConnectors: zone.maxConnectors,
                maxCapacity: zone.maxCapacity,
                racks: new Map<string, RackDetails>()
            }

            zone.racks.forEach((rack, name) => {
                data.racks.set(name, {
                    blades: new Map<number, BladeDescription>(),
                    pdus: new Map<number, PduDescription>(),
                    tors: new Map<number, TorDescription>(),
                    detailsLoaded: false,
                    uri: rack.uri
                })
            })

            return data
        })
}

// Get the detail information for a rack.
export function getRackDetails(rack: RackDetails): Promise<RackDetails> {
    const request = new Request(rack.uri, {method: "GET"})

    return getJson<any>(request)
        .then((item: any) => {
            // Processing here is similar to the processing of the
            // Rack summary data above.
            const value = new External_Rack(item)
            let newRack: RackDetails = {...rack, detailsLoaded: true}

            value.tors.forEach((tor, key) => {
                newRack.tors.set(key, {
                    tor: tor,
                })
            })

            value.pdus.forEach((pdu, key) => {
                newRack.pdus.set(key, {
                    pdu: pdu,
                })
            })

            value.fullBlades.forEach((blade, key) => {
                newRack.blades.set(key, {
                    blade: new External_Blade(blade),
                    usage: fakeUsage(blade.capacity.cores)
                })
            })

            return newRack
        })
}
