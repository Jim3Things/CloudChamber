// Display a rack, with its contents

import React from "react"
import {RackDetails} from "../proxies/InventoryProxy"
import {Colors} from "./SimulatedInventory"
import {Blade} from "./Blade"
import {Tor} from "./Tor"
import {PDU} from "./PDU"
import {BladeCapacity} from "../pkg/protos/inventory/capacity"

export function Rack(props: {
    bladeLimit: number,
    capacityLimit: BladeCapacity,
    rack: RackDetails,
    palette: Colors
}) {
    const bladeHeight = 20
    const yGap = 1
    const headerSpace = 50

    const innerLeft = 5
    const innerTop = 5

    const innerHeight = ((bladeHeight + yGap) * props.bladeLimit) - yGap + headerSpace
    const innerWidth = 150

    const fullHeight = innerHeight + innerTop + innerTop
    const fullWidth = innerWidth + innerLeft + innerLeft

    let yPos = headerSpace + 5

    return (
        <svg
            width={fullWidth}
            height={fullHeight}>
            <rect
                x={0}
                y={0}
                width={fullWidth}
                height={fullHeight}
                fill="lightgrey"
                strokeWidth="5px"
                stroke="SteelBlue"
            />

            <Tor
                x={innerLeft}
                y={innerTop}
                width={innerWidth}
                height={bladeHeight}
                details={props.rack.tor}
                palette={props.palette}/>

            <PDU
                x={innerLeft}
                y={bladeHeight + yGap + innerTop}
                width={innerWidth}
                height={bladeHeight}
                details={props.rack.pdu}
                palette={props.palette}/>

            {Array.from(props.rack.blades).map((v) => {
                const thisY = yPos
                yPos += bladeHeight + yGap

                return <Blade
                    x={innerLeft}
                    y={thisY}
                    width={innerWidth}
                    height={bladeHeight}
                    index={v[0]}
                    details={v[1]}
                    limits={props.capacityLimit}
                    palette={props.palette}/>
            })}
        </svg>
    )
}
