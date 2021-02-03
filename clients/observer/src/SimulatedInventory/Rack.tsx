// Display a rack, with its contents

import React, {FunctionComponent} from "react";
import {JsonBladeCapacity, RackDetails} from "../proxies/InventoryProxy";
import {Colors} from "./SimulatedInventory";
import {Blade} from "./Blade";
import {Tor} from "./Tor";
import {PDU} from "./PDU";

export const Rack : FunctionComponent<{
    bladeLimit: number,
    capacityLimit: JsonBladeCapacity,
    rack: RackDetails,
    palette: Colors}> = ({bladeLimit, capacityLimit, rack, palette}) => {
    const bladeHeight = 20
    const yGap = 1
    const headerSpace = 50

    const rackHeight = ((bladeHeight + yGap) * bladeLimit) - yGap + headerSpace
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
                details={rack.tor}
                palette={palette}/>

            <PDU
                x={0}
                y={bladeHeight + yGap}
                width={160}
                height={bladeHeight}
                details={rack.pdu}
                palette={palette}/>

            {Array.from(rack.blades).map((v) => {
                const thisY = yPos
                yPos += bladeHeight + yGap

                return <Blade
                    x={0}
                    y={thisY}
                    width={160}
                    height={bladeHeight}
                    index={v[0]}
                    details={v[1]}
                    limits={capacityLimit}
                    palette={palette} />
            })}
        </svg>
    )
}