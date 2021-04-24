// Display a rack, with its contents

import React from "react";
import {RackDetails} from "../proxies/InventoryProxy";
import {Colors} from "./SimulatedInventory";
import {Blade} from "./Blade";
import {Tor} from "./Tor";
import {PDU} from "./PDU";
import {BladeCapacity} from "../pkg/protos/inventory/capacity";

export function Rack(props: {
    bladeLimit: number,
    capacityLimit: BladeCapacity,
    rack: RackDetails,
    palette: Colors}) {
    const bladeHeight = 20
    const yGap = 1
    const headerSpace = 50

    const rackHeight = ((bladeHeight + yGap) * props.bladeLimit) - yGap + headerSpace
    let yPos = headerSpace

    return (
        <svg
            width={160}
            height={rackHeight}>
            <Tor
                x={0}
                y={0}
                width={160}
                height={bladeHeight}
                details={props.rack.tor}
                palette={props.palette}/>

            <PDU
                x={0}
                y={bladeHeight + yGap}
                width={160}
                height={bladeHeight}
                details={props.rack.pdu}
                palette={props.palette}/>

            {Array.from(props.rack.blades).map((v) => {
                const thisY = yPos
                yPos += bladeHeight + yGap

                return <Blade
                    x={0}
                    y={thisY}
                    width={160}
                    height={bladeHeight}
                    index={v[0]}
                    details={v[1]}
                    limits={props.capacityLimit}
                    palette={props.palette} />
            })}
        </svg>
    )
}
