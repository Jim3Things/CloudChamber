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
    torLimit: number,
    pduLimit: number,
    connectorLimit: number,
    capacityLimit: BladeCapacity,
    rack: RackDetails,
    palette: Colors
}) {
    const slotHeight = 20
    const yGap = 1
    const headerGap = 10
    const headerSpace = (props.torLimit + props.pduLimit) * (slotHeight + yGap) + headerGap

    const innerLeft = 5
    const innerTop = 5

    const innerHeight = ((slotHeight + yGap) * props.bladeLimit) - yGap + headerSpace
    const innerWidth = 150

    const fullHeight = innerHeight + innerTop + innerTop
    const fullWidth = innerWidth + innerLeft + innerLeft

    let headerY = innerTop
    let bladeY = headerSpace + innerTop

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

            {Array.from(props.rack.tors).map((v) => {
                const thisY = headerY
                headerY += slotHeight + yGap

                return <Tor
                    x={innerLeft}
                    y={thisY}
                    width={innerWidth}
                    height={slotHeight}
                    details={v[1]}
                    palette={props.palette}
                    index={v[0]}
                />
            })}

            {Array.from(props.rack.pdus).map(v => {
                const thisY = headerY
                headerY += slotHeight + yGap

                return <PDU
                    x={innerLeft}
                    y={thisY}
                    width={innerWidth}
                    height={slotHeight}
                    details={v[1]}
                    palette={props.palette}
                    index={v[0]}
                />
            })}

            {Array.from(props.rack.blades).map((v) => {
                const thisY = bladeY
                bladeY += slotHeight + yGap

                return <Blade
                    x={innerLeft}
                    y={thisY}
                    width={innerWidth}
                    height={slotHeight}
                    index={v[0]}
                    details={v[1]}
                    limits={props.capacityLimit}
                    palette={props.palette}/>
            })}
        </svg>
    )
}
